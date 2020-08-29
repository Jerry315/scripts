package common

import (
	"encoding/json"
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"
)

var CycleMap = map[int64]int64{
	-1: -1,
	0:  0,
	1:  7,
	2:  30,
	3:  90,
	4:  15,
	5:  60,
	6:  180,
	7:  365,
	15: 99999,
}

var BaseDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))

var SecretKey = "tDz0pN0Lj3w2EH1IhhSJyYLl"

type Token struct {
	Token  string `json:"token"`
	Msg    string `json:"msg"`
	Status bool   `json:"status"`
}

type DBOption struct {
	Db     string   `yaml:"db"`
	Table  string   `yaml:"table"`
	Url    string   `yaml:"url"`
	Fields []string `yaml:"fields"`
}

type Config struct {
	Mongodb struct {
		Camera   DBOption `yaml:"camera"`
		Devices  DBOption `yaml:"devices"`
		MawarApp DBOption `yaml:"mawarapp"`
	}
	CheckServer string `yaml:"checkServer"`
	Limit       struct {
		Timeout          int   `yaml:"timeout"`
		MessageTimestamp int64 `yaml:"message_timestamp"`
		Step             int   `yaml:"step"`
	}
	Whitelist struct {
		All     []int64 `yaml:"all"`
		Picture []int64 `yaml:"picture"`
		Video   []int64 `yaml:"video"`
		Relay   []int64 `yaml:"relay"`
	} `yaml:"whitelist"`
	Log struct {
		Level       string `yaml:"level" json:"level" bson:"level"`
		Layout      string `yaml:"layout"`
		Path        string `yaml:"path" json:"path" bson:"path"`
		Filename    string `yaml:"filename" json:"filename" bson:"filename"`
		Format      string `yaml:"format" json:"format" bson:"format"`
		CycleFile   string `yaml:"cycleFile"`
		TimeOutFile string `yaml:"timeOutFile"`
		Expire      int64  `yaml:"expire"`
	}
	Relay struct {
		Urls     []string `yaml:"urls"`
		Username string   `yaml:"username"`
		Password string   `yaml:"password"`
	}
	Timeout   int64  `yaml:"timeout"`
	Project   string `yaml:"project"`
	Zname     string `yaml:"zname"`
	SecretId  string `yaml:"secretid"`
	SecretKey string `yaml:"secretkey"`
	Server    string `yaml:"server"`
}

type UserCredentials struct {
	SecretId  string `json:"secretid"`
	SecretKey string `json:"secretkey"`
}

type Timeoutcid struct {
	CID             int64 `json:"CID"`
	LatestImageTime int64 `json:"LatestImageTime"`
	LatestVideoTime int64 `json:"LatestVideoTime"`
}

type CheckServerResp0 struct {
	Timeoutcids []Timeoutcid `json:"timeoutcids"`
}

type CheckServerResp1 struct {
	Timeoutcids []struct {
		CID             int64 `json:"CID"`
		LatestVideoTime int64 `json:"LatestVideoTime"`
	} `json:"timeoutcids"`
}

type Streams struct {
	Info []*StreamInfo
}

type StreamInfo struct {
	Cid      string
	BwIn     string
	CostTime string
}

type Cid struct {
	CID int64 `bson:"_id"`
}

type UnCid struct {
	CID             int64
	MPIC            int64
	OPIC            int64
	MVideo          int64
	OVideo          int64
	SN              string
	Name            string
	Brand           string
	Group           string
	Model           string
	SoftwareVersion string
	SoftwareBuild   string
}

type DevicesDoc struct {
	CID             int64  `bson:"_id"`
	VideoStorage    string `bson:"storage"`
	PicStorage      string `bson:"pic_storage"`
	SN              string `bson:"sn"`
	Name            string `bson:"name"`
	Brand           string `bson:"brand"`
	Model           string `bson:"model"`
	SoftwareVersion string `bson:"software_version"`
	SoftwareBuild   string `bson:"software_build"`
}

type MawarAppDoc struct {
	CID   int64  `bson:"_id"`
	Group string `bson:"group"`
}

type MawarDoc struct {
	CID             int64
	VideoStorage    string
	PicStorage      string
	SN              string
	Name            string
	Brand           string
	Group           string
	Model           string
	SoftwareVersion string
	SoftwareBuild   string
}

type CameraDoc struct {
	CID              int64 `bson:"_id"`
	MessageTimestamp int64 `bson:"message_timestamp"`
	PushState        int64 `bson:"push_state"`
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

type CycleResponse struct {
	Status     bool    `json:"status"`
	Data       []UnCid `json:"data"`
	Msg        string  `json:"msg"`
	Project    string  `json:"project"`
	Zname      string  `json:"zname"`
	Module     string  `json:"module"`
	CreateTime int64   `json:"create_time"`
}

type TimeOutResponse struct {
	Status     bool         `json:"status"`
	Data       []Timeoutcid `json:"data"`
	Msg        string       `json:"msg"`
	Project    string       `json:"project"`
	Zname      string       `json:"zname"`
	Module     string       `json:"module"`
	CreateTime int64        `json:"create_time"`
	Total      int64        `json:"total"`
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
	confFile := path.Join(basePath, "converge_alert_mail_agent.yaml")
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
	// 自定义日志时间格式
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func InitLogger() *zap.Logger {
	conf := GetConf()
	logDir := BaseDir //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if conf.Log.Path != "" {
		logDir = path.Join(BaseDir, conf.Log.Path)
	}
	logFile := path.Join(logDir, conf.Log.Filename)
	//logFile := path.Join(baseDir, conf.Log.Filename)
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

func CompareCid(s1, s2, s3 []int64) (all, picture, video []int64) {
	// 白名单中的cid，去掉在全部白名单中的重复cid
	for _, c1 := range s1 {
		all = append(all, c1)
	}
	for _, c2 := range s2 {
		flag := true
		for _, c3 := range s3 {
			if c2 == c3 {
				all = append(all, c2)
				flag = false
				break
			}
		}
		if flag {
			picture = append(picture, c2)
		}
	}
	for _, c3 := range s3 {
		flag := true
		for _, c2 := range s2 {
			if c2 == c3 {
				flag = false
				break
			}
		}
		if flag {
			video = append(video, c3)
		}
	}
	return
}

func RecordCid(content interface{}, name string, logger *zap.Logger) {
	// 记录检测失败的cid内容到文件
	cids, _ := json.Marshal(content)
	err := ioutil.WriteFile(name, cids, 0644)
	if err != nil {
		logger.Error(fmt.Sprintf("cid %s 写入文件失败，%v", string(cids), err))
	}
}

func ClearExpireData(conf Config) {
	// 清除过期的日志目录下的文件
	logPath := path.Join(BaseDir, conf.Log.Path)
	nt := time.Now().Unix()
	et := nt - conf.Log.Expire*86400
	reg := regexp.MustCompile(`\d{8}`)
	files, err := ioutil.ReadDir(logPath)
	if err != nil {
		fmt.Printf("backup dir is not exist. %v\n", err)
		os.Exit(1)
	}
	os.Chdir(logPath)
	for _, fileName := range files {
		result := reg.FindAllString(fileName.Name(), -1)
		if len(result) == 0 {
			continue
		}
		ft, _ := time.Parse(conf.Log.Layout, result[0])
		if ft.Unix() < et {
			os.Remove(fileName.Name())
		}
	}
}
