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
	ApiUrl     string `yaml:"apiUrl"`
	AppId      string `yaml:"appId"`
	AppKey     string `yaml:"appKey"`
	OssUrl     string `yaml:"ossUrl"`
	Sn         string `yaml:"sn"`
	Cid        int    `yaml:"cid"`
	ExpireType int    `yaml:"expireType"`
	Log        struct {
		Level    string `yaml:"level"`
		Path     string `yaml:"path"`
		FileName string `yaml:"fileName"`
	}
}

type TokenInfo struct {
	RequestId string
	Tokens    []Tokens
}

type Tokens struct {
	Cid   int
	Token string
}

type ResponseUpload2 struct {
	Obj_id    string `yaml:"obj_id"`
	Name      string `yaml:"name"`
	Key       string `yaml:"key"`
	File_size int    `yaml:"file_size"`
}

type ResponseUpload3 struct {
	Attachments []struct {
		Key       string `yaml:"key"`
		File_name string `yaml:"file_name"`
	}
}

type ResponseMawar struct {
	Base_url string
	Obj_infos []struct{
		Obj_id string
		Upload_time int
	}
}

type CustomError struct {
	s string
}

func GetConf() Config {
	conf := new(Config)
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("The basePath failed: %s\n", err.Error())
	}
	confFile := path.Join(basePath, "chain_check.yaml")
	confData, err := ioutil.ReadFile(confFile)
	if err != nil {
		fmt.Printf("confFile Get err %#v", err)
	}
	err = yaml.Unmarshal(confData, conf)
	if err != nil {
		fmt.Printf("Unmarshal: %#v", err)
	}
	return *conf
}

func InitLogger(logFile, logLevel string) *zap.Logger {
	hook := lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    128,
		MaxBackups: 7,
		MaxAge:     7,
		Compress:   true,
	}
	w := zapcore.AddSync(&hook)

	var level zapcore.Level
	switch logLevel {
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

func (e CustomError) Error() string {
	return e.s
}