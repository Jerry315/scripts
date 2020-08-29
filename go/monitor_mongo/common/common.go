package common

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"github.com/prometheus/common/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

type ServerStatus struct {
	Host           string `bson:"host"`
	Version        string `bson:"version"`
	Process        string `bson:"process"`
	Pid            int `bson:"pid"`
	Uptime         int `bson:"uptime"`
	LocalTime      time.Time `bson:"localTime"`
	Asserts        Asserts `bson:"asserts"`
	Connections    Connections `bson:"connections"`
	Extra_info     ExtraInfo `bson:"extra_info"`
	GlobalLock     GlobalLock `bson:"globalLock"`
	Network        Network `bson:"network"`
	Opcounters     Opcounters `bson:"opcounters"`
	OpcountersRepl OpcountersRepl `bson:"opcountersRepl"`
	Mem            Memory `bson:"mem"`
}

type Asserts struct {
	Regular   int `bson:"regular"`
	Warning   int `bson:"warning"`
	Msg       int `bson:"msg"`
	User      int `bson:"user"`
	Rollovers int `bson:"rollovers"`
}

type Connections struct {
	Current      int `bson:"current"`
	Available    int `bson:"available"`
	TotalCreated int `bson:"totalCreated"`
}

type ExtraInfo struct {
	Note             string `bson:"note"`
	Heap_usage_bytes int `bson:"heap_usage_bytes"`
	Page_faults      int `bson:"page_faults"`
}

type Memory struct {
	Bits              int `bson:"bits"`
	Resident          int `bson:"resident"`
	Virtual           int `bson:"virtual"`
	Supported         bool `bson:"supported"`
	Mapped            int `bson:"mapped"`
	MappedWithJournal int `bson:"mappedWithJournal"`
}

type Network struct {
	BytesIn     int64 `bson:"bytesIn"`
	BytesOut    int64 `bson:"bytesOut"`
	NumRequests int64 `bson:"numRequests"`
}

type GlobalLock struct {
	TotalTime     int64 `bson:"totalTime"`
	CurrentQueue  CurrentQueue `bson:"currentQueue"`
	ActiveClients ActiveClients `bson:"activeClients"`
}

type ActiveClients struct {
	Total   int `bson:"total"`
	Readers int `bson:"readers"`
	Writers int `bson:"writers"`
}

type CurrentQueue struct {
	Total   int `bson:"total"`
	Readers int `bson:"readers"`
	Writers int `bson:"writers"`
}

type Opcounters struct {
	Insert  int `bson:"insert"`
	Query   int `bson:"query"`
	Update  int `bson:"update"`
	Delete  int `bson:"delete"`
	Getmore int `bson:"getmore"`
	Command int `bson:"command"`
}

type OpcountersRepl struct {
	Insert  int `bson:"insert"`
	Query   int `bson:"query"`
	Update  int `bson:"update"`
	Delete  int `bson:"delete"`
	Getmore int `bson:"getmore"`
	Command int `bson:"command"`
}

type ReplSetGetStatus struct {
	Members []Member `bson:"members"`
}

type Member struct {
	Name       string `bson:"name"`
	Health     int `bson:"health"`
	StateStr   string `bson:"statStr"`
	OptimeDate time.Time `bson:"optimeDate"`
}

type Config struct {
	Instance []struct {
		Address string `yaml:"address"`
		Db      string `yaml:"db"`
		Port    int    `yaml:"port"`
	}
	Log struct {
		FileName string `yaml:"fileName"`
		Level    string `yaml:"level"`
		Path     string `yaml:"path"`
	}
}

func GetConf() Config {
	conf := new(Config)
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Error("The basePath failed: %s\n", err)
	}
	confFile := path.Join(basePath, "monitor_mongo.yaml")
	confData, err := ioutil.ReadFile(confFile)
	if err != nil {
		log.Error("confFile Get err #%v", err)
	}
	err = yaml.Unmarshal(confData, conf)
	if err != nil {
		log.Error("Unmarshal: #%v", err)
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

func GetServerStatus(url string, logger *zap.Logger) (ss ServerStatus, err error) {
	session, err := mgo.Dial(url)
	if err != nil {
		logger.Error(fmt.Sprintf("create session failed. %v", err))
		return
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	err = session.Run(bson.D{{"serverStatus", 1}}, &ss)
	if err != nil {
		logger.Error(fmt.Sprintf("execute command failed. %v", err))
		return
	}
	return
}

func GetReplStatus(url string, logger *zap.Logger) (rs ReplSetGetStatus, err error) {
	session, err := mgo.Dial(url)
	if err != nil {
		logger.Error(fmt.Sprintf("create session failed. %v", err))
		return
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	err = session.Run(bson.D{{"replSetGetStatus", 1}}, &rs)
	if err != nil {
		logger.Error(fmt.Sprintf("execute command failed. %v", err))
		return
	}
	return
}
