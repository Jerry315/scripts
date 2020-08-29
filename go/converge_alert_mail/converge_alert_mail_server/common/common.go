package common

import (
	"bytes"
	"fmt"
	"github.com/jordan-wright/email"
	"github.com/natefinch/lumberjack"
	"github.com/tealeg/xlsx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"net/smtp"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

var BaseDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))

var SecretKey = "tDz0pN0Lj3w2EH1IhhSJyYLl"

type Token struct {
	Token  string `json:"token"`
	Msg    string `json:"msg"`
	Status bool   `json:"status"`
}

type ModuleInfo struct {
	Name     string `yaml:"name"`
	Template string `yaml:"template`
	Subject  string `yaml:"subject"`
}

type Config struct {
	Bind    string `yaml:"bind"`
	Port    string `yaml:"port"`
	Mongodb struct {
		Db    string `yaml:"db"`
		Table string `yaml:"table"`
		Url   string `yaml:"url"`
	}
	Log struct {
		Expire int64  `yaml:"expire"`
		Exfile string `yaml:"exfile"`
		Level  string `yaml:"level"`
		Layout string `yaml:"layout"`
		Path   string `yaml:"path"`
		Name   string `yaml:"name"`
		Format string `yaml:"format"`
	}
	SecretId      string     `yaml:"secretid"`
	SecretKey     string     `yaml:"secretkey"`
	DeviceCycle   ModuleInfo `yaml:"deviceCycle"`
	DeviceTimeOut ModuleInfo `yaml:"deviceTimeOut"`
	Mail          struct {
		Send struct {
			From     string `yaml:"from"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
			Server   string `yaml:"server"`
		}
		Recive struct {
			Normal []string `yaml:"normal"`
			Admin  []string `yaml:"admin"`
		}
	}
}

type UserCredentials struct {
	SecretId  string `json:"secretid"`
	SecretKey string `json:"secretkey"`
}

type TimeoutCid struct {
	CID             int64 `json:"CID"`
	LatestImageTime int64 `json:"LatestImageTime"`
	LatestVideoTime int64 `json:"LatestVideoTime"`
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

type DeviceCycleDoc struct {
	Data       []UnCid `bson:"data" json:"data"`
	Project    string  `bson:"project" json:"project"`
	Zname      string  `bson:"zname" json:"zname"`
	Module     string  `bson:"module" json:"module"`
	CreateTime int64   `bson:"create_time" json:"create_time"`
	Status     bool    `bson:"status" json:"status"`
	Msg        string  `bson:"msg" json:"msg"`
}

type DeviceCycleReportData struct {
	Data     []UnCid `json:"data"`
	Zname    string  `json:"zname"`
	PicNum   int     `json:"pic_num"`
	VideoNum int     `json:"video_num"`
}

type DeviceTimeOutDoc struct {
	Data       []TimeoutCid `bson:"data" json:"data"`
	Project    string       `bson:"project" json:"project"`
	Zname      string       `bson:"zname" json:"zname"`
	Module     string       `bson:"module" json:"module"`
	CreateTime int64        `bson:"create_time" json:"create_time"`
	Total      int64        `bson:"total" json:"total"`
	Status     bool         `bson:"status" json:"status"`
	Msg        string       `bson:"msg" json:"msg"`
}

type Error struct {
	ErrCode int
	ErrMsg  string
}

type Response struct {
	Status bool
	Msg    string
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
	confFile := path.Join(basePath, "converge_alert_mail_server.yaml")
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
	logFile := path.Join(logDir, conf.Log.Name)
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

func ClearExpireData(conf Config) {
	// 让日志目录保留指定日期范围内的日志
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

func ZeroTime() int64 {
	// 获取当前零点时刻的时间戳
	year := time.Now().Year()
	month := time.Now().Month()
	day := time.Now().Day()
	td := fmt.Sprintf("%d-%d-%d", year, month, day)
	if len(string(day)) == 1 {
		td = fmt.Sprintf("%d-0%d-%d", year, month, day)
	}
	loc, _ := time.LoadLocation("Asia/Shanghai")

	t, _ := time.ParseInLocation("2006-01-02", td, loc)
	return t.Unix()
}

func ConvertTime(timestamp int64) string {
	// 将时间戳转成字符串
	tm := time.Unix(timestamp, 0)
	return tm.Format("2006-01-02 15:04:05")
}

func GetTimeOutSummaryInfo(zname string, t1 int64, data []TimeoutCid) string {
	// 应用于模板语言，获取标题
	return fmt.Sprintf("%s，抽查 cid 总数量为 %d 个，异常 cid 数量为 %d 个。", zname, t1, len(data))
}

func CleanCycleData(data []DeviceCycleDoc) (reportData []DeviceCycleReportData) {
	// 清洗存储周期不一致的数据，如果是对象存储返回的存储周期为-1在这里面剔除掉，返回真实存储周期不一致的cid信息
	for _, item := range data {
		if len(item.Data) == 0 {
			continue
		}
		var tmp DeviceCycleReportData
		tmp.Zname = item.Zname
		for _, cid := range item.Data {
			if cid.OVideo != cid.MVideo {
				tmp.VideoNum++
			}
			if cid.OPIC != cid.MPIC {
				tmp.PicNum++
			}
			if (cid.OVideo != cid.MVideo && cid.OVideo != -1) || (cid.OPIC != cid.MPIC && cid.OPIC != -1) {
				flag := true
				for _, cr := range tmp.Data {
					if cr.CID == cid.CID {
						flag = false
						break
					}
				}
				if flag {
					tmp.Data = append(tmp.Data, cid)
				}
			}
		}
		reportData = append(reportData, tmp)
	}
	return
}

func GetCycleSummaryInfo(zname string, t1, t2 int) string {
	title := fmt.Sprintf("%s，视频存储周期不一致cid数量为 %d个，图片存储周期不一致cid数量为 %d个。", zname, t1, t2)
	return title
}

func IsDisplay(data []DeviceCycleReportData) string {
	// 控制“详细信息”这个标题是否显示，如果后续数据都为空，则不显示
	display := "none"
	for _,item := range data{
		if len(item.Data) > 0{
			display = "block"
		}
	}
	return display
}

func SendReport(conf Config, reportData interface{}, name, templateName, subject string, attached bool, logger *zap.Logger) error {
	//以html形式发送邮件
	basePath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	//定义邮箱服务器连接信息
	e := email.NewEmail()
	if attached {
		_, err := e.AttachFile(name)
		if err != nil {
			logger.Error(fmt.Sprintf("[SendReport] 发送附件失败。%v", err))
		}
	}

	e.From = conf.Mail.Send.From
	e.To = conf.Mail.Recive.Normal
	e.Subject = subject + " " + time.Now().Format("2006-01-02")
	// 自定义模板函数传入
	funcMap := template.FuncMap{"ConvertTime": ConvertTime, "GetTimeOutSummaryInfo": GetTimeOutSummaryInfo, "GetCycleSummaryInfo": GetCycleSummaryInfo,"IsDisplay":IsDisplay}
	t, err := template.New(templateName).Funcs(funcMap).ParseFiles(path.Join(basePath, templateName+".html"))

	if err != nil {
		logger.Error(fmt.Sprintf("[SendReport] 创建模板渲染失败. %v", err))
		return err
	}
	//t = template.Must(tp,err)
	//buffer是一个实现了读写方法的可变大小的字节缓冲
	bufferBody := new(bytes.Buffer)
	err = t.Execute(bufferBody, struct {
		ReportData interface{}
	}{reportData})
	if err != nil {
		logger.Error(fmt.Sprintf("[SendReport] 模板渲染失败. %v", err))
		return err
	}
	//html形式的消息
	e.HTML = bufferBody.Bytes()
	return e.Send(conf.Mail.Send.Server+":25", smtp.PlainAuth("", conf.Mail.Send.Username, conf.Mail.Send.Password, conf.Mail.Send.Server))
}

func SendMail(conf Config, body, subject string) error {
	//发送纯文本内容的邮件
	//定义邮箱服务器连接信息
	e := email.NewEmail()
	e.From = conf.Mail.Send.From
	e.To = conf.Mail.Recive.Admin
	e.Subject = subject + " " + time.Now().Format("2006-01-02")
	e.Text = []byte(body)
	return e.Send(conf.Mail.Send.Server+":25", smtp.PlainAuth("", conf.Mail.Send.Username, conf.Mail.Send.Password, conf.Mail.Send.Server))
}

func WriteExcel(uncids []DeviceCycleDoc, config Config, logger *zap.Logger) {
	var ExFile = path.Join(BaseDir, config.Log.Path, config.Log.Exfile)
	exHandle := xlsx.NewFile()
	esHeaders := []string{
		"CID", "设备名", "分组", "设备SN", "厂商", "型号", "软件版本", "build时间", "通配视频周期", "对象存储视频周期", "通配图片周期", "对象存储图片周期",
	}
	for _, project := range uncids {
		sheet, err := exHandle.AddSheet(project.Zname)
		if err != nil {
			logger.Error(fmt.Sprintf("[WriteExcel] 创建sheet失败. %v", err))
		}
		headerRow := sheet.AddRow()
		for _, header := range esHeaders {
			tmpCell := headerRow.AddCell()
			tmpCell.Value = header
		}
		for _, uncid := range project.Data {
			tmpRow := sheet.AddRow()
			cidCell := tmpRow.AddCell()
			cid := strconv.Itoa(int(uncid.CID))
			cidCell.Value = cid
			nameCell := tmpRow.AddCell()
			nameCell.Value = uncid.Name
			groupCell := tmpRow.AddCell()
			groupCell.Value = uncid.Group
			snCell := tmpRow.AddCell()
			snCell.Value = uncid.SN
			brandCell := tmpRow.AddCell()
			brandCell.Value = uncid.Brand
			modelCell := tmpRow.AddCell()
			modelCell.Value = uncid.Model
			softCell := tmpRow.AddCell()
			softCell.Value = uncid.SoftwareVersion
			buildCell := tmpRow.AddCell()
			buildCell.Value = uncid.SoftwareBuild
			mvCell := tmpRow.AddCell()
			mVideo := strconv.Itoa(int(uncid.MVideo))
			mvCell.Value = mVideo
			ovCell := tmpRow.AddCell()
			oVideo := strconv.Itoa(int(uncid.OVideo))
			ovCell.Value = oVideo
			mpCell := tmpRow.AddCell()
			mp := strconv.Itoa(int(uncid.MPIC))
			mpCell.Value = mp
			opCell := tmpRow.AddCell()
			op := strconv.Itoa(int(uncid.OPIC))
			opCell.Value = op
		}
	}
	err := exHandle.Save(ExFile)
	if err != nil {
		logger.Error(fmt.Sprintf("[WriteExcel] create excel failed. %v", err))
	} else {
		logger.Info("[WriteExcel] create excel success.")
	}
}
