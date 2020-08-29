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
	Url    string   `yaml:"url"`
	Db     string   `yaml:"db"`
	Fields []string `yaml:"fields"`
	Log    struct {
		Level    string `yaml:"level"`
		Path     string `yaml:"path"`
		Filename string `yaml:"filename"`
	}
}

type Doc struct {
	ID             int64  `bson:"_id"`
	Group_id       string `bson:"group_id"`
	Sn             int    `bson:"sn"`
	Status         int    `bson:"status"`
	App_id         string `bson:"app_id"`
	Name           string `bson:"name"`
	Uid            int64  `bson:"uid"`
	Osd            string `bson:"osd"`
	Is_bind        bool   `bson:"is_bind"`
	Type           string `bson:"type"`
	Video          string `bson:"video"`
	Brand          string `bson:"brand"`
	Alias_brand    string `bson:"alias_brand"`
	Bitrate        int    `bson:"bitrate"`
	Bitlevel       int    `bson:"bitlevel"`
	Model          string `bson:"model"`
	Alias_model    string `bson:"alias_model"`
	Storage        string `bson:"storage"`
	Pic_storage    string `bson:"pic_storage"`
	Image_invert   int    `bson:"image_invert"`
	Alarm          int    `bson:"alarm"`
	Alarm_interval int    `bson:"alarm_interval"`
	Alarm_count    int    `bson:"alarm_count"`
	Alarm_zone     string `bson:"alarm_zone"`
	Validate_code  string `bson:"validate_code"`
	With_platform  int    `bson:"with_platform"`
	Video_quality  int    `bson:"video_quality"`
	Project        string `bson:"project"`
	Group          string `bson:"group"`
	Geo            struct {
		Address   string `bson:"address"`
		Name      string `bson:"name"`
		Latitude  string `bson:"latitude"`
		Longitude string `bson:"longitude"`
	} `bson:"geo"`
	Status_light                 int       `bson:"status_light"`
	Nightmode                    int       `bson:"nightmode"`
	Contact                      string    `bson:"contact"`
	Signature                    string    `bson:"signature"`
	Resolution                   string    `bson:"resolution"`
	Place                        int       `bson:"place"`
	Adcode                       string    `bson:"adcode"`
	Detect_type                  int       `bson:"detect_type"`
	Min_face_size_width          int       `bson:"min_face_size_width"`
	Min_face_size_heigh          int       `bson:"min_face_size_heigh"`
	Min_face_size_type           int       `bson:"min_face_size_type"`
	Face_detect_confidence_level int       `bson:"face_detect_confidence_level"`
	Ys_upload_face_coordinate    int       `bson:"ys_upload_face_coordinate"`
	Framerate                    int       `bson:"framerate"`
	Exposemode                   int       `bson:"exposemode"`
	Ai_face_ori                  int       `bson:"ai_face_ori"`
	Ai_type                      int       `bson:"ai_type"`
	Ai_face_frame                int       `bson:"ai_face_frame"`
	Ai_face_position             int       `bson:"ai_face_position"`
	Ai_face_pps                  int       `bson:"ai_face_pps"`
	Pic_server_address           string    `bson:"pic_server_address"`
	Created                      time.Time `bson:"created"`
	Updated                      time.Time `bson:"updated"`
	Software_version             string    `bson:"software_version"`
	Software_build               string    `bson:"software_build"`
	Firmware                     string    `bson:"firmware"`
	Debug_model                  int       `bson:"debug_model"`
	Configuration_file           string    `bson:"configuration_file"`
	Last_login                   time.Time `bson:"last_login"`
	Last_register                time.Time `bson:"last_register"`
}

type Result struct {
	Query []*Doc
}

func GetConf() Config {
	conf := new(Config)
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("The basePath failed: %s\n", err)
	}
	confFile := path.Join(basePath, "mongo_to_csv.yaml")
	confData, err := ioutil.ReadFile(confFile)
	if err != nil {
		fmt.Printf("confFile Get err #%v", err)
	}
	err = yaml.Unmarshal(confData, conf)
	if err != nil {
		fmt.Printf("Unmarshal: #%v", err)
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
