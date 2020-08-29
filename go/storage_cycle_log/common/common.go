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
	"time"
)

type Config struct {
	Mongodb struct {
		Mawar struct {
			Url   string `yaml:"url"`
			Db    string `yaml:"db"`
			Table string `yaml:"table"`

			Fields []string `yaml:"fields"`
		}
	}
	Url struct {
		CheckServer string `yaml:"checkServer"`
	}
	Step int `yaml:"step"`
	Log  struct {
		Level    string `yaml:"level"`
		Path     string `yaml:"path"`
		Filename string `yaml:"filename"`
		Format   string `yaml:"format"`
		Cycle    int    `yaml:"cycle"`
	}
	MonitorFile string `yaml:"monitorFile"`
}

type CidInfo struct {
	Media string
	Cycle int64
	Count int64
	Time  int64
}

type MawarDoc struct {
	CID int64 `bson:"_id"`
}

type Response struct {
	Cids []struct {
		CID   int64
		Cycle int64
		Time  int64
	}
}

type Error struct {
	ErrCode int
	ErrMsg  string
}

func NewError(code int, msg string) *Error {
	return &Error{ErrCode: code, ErrMsg: msg}
}

func (err *Error) Error() string {
	return err.ErrMsg
}

func GetConf() Config {
	//获取配置文件
	conf := new(Config)
	basePath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	confFile := path.Join(basePath, "storage_cycle_log.yaml")
	data, err := ioutil.ReadFile(confFile)
	if err != nil {
		fmt.Printf("confFile Get err %v", err)
	}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}
	return *conf
}

func sysLogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func InitLogger() *zap.Logger {
	conf := GetConf()
	logDir, _ := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if conf.Log.Path != "" {
		logDir = path.Join(logDir, conf.Log.Path)
	}
	logFile := path.Join(logDir, conf.Log.Filename)
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

func TraceFile(content, name string) {
	fd, _ := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	buf := []byte(content)
	fd.Write(buf)
	fd.Close()
}

func CleanFile(dir string, cycle int) {
	nt := time.Now().Unix()
	et := int(nt) - cycle*86400
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		if int(f.ModTime().Unix()) <= et {
			rFile := filepath.Join(dir, f.Name())
			os.Remove(rFile)
		}
	}

}
