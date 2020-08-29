package common

import (
	"encoding/json"
	"fmt"
	"github.com/natefinch/lumberjack"
	"github.com/prometheus/common/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

type Yaml struct {
	Codisktrackerurl string `yaml:"codisktracker_url"`
	Limit            int    `yaml:"limit"`
	Log              struct {
		Level    string `yaml:"level"`
		Path     string `yaml:"path"`
		Filename string `yaml:"filename"`
	}
}

type GroupInfos struct {
	Group_infos     []*DiscGroups
	Total_group_num int
	Total_disk_um   int
	Online_disk_num int
	Total_size      int
	Total_left      int
}

type DiscGroups struct {
	Group_id       int     `yaml:"group_id"`
	Create_time    int     `yaml:"create_time"`
	Group_type     int     `yaml:"group_type"`
	Onlines        int     `yaml:"onlines"`
	Ability        int     `yaml:"ability"`
	LockTime       int     `yaml:"lockTime"`
	Cycle          int     `yaml:"cycle"`
	Disc_infos     []*Disc `yaml:"disc_infos"`
	StorageScheme  []int   `yaml:"StorageScheme"`
	Dispatcher_ids []int   `yaml:"dispatcher_ids"`
}

type Disc struct {
	DiscID     int
	Public_ip  string
	Local_ip   string
	Port       int
	Cycle      int
	Is_online  int
	Used       int
	Left       int
	Begin_time int
	End_time   int
}

type ZeroGroup struct {
	Group_infos     []*ZeroInfo
	Online_disk_num int
	Total_size      int
	Total_group_num int
	Total_disk_num  int
	Total_left      int
}

type ZeroInfo struct {
	Ability        int
	Onlines        int
	Disc_infos     []*Disc
	Dispatcher_ids []int
	StorageScheme  []int
	Create_time    int
	LockTime       int
	Group_id       int
	Group_type     int
	Cycle          int
}

type Partition struct {
	LocalIP string
	GroupId int
	Port    int
	Cycle   int
	DiscID  int
}

type ZeroData struct {
	Data map[int]map[string]int
}

type ZeroQueryData struct {
	Data []map[string]int
}

type FirstTimeInfo struct {
	Disc_first_time_infos []DiscID
}

type DiscID struct {
	Disc_id    int `yaml:"disc_id"`
	First_time int `yaml:"first_time"`
}

type QueryData struct {
	Group []DiskInfo
	Cycle map[string]int
}

type DiskInfo struct {
	CreateTime   string
	GroupId      int
	GroupType    int
	Cycle        int
	Onlines      int
	LockTime     int
	DispatcherId int
	TotalUse     string
	TotaLeft     string
	Disks        []DiskDetail
}

type DiskDetail struct {
	FirstTime string
	DiscId    int
	IsOnline  int
	PublicIp  string
	LocalIp   string
	Port      int
	Left      string
}

type Disk struct {
}

func GetConf() Yaml {
	conf := new(Yaml)
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Error("The basePath failed: %s\n", err.Error())
	}
	conf_file := path.Join(basePath, "diskgroup_info.yaml")
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
	logFile := path.Join(conf.Log.Path, conf.Log.Filename)
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

func GetGroupData(url string, limit int) *GroupInfos {
	logger := InitLogger()
	groupInfos := new(GroupInfos)
	newUrl := url + "/console/group_info?disk_type=2&time=" + strconv.Itoa(int(time.Now().Unix())-limit)
	req, err := http.NewRequest("GET", newUrl, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("GetGroupData request data failed %#v", err))
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("GetGroupData parse request data failed %#v", err))
	}
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)

		logger.Info("GetGroupData request data success")
		err := json.Unmarshal(body, &groupInfos)
		if err != nil {
			logger.Error(fmt.Sprintf("GetGroupData parse request data to json failed %#v", err))
		}
	} else {
		logger.Error(fmt.Sprintf("GetGroupData something wrong with request %#v", err), zap.Int("statusCode", resp.StatusCode))
	}
	return groupInfos
}

func GetFirstTime(url string) FirstTimeInfo {
	logger := InitLogger()
	firstTimeInfo := FirstTimeInfo{}
	newUrl := url + "/console/get_disc_first_time"
	req, err := http.NewRequest("GET", newUrl, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("GetFirstTime request data failed %#v", err))
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("GetFirstTime parse request data failed %#v", err))
	}
	if resp.StatusCode == 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		logger.Info("GetFirstTime request data success")
		err := json.Unmarshal(body, &firstTimeInfo)
		if err != nil {
			logger.Error(fmt.Sprintf("GetFirstTime parse request data to json failed %#v", err))
		}
	} else {
		logger.Error(fmt.Sprintf("GetFirstTime something wrong with request %#v", err), zap.Int("statusCode", resp.StatusCode))
	}
	return firstTimeInfo
}
