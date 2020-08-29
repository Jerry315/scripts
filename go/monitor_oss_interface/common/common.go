package common

import (
	"encoding/json"
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"lytoken"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const Version = "0.0.5"

var CycleMap = map[int]int{
	0:  0,
	1:  7,
	2:  30,
	3:  90,
	4:  15,
	5:  60,
	7:  365,
	15: 99999,
}

//type Duration int64

type Config struct {
	ApiUrl  string        `yaml:"apiUrl"`
	AppId   string        `yaml:"appId"`
	AppKey  string        `yaml:"appKey"`
	OssUrl  string        `yaml:"ossUrl"`
	RtmpUrl string        `yaml:"rtmpUrl"`
	Timeout time.Duration `yaml:"timeout"`
	Retry   int           `yaml:"retry"`
	Video   []struct {
		Cycle int `yaml:"cycle"`
		Cid   int `yaml:"cid"`
	} `yaml:"video"`
	Picture []struct {
		Cycle int `yaml:"cycle"`
		Cid   int `yaml:"cid"`
	} `yaml:"picture"`
	Log struct {
		Level    string `yaml:"level"`
		Path     string `yaml:"path"`
		FileName string `yaml:"fileName"`
	}
}

type Record struct {
	Cid       int
	Cycle     int
	ObjectId  string
	Size      int
	Timestamp int
	Media     int
}

type Response struct {
	Obj_id    string `yaml:"obj_id"`
	Name      string `yaml:"name"`
	File_size int    `yaml:"file_size"`
}

type RecordTs struct {
	TimeList [] struct {
		BaseIndex int    `json:"base_index"`
		Oid       string `json:"oid"`
		Begin     int    `json:"begin"`
		End       int    `json:"end"`
	} `json:"time_list"`
}

func GetConf() Config {
	conf := new(Config)
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("The basePath failed: %s\n", err.Error())
	}
	confFile := path.Join(basePath, "monitor_oss_interface.yaml")
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

func WriteFile(records []Record, name string, logger *zap.Logger) {
	tmp, _ := json.Marshal(records)
	err := ioutil.WriteFile(name, tmp, 0644)
	if err != nil {
		logger.Error(fmt.Sprintf("writeFile [WriteFile]记录 %s 写入文件失败，%v", string(tmp), err))
	}
}

type CustomError struct {
	S string
}

func (e CustomError) Error() string {
	return e.S
}

func GetToken(cid, cycle int, key string) (token string) {
	binStr := strconv.FormatInt(int64(cycle), 2)
	length := len(binStr)
	if length < 4 {
		binStr = fmt.Sprintf("%v%v", strings.Repeat("0", 4-length), binStr)
	}
	control := lytoken.NewControl()
	control.SetOption(lytoken.OptionStorage, binStr)
	ss := lytoken.New(uint32(cid), control.Number(), time.Hour)
	token, _ = ss.Str([]byte(key))
	return

}

func ListFile(dir string) (files []string) {
	tmp, _ := ioutil.ReadDir(dir)
	for _, file := range tmp {
		if !file.IsDir() {
			files = append(files, file.Name())
		}
	}
	return
}

func CheckHistory(name string) (records []Record) {
	//获取历史记录列表
	if fileObj, err := os.Open(name); err == nil {
		defer fileObj.Close()
		if contents, err := ioutil.ReadAll(fileObj); err == nil {
			err := json.Unmarshal(contents, &records)
			if err != nil {
				return
			}
		}
	}
	return
}

func CleanExpireRecord(rs []Record, currentTime int64) (records []Record) {
	for _, record := range rs {
		if record.Timestamp+CycleMap[record.Cycle]*86400 > int(currentTime) {
			records = append(records, record)
		}
	}
	return
}

func RemoveRepeatedElement(cids []int) (newCids []int) {
	sort.Ints(cids)
	for i := 0; i < len(cids); i++ {
		repeat := false
		for j := i + 1; j < len(cids); j++ {
			if cids[i] == cids[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newCids = append(newCids, cids[i])
		}
	}
	return
}

func ParseTimeStamp(timestamp int) string {
	t := time.Unix(int64(timestamp),0)
	ts := t.Format("2006-01-02 15:04:05")
	return ts
}

func GetVideoCids(conf Config) (cids []int) {
	for _, item := range conf.Video {
		cids = append(cids, item.Cid)
	}
	cids = RemoveRepeatedElement(cids)
	return
}
