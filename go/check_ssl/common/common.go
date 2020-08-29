package common

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

type Config struct {
	Expire int `yaml:"expire"`
	LvsUrl string `yaml:"lvsUrl"`
	Domain []struct {
		Url   string   `yaml:"url"`
		Hosts []string `yaml:"hosts"`
	}
	Log struct {
		Level    string `yaml:"level" json:"level" bson:"level"`
		Path     string `yaml:"path" json:"path" bson:"path"`
		Filename string `yaml:"filename" json:"filename" bson:"filename"`
	}
}

func GetConf() Config {
	conf := new(Config)
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("The eacePath failed: %s\n", err.Error())
	}
	confFile := path.Join(basePath, "check_ssl.yaml")
	confData, err := ioutil.ReadFile(confFile)
	if err != nil {
		fmt.Printf("confFile Get err #%v", err)
	}
	err = yaml.Unmarshal(confData, conf)
	if err != nil {
		fmt.Printf("Unmarshal: #%v", err)
	}
	return *conf
}

func InitLogger() *zap.Logger {
	conf := GetConf()
	basePath := conf.Log.Path
	if basePath == "" {
		basePath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	}
	logFile := path.Join(basePath, conf.Log.Filename)
	hook := lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    128,
		MaxBackups: 7,
		MaxAge:     7,
		Compress:   true,
	}
	w := zapcore.AddSync(&hook)

	var level zapcore.Level
	switch conf.Log.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		level,
	)
	logger := zap.New(core)
	return logger
}
