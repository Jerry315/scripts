package common

import (
	"fmt"
	"github.com/jordan-wright/email"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/smtp"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Config struct {
	Bind     string   `yaml:"bind"`
	From     string   `yaml:"from"`
	Receiver []string `yaml:"receiver"`
	Subject  string   `yaml:"subject"`
	Smtp     struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Server   string `yaml:"server"`
	}
	Log struct {
		Level    string `yaml:"level" json:"level" bson:"level"`
		Path     string `yaml:"path" json:"path" bson:"path"`
		Filename string `yaml:"filename" json:"filename" bson:"filename"`
		Format   string `yaml:"format" json:"format" bson:"format"`
	}
}

func GetConf() Config {
	//获取配置文件
	conf := new(Config)
	basePath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	confFile := path.Join(basePath, "mail_api.yaml")
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

func SendMail(conf Config, body string) error {
	//发送纯文本内容的邮件
	//定义邮箱服务器连接信息
	e := email.NewEmail()
	e.From = conf.From
	e.To = conf.Receiver
	e.Subject = conf.Subject + " " + time.Now().Format("2006-01-02")
	e.Text = []byte(body)
	return e.Send(conf.Smtp.Server+":25", smtp.PlainAuth("", conf.Smtp.Username, conf.Smtp.Password, conf.Smtp.Server))
}
