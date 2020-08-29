package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/template"
	"github.com/jordan-wright/email"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math"
	"math/rand"
	"net/smtp"
	"os"
	"path"
	"path/filepath"
	"time"
)

type DBOptions struct {
	Db     string   `yaml:"db"`
	Table  string   `yaml:"table"`
	Url    string   `yaml:"url"`
	Fields []string `yaml:"fields"`
}

type Config struct {
	Mongodb struct {
		Camera   DBOptions `yaml:"camera"`
		Devices  DBOptions `yaml:devices`
		MawarApp DBOptions `yaml:"mawarapp"`
	}
	Url struct {
		CheckServer string `yaml:"checkServer"`
		Relay       struct {
			Url      string `yaml:"url"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		}
	}
	Limit struct {
		Cid_num           int   `yaml:"cid_num"`
		Timeout           int   `yaml:"timeout"`
		Message_timestamp int64 `yaml:"message_timestamp"`
		Step              int   `yaml:"step"`
	}
	Whitelist struct {
		All     []int64 `yaml:"all"`
		Picture []int64 `yaml:"picture"`
		Video   []int64 `yaml:"video"`
	} `yaml:"whitelist"`
	Mail struct {
		Enable bool `yaml:"enable"`
		Send   struct {
			From     string `yaml:"from"`
			Subject  string `yaml:"subject"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
			Server   string `yaml:"server"`
		}
		Recive struct {
			Normal []string `yaml:"normal"`
			Admin  []string `yaml:"admin"`
		}
	}
	Log struct {
		Level    string `yaml:"level" json:"level" bson:"level"`
		Path     string `yaml:"path" json:"path" bson:"path"`
		Filename string `yaml:"filename" json:"filename" bson:"filename"`
		Format   string `yaml:"format" json:"format" bson:"format"`
		Exfile   string `yaml:"exfile"`
	}
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
	Group           string
	Brand           string
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

func GetConf() Config {
	//获取配置文件
	conf := new(Config)
	basePath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	confFile := path.Join(basePath, "check_storage_cycle.yaml")
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

func FTI(x int, y float64) int {
	//四舍五入，返回整数
	w, f := math.Modf(float64(x) * y)
	if f >= 0.5 {
		w++
	}
	return int(w)
}

func RandSlice(x, y int) (r []int) {
	//在0-x之间生成y个不同的整数，并返回整数列表
	c := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < y; i++ {
		r = append(r, c.Intn(x))
	}
	return r
}

func CheckHistory(name string) (cids map[string][]int64) {
	//检测上一次执行时异常存储的cid文件，如果存在返回上次检测的cid列表
	if fileObj, err := os.Open(name); err == nil {
		defer fileObj.Close()
		if contents, err := ioutil.ReadAll(fileObj); err == nil {
			err := json.Unmarshal(contents, &cids)
			if err != nil {
				return
			}
		}
	}
	return
}

func RecordCid(uncid map[string][]int64, name string, logger *zap.Logger) {
	cids, _ := json.Marshal(uncid)
	err := ioutil.WriteFile(name, cids, 0644)
	if err != nil {
		logger.Error(fmt.Sprintf("cid %s 写入文件失败，%v", string(cids), err))
	}
}

func CompareCid(s1, s2, s3 []int64) (all, picture, video []int64) {
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

func SendReport(conf Config, uncids []UnCid, name string, logger *zap.Logger) error {
	if !conf.Mail.Enable{
		return nil
	}
	//以html形式发送邮件
	basePath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	//定义邮箱服务器连接信息
	e := email.NewEmail()
	_, err := e.AttachFile(name)
	if err != nil {
		logger.Error(fmt.Sprintf("发送附件失败。%v", err))
	}
	e.From = conf.Mail.Send.From
	e.To = conf.Mail.Recive.Normal
	e.Subject = conf.Mail.Send.Subject + " " + time.Now().Format("2006-01-02")
	t, err := template.ParseFiles(path.Join(basePath, "email-template.html"))
	if err != nil {
		return err
	}
	//buffer是一个实现了读写方法的可变大小的字节缓冲
	bufferBody := new(bytes.Buffer)
	_ = t.Execute(bufferBody, struct {
		UnCids []UnCid
	}{uncids})
	//html形式的消息
	e.HTML = bufferBody.Bytes()
	return e.Send(conf.Mail.Send.Server+":25", smtp.PlainAuth("", conf.Mail.Send.Username, conf.Mail.Send.Password, conf.Mail.Send.Server))
}

func SendMail(conf Config, body string) error {
	//发送纯文本内容的邮件
	//定义邮箱服务器连接信息
	if !conf.Mail.Enable{
		return nil
	}
	e := email.NewEmail()
	e.From = conf.Mail.Send.From
	e.To = conf.Mail.Recive.Admin
	e.Subject = conf.Mail.Send.Subject + " " + time.Now().Format("2006-01-02")
	e.Text = []byte(body)
	return e.Send(conf.Mail.Send.Server+":25", smtp.PlainAuth("", conf.Mail.Send.Username, conf.Mail.Send.Password, conf.Mail.Send.Server))
}
