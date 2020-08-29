package stdlog

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chow1937/lumberjack"
	"go.uber.org/zap/zapcore"
)

// 日志级别
const (
	Debug = "DEBUG"
	Info  = "INFO"
	Warn  = "WARN"
	Error = "ERROR"
	Fatal = "FATAL"
)

var (
	levels = map[string]bool{
		Debug: true,
		Info:  true,
		Warn:  true,
		Error: true,
		Fatal: true,
	}
)

// Encoder 类型
const (
	Console = "console"
	JSON    = "json"
)

// 缺省默认值
const (
	DefaultLevel   = Info    // 默认级别, INFO
	DefaultBackups = 7       // 默认备份数量, 7
	DefaultMaxSize = 50      // 默认大小, 50M
	DefaultEncoder = Console // 默认 Encoder

	defaultPathFMT  = "%s.default.log" // 默认文件路径, <processname>.default.log
	defaultZapLevel = zapcore.DebugLevel
)

// Options 日志选项
type Options struct {
	path    string
	level   string
	backups int
	maxSize int
}

// NewOptions 创建日志选项
func NewOptions() *Options {
	defaultPath := fmt.Sprintf(defaultPathFMT, filepath.Base(os.Args[0]))
	return &Options{
		path:    defaultPath,
		level:   DefaultLevel,
		backups: DefaultBackups,
		maxSize: DefaultMaxSize,
	}
}

// getWriteSyncer 从选项获取 zap 的日志同步器
func (options *Options) getWriteSyncer() *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:         options.path,
		MaxSize:          options.maxSize,
		MaxBackups:       options.backups,
		MaxAge:           28,
		LocalTime:        true,
		BackupNameFunc:   backupName,
		BackupTimeFormat: backupTimeFormat,
		BackupSep:        ".",
	}
}

// ZapLevel 根据配置项返回 zap 的 Level 值
func (options *Options) ZapLevel() zapcore.Level {
	m := map[string]zapcore.Level{
		Debug: zapcore.DebugLevel,
		Info:  zapcore.InfoLevel,
		Warn:  zapcore.WarnLevel,
		Error: zapcore.ErrorLevel,
		Fatal: zapcore.FatalLevel,
	}
	if level, exists := m[options.level]; exists {
		return level
	}
	return defaultZapLevel
}

// WithLevel 设置级别
func (options *Options) WithLevel(level string) *Options {
	if _, exists := levels[level]; exists {
		options.level = level
	}
	return options
}

// WithPath 设置输出路径
func (options *Options) WithPath(path string) *Options {
	if path != "" {
		options.path = path
	}
	return options
}

// WithBackups 设置保存的文件数量
func (options *Options) WithBackups(backups int) *Options {
	if backups > 0 {
		options.backups = backups
	}
	return options
}

// WithMaxSize 设置文件大小, 单位 M
func (options *Options) WithMaxSize(maxSize int) *Options {
	if maxSize > 0 {
		options.maxSize = maxSize
	}
	return options
}
