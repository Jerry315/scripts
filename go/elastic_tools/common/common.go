package common

import (
	"github.com/natefinch/lumberjack"
	"github.com/prometheus/common/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Config struct {
	Timeout                int    `yaml:"timeout"`
	EsUrl                  string `yaml:"esUrl"`
	MaxSnapshotBytesPerSec string `yaml:"maxSnapshotBytesPerSec"`
	MaxRestoreBytesPerSec  string `yaml:"maxRestoreBytesPerSec"`
	Snapshot               []struct {
		Index     []string `yaml:"index"`
		Enable    bool     `yaml:"enable"`
		DelayDays int      `yaml:"delayDays"`
		DateFmt   string   `yaml:"dateFmt"`
	} `yaml:"snapshot"`
	Delete []struct {
		Index     []string `yaml:"index"`
		Enable    bool     `yaml:"enable"`
		DelayDays int      `yaml:"delayDays"`
		DateFmt   string   `yaml:"dateFmt"`
	} `yaml:"delete"`
	Settings []struct {
		Index     []string `yaml:"index"`
		Enable    bool     `yaml:"enable"`
		DelayDays int      `yaml:"delayDays"`
		DateFmt   string   `yaml:"dateFmt"`
		Tag       string   `yaml:"tag"`
	} `yaml:"settings"`
	Log struct {
		Level    string `yaml:"level"`
		Path     string `yaml:"path"`
		Filename string `yaml:"filename"`
	}
}

func GetConf() Config {
	conf := new(Config)
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Error("The basePath failed: %s\n", err)
	}
	confFile := path.Join(basePath, "elastic_tools.yaml")
	confData, err := ioutil.ReadFile(confFile)
	if err != nil {
		log.Error("confFile Get err %v", err)
	}
	err = yaml.Unmarshal(confData, conf)
	if err != nil {
		log.Error("Unmarshal: %v", err)
	}
	return *conf
}

func sysLogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
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
	encoderConfig.EncodeTime = sysLogTimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		level,
	)
	logger := zap.New(core)
	return logger
}
