package stdlog

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 日志类型
var (
	Basic  = "basic"  // 基础日志
	Data   = "data"   // 数据日志
	Detail = "detail" // 详细日志, 可以打印详细的堆栈信息

	logTypeEncoders = map[string]string{
		Basic:  Console,
		Data:   JSON,
		Detail: Console,
	}
)

// Logger 日志
type Logger struct {
	lg          *zap.SugaredLogger
	options     *Options
	hostname    string
	logType     string
	encoderType string
}

// NewBasic 创建 Basic 类型日志器, 封装 New 实现
func NewBasic(options *Options) (*Logger, error) {
	return New(Basic, options)
}

// NewData 创建 Data 类型日志器, 封装 New 实现
func NewData(options *Options) (*Logger, error) {
	return New(Data, options)
}

// NewDetail 创建 Detail 类型日志器, 封装 New 实现
func NewDetail(options *Options) (*Logger, error) {
	return New(Detail, options)
}

// New 创建日志, 支持 `Basic: 基础日志`, `Data: 数据日志`, `Detail: 详细日志`
func New(logType string, options *Options) (*Logger, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	core, err := getZapCore(logType, options)
	if err != nil {
		return nil, err
	}
	encoderType := logTypeEncoders[logType]
	return &Logger{
		lg:          zap.New(core).Sugar(),
		options:     options,
		hostname:    hostname,
		logType:     logType,
		encoderType: encoderType,
	}, nil
}

// Debug 写入 Debug 级别日志
func (l *Logger) Debug(logType string, pairs ...interface{}) {
	pairs = l.appendFixedFields(logType, pairs...)
	l.lg.Debugw("", pairs...)
}

// Info 写入 Info 级别日志
func (l *Logger) Info(logType string, pairs ...interface{}) {
	pairs = l.appendFixedFields(logType, pairs...)
	l.lg.Infow("", pairs...)
}

// Warn 写入 Warn 级别日志
func (l *Logger) Warn(logType string, pairs ...interface{}) {
	pairs = l.appendFixedFields(logType, pairs...)
	l.lg.Warnw("", pairs...)
}

// Error 写入 Error 级别日志
func (l *Logger) Error(logType string, pairs ...interface{}) {
	pairs = l.appendFixedFields(logType, pairs...)
	l.lg.Errorw("", pairs...)
}

// Fatal 写入 Fatal 级别日志
func (l *Logger) Fatal(logType string, pairs ...interface{}) {
	pairs = l.appendFixedFields(logType, pairs...)
	l.lg.Fatalw("", pairs...)
}

// appendFixedFields 添加固定的字段数据
func (l *Logger) appendFixedFields(logType string, pairs ...interface{}) []interface{} {
	var pre []interface{}
	if l.encoderType != Console {
		pre = []interface{}{"log_type", logType, "hostname", l.hostname}
	} else {
		pre = []interface{}{"log_type", logType}
	}
	pairs = append(pre, pairs...)
	return pairs
}

// getZapCore 根据日志类型及选项返回 zapcore.Core
func getZapCore(logType string, options *Options) (zapcore.Core, error) {
	conf := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	var encoder zapcore.Encoder
	encoderType, exists := logTypeEncoders[logType]
	if !exists {
		return nil, fmt.Errorf("getZapCore: invalid logType %s", logType)
	}
	detail := logType == Detail
	switch encoderType {
	case Console:
		conf.EncodeLevel = consoleLevelEncoder
		conf.EncodeTime = consoleTimeEncoder
		encoder = NewConsoleEncoder(conf, detail)
	case JSON:
		conf.EncodeLevel = jsonLevelEncoder
		conf.EncodeTime = jsonTimeEncoder
		encoder = zapcore.NewJSONEncoder(conf)
	}

	writeSyncer := zapcore.AddSync(options.getWriteSyncer())
	return zapcore.NewCore(encoder, writeSyncer, options.ZapLevel()), nil
}
