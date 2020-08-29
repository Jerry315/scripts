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
)

type Yaml struct {
	AppId      string `yaml:"appId"`
	AppKey     string `yaml:"appKey"`
	ApiUrl     string `yaml:"apiUrl"`
	OssUrl     string `yaml:"ossUrl"`
	Cid        int    `yaml:"cid"`
	Expiretype int    `yaml:"expiretype"`
	Area       string `yaml:"area"`
	Upload     struct {
		Path   string `yaml:"path"`
		Layout string `yaml:"layout"`
	}
	Download struct {
		Path   string   `yaml:"path"`
		ObjIds []string `yaml:"objIds"`
	}
	DataName string `yaml:"dataName"`
	Log      struct {
		Level    string `yaml:"level"`
		Path     string `yaml:"path"`
		Filename string `yaml:"filename"`
	}
}

type TokenInfo struct {
	Request_id string
	Tokens     []Tokens
}

type Tokens struct {
	Cid   int
	Token string
}

type Response struct {
	Obj_id      string
	Name        string
	C_id        int
	File_size   int
	Expire_type int
	Area_id     int
}

type StoreData struct {
	Upload   []*FileInfo
	Download []*FileInfo
}

type FileInfo struct {
	Area     string
	FileSize int
	Md5Str   string
	FileName string
	ObjId    string
}

func GetConf() Yaml {
	conf := new(Yaml)
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Error("The basePath failed: %s\n", err.Error())
	}
	conf_file := path.Join(basePath, "backup_oss.yaml")
	confFile, err := ioutil.ReadFile(conf_file)
	if err != nil {
		log.Error("confFile Get err %#v", err)
	}
	err = yaml.Unmarshal(confFile, conf)
	if err != nil {
		log.Error("Unmarshal: %#v", err)
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
