package common

import (
	"github.com/natefinch/lumberjack"
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Yaml struct {
	Httpbasicauth struct{
		Stat struct{
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		}
		Api struct{
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		}
	}
	Urls struct{
		Relay string `yaml:"relay"`
		Api string `yaml:"api"`
		Oss string `yaml:"oss"`
	}
	Log struct{
		Level    string `yaml:"level" json:"level" bson:"level"`
		Path     string `yaml:"path" json:"path" bson:"path"`
		Filename string `yaml:"filename" json:"filename" bson:"filename"`
		Format   string `yaml:"format" json:"format" bson:"format"`
	}
	Ignorecid []int `yaml:"ignorecid"`

}

type Tokens struct {
	Cid   int
	Token string
}

type TokenInfo struct {
	Request_id string
	Tokens     []*Tokens
}

type TimeLines struct {
	Timelines []*PeriodTime
}

type PeriodTime struct {
	Begin int
	End int
}

type Streams struct {
	Info []*StreamInfo
}

type StreamInfo struct {
	Cid string
	BwIn string
	CostTime string
}

func GetConf() Yaml {
	conf := new(Yaml)
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Error("The basePath failed: %s\n", err.Error())
	}
	conf_file := path.Join(basePath, "health_check.yaml")
	confFile, err := ioutil.ReadFile(conf_file)
	if err != nil {
		log.Error("confFile Get err #%v", err)
	}
	err = yaml.Unmarshal(confFile, conf)
	if err != nil {
		log.Error("Unmarshal: #%v", err)
	}
	return *conf
}

func sysLogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func InitLogger() *zap.Logger {
	conf := GetConf()
	baseDir := conf.Log.Path
	if baseDir == "" {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
		if err != nil {
			log.Fatal(err)
		} else {
			baseDir = dir
		}
	}
	logFile := path.Join(baseDir, conf.Log.Filename)
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
	encoderConfig.EncodeTime = sysLogTimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		level,
	)
	logger := zap.New(core)
	return logger
}
