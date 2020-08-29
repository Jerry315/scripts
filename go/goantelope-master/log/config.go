package log

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	levels = map[string]zapcore.Level{
		"debug":  zapcore.DebugLevel,
		"info":   zapcore.InfoLevel,
		"warn":   zapcore.WarnLevel,
		"error":  zapcore.ErrorLevel,
		"dpanic": zapcore.DPanicLevel,
		"panic":  zapcore.PanicLevel,
		"fatal":  zapcore.FatalLevel,
	}
	defaultLevel        = zapcore.InfoLevel
	defaultEncodeCaller = zapcore.ShortCallerEncoder
)

const (
	defaultEncoding      = "json"
	defaultLevelStr      = "info"
	defaultFlushCount    = 10
	defaultFlushInterval = 5
)

// Default value
const (
	DefaultLevel         = defaultLevelStr
	DefaultFlushCount    = defaultFlushCount
	DefaultFlushInterval = defaultFlushInterval
)

// Config 日志配置
type Config struct {
	// Level 日志级别, 支持 `debug`, `info`, `warn`, `error`, `dpanic`, `panic`, `fatal`
	Level string `json:"level" yaml:"level"`
	// Path 日志文件路径
	Path string `json:"path" yaml:"path"`
	// Development 是否为开发环境, 开发环境下如果记录的是错误日志会加上调用栈信息
	Development bool `json:"development" yaml:"development"`
	// FlushCount 缓存的日志记录数量, 达到或超过则进行刷盘
	FlushCount int `json:"flush_count" yaml:"flush_count"`
	// FlushInterval 缓存日志刷盘时间间隔, 单位秒
	FlushInterval int `json:"flush_interval" yaml:"flush_interval"`
	// EncodeCaller caller的路径格式, 支持`full`, `short`, 默认 `short`
	EncodeCaller string `json:"encode_caller" yaml:"encode_caller"`
	// Encoding 日志的编码方式, 支持 `json`, `console`, 默认 `json`
	Encoding string `json:"encoding" yaml:"encoding"`
}

// Validate 校验日志配置是否正确
func (cfg *Config) Validate() bool {
	if cfg.Path == "" {
		return false
	}
	if cfg.Level == "" {
		log.Printf("log: no level setup, use default level info")
	}
	return true
}

// level 返回配置的日志 level, 如果没配置或者错误则使用默认值
func (cfg *Config) level() zapcore.Level {
	l, ok := levels[cfg.Level]
	if !ok {
		l = defaultLevel
	}
	return l
}

// zapCfg 根据配置返回 zap 日志的配置
func (cfg *Config) zapCfg() zap.Config {

	encodeCaller := defaultEncodeCaller
	if cfg.EncodeCaller == "full" {
		encodeCaller = zapcore.FullCallerEncoder
	}

	encoding := defaultEncoding
	if cfg.Encoding == "console" {
		encoding = "console"
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:      "ts",
		LevelKey:     "l",
		NameKey:      "logger",
		CallerKey:    "caller",
		MessageKey:   "msg",
		EncodeLevel:  zapcore.LowercaseLevelEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeCaller: encodeCaller,
	}

	zapCfg := zap.Config{
		Level:         zap.NewAtomicLevel(),
		Development:   cfg.Development,
		Encoding:      encoding,
		EncoderConfig: encoderCfg,
		OutputPaths:   []string{cfg.Path},
	}
	zapCfg.Level.SetLevel(cfg.level())
	if zapCfg.Development {
		if !contain(zapCfg.OutputPaths, "stdout") {
			zapCfg.OutputPaths = append(zapCfg.OutputPaths, "stdout")
		}
		zapCfg.EncoderConfig.StacktraceKey = "stack"
	}
	return zapCfg
}

func contain(array []string, str string) bool {
	for _, value := range array {
		if value == str {
			return true
		}
	}

	return false
}
