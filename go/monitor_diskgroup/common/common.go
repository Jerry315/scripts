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
	MonitorUrl   string `yaml:"mointor_url"`
	GroupInfoUrl string `yaml:"group_info_url"`
	Log          struct {
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
	Group_id      int     `yaml:"group_id"`
	Create_time   int     `yaml:"create_time"`
	Group_type    int     `yaml:"group_type"`
	Onlines       int     `yaml:"onlines"`
	Ability       int     `yaml:"ability"`
	LockTime      int     `yaml:"lockTime"`
	Cycle         int     `yaml:"cycle"`
	Disc_infos    []*Disc `yaml:"disc_infos"`
	StorageScheme []int   `yaml:"StorageScheme"`
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

type Disk struct {
	DiscID      int
	Ability     int
	LocalIP     string
	Port        int
	Upload_rate int
	Used        int
	Size        int
}

type DiskGroup struct {
	Group_id        int
	Group_type      int
	Cycle           int
	Size            int
	Max_upload_rate int
	Upload_rate     int
	Disk_infos      []*Disk
	Onlines         int
	DispatcherIDs   []int
}

type MonitorInfo struct {
	Monitor_info []*DiskGroup
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

type ZeroData struct {
	Data map[int]map[string]int
}

type ZeroQueryData struct {
	Data []map[string]int
}

func GetConf() Yaml {
	conf := new(Yaml)
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Error("The basePath failed: %s\n", err.Error())
	}
	conf_file := path.Join(basePath, "monitor_diskgroup.yaml")
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
	basePath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	logFile := path.Join(basePath, conf.Log.Path, conf.Log.Filename)
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

func GetGroupData(url string) GroupInfos {
	logger := InitLogger()
	groupInfos := new(GroupInfos)
	req, err := http.NewRequest("GET", url, nil)
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
	return *groupInfos
}

func GetMonitorData(url string) MonitorInfo {
	logger := InitLogger()
	monitorInfo := new(MonitorInfo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("GetMonitorData request data failed %#v", err))
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("GetMonitorData parse request data failed %#v", err))
	}

	if resp.StatusCode == 200 {
		logger.Info("GetMonitorData request data success")
		body, _ := ioutil.ReadAll(resp.Body)
		err := json.Unmarshal(body, &monitorInfo)
		if err != nil {
			logger.Error(fmt.Sprintf("GetMonitorData parse request data to json failed %#v", err))
		}
	} else {
		logger.Error(fmt.Sprintf("GetMonitorData something wrong with request %#v", err), zap.Int("statusCode", resp.StatusCode))
	}
	return *monitorInfo
}

func GetZeroData(url string) ZeroGroup {
	logger := InitLogger()
	zeroGroup := new(ZeroGroup)
	currentTime := time.Now().Unix() - 86400
	url = url + "&time=" + strconv.Itoa(int(currentTime)) + "&group_id=0"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("GetZeroData request data failed %#v", err))
	}
	req.Header.Add("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode == 200 {
		logger.Info("GetZeroData request data success")
		body, _ := ioutil.ReadAll(resp.Body)
		err := json.Unmarshal(body, &zeroGroup)
		if err != nil {
			logger.Error(fmt.Sprintf("GetZeroData parse request data to json failed %#v", err))
		}
	} else {
		logger.Error(fmt.Sprintf("GetZeroData something wrong with request %#v", err), zap.Int("statusCode", resp.StatusCode))
	}
	return *zeroGroup
}
