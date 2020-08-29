package log

import (
	"context"
	"log"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Logger 日志器数据类型, 简单包装 zap 日志库
type Logger struct {
	name     string
	cfg      *Config
	zap      *zap.Logger
	zapSugar *zap.SugaredLogger

	// 日志 Sync 计数器, 粗略使用, 不需要加锁控制
	zapCounter      int
	zapSugarCounter int

	ctx                context.Context
	cancelAutoSyncFunc context.CancelFunc
	once               sync.Once

	// 统计
	zapSyncs      int64
	zapSugarSyncs int64
}

// New 创建一个新的日志器
func New(
	name, level, path string, flushCount, flushInterval int, development bool, opts ...zap.Option,
) (*Logger, error) {

	cfg := &Config{
		Level:         level,
		Path:          path,
		Development:   development,
		FlushCount:    flushCount,
		FlushInterval: flushInterval,
	}

	return NewWithConfig(name, cfg, opts...)
}

// NewDefault 创建一个使用默认配置的日志器
func NewDefault(
	name, path string, development bool, opts ...zap.Option) (*Logger, error) {

	cfg := &Config{
		Level:         defaultLevelStr,
		Path:          path,
		Development:   development,
		FlushCount:    defaultFlushCount,
		FlushInterval: defaultFlushInterval,
	}

	return NewWithConfig(name, cfg, opts...)
}

// NewWithConfig 直接从配置创建一个新的日志器
func NewWithConfig(name string, cfg *Config, opts ...zap.Option) (*Logger, error) {

	if cfg.Path == "" {
		cfg.Path = "stdout"
		log.Printf("log: no out path specified, use stdout\n")
	}

	zapCfg := cfg.zapCfg()
	zapLogger, err := zapCfg.Build(opts...)
	if err != nil {
		return nil, err
	}

	ctx, cancelF := context.WithCancel(context.Background())
	return &Logger{
		name:     name,
		cfg:      cfg,
		zap:      zapLogger,
		zapSugar: zapLogger.Sugar(),

		ctx:                ctx,
		cancelAutoSyncFunc: cancelF,
		once:               sync.Once{},
	}, nil
}

// WithConfig 从配置创建一个新的日志器
func WithConfig(cfg *Config, opts ...zap.Option) (*Logger, error) {
	return New(
		"", cfg.Level, cfg.Path, cfg.FlushCount,
		cfg.FlushInterval, cfg.Development, opts...)
}

// Sync 主动刷盘同步日志
func (l *Logger) Sync() error {
	err := l.syncZapLogger()
	if err != nil {
		log.Printf("log: zap logger sync failed: %v\n", err)
		return err
	}
	err = l.syncZapSugarLogger()
	if err != nil {
		log.Printf("log: zap sugar logger sync failed: %v\n", err)
		return err
	}
	return nil
}

// RunSyncTicker 启动自动刷盘同步计时器
func (l *Logger) RunSyncTicker() {
	f := func() {
		go l.syncTicker(l.ctx)
	}
	l.once.Do(f)
}

// StopSyncTicker 停止自动刷盘同步计时器
func (l *Logger) StopSyncTicker() {
	log.Println("log: stoping logger sync ticker")
	l.cancelAutoSyncFunc()
	// 重置
	l.once = sync.Once{}
	err := l.Sync()
	if err != nil {
		log.Printf("log: stoping logger sync error %v\n", err)
	}
}

// syncZapLogger zap 日志同步刷盘
func (l *Logger) syncZapLogger() error {
	err := l.zap.Sync()
	if err != nil {
		return err
	}
	l.zapCounter = 0
	l.zapSyncs++
	return nil
}

// syncZapSugarLogger zap sugar 日志同步刷盘
func (l *Logger) syncZapSugarLogger() error {
	err := l.zapSugar.Sync()
	if err != nil {
		return err
	}
	l.zapSugarCounter = 0
	l.zapSugarSyncs++
	return nil
}

// trySync 检查并确定是否需要进行刷盘同步缓存数据
func (l *Logger) trySync() {
	if l.cfg.FlushCount == 0 {
		l.cfg.FlushCount = defaultFlushCount
	}
	flushCount := l.cfg.FlushCount
	if l.zapCounter >= flushCount {
		err := l.syncZapLogger()
		if err != nil {
			log.Printf("log: zap logger sync failed: %v\n", err)
		}
	}
	if l.zapSugarCounter >= flushCount {
		err := l.syncZapSugarLogger()
		if err != nil {
			log.Printf("log: zap sugar logger sync failed: %v\n", err)
		}
	}
}

// syncTicker 同步刷盘定时器
func (l *Logger) syncTicker(ctx context.Context) {
	log.Println("log: starting logger sync ticker")
	if l.cfg.FlushInterval == 0 {
		l.cfg.FlushInterval = defaultFlushInterval
	}
	flushInterval := l.cfg.FlushInterval
	timer := time.NewTicker(time.Second * time.Duration(flushInterval))
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			err := l.Sync()
			if err != nil {
				log.Printf("log: sync error %v\n", err)
			}
		}
	}
}

// zap.Logger 包装函数

// Debug zap logger wraper
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
	l.zapCounter++
	l.trySync()
}

// Info zap logger wraper
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
	l.zapCounter++
	l.trySync()
}

// Warn zap logger wraper
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
	l.zapCounter++
	l.trySync()
}

// Error zap logger wraper
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
	l.zapCounter++
	l.trySync()
}

// DPanic zap logger wraper
func (l *Logger) DPanic(msg string, fields ...zap.Field) {
	l.zap.DPanic(msg, fields...)
	l.zapCounter++
	l.trySync()
}

// Panic zap logger wraper
func (l *Logger) Panic(msg string, fields ...zap.Field) {
	l.zap.Panic(msg, fields...)
	l.zapCounter++
	l.trySync()
}

// Fatal zap logger wraper
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
	l.zapCounter++
	l.trySync()
}

// zap.SugaredLogger 包装函数

// SDebug zap sugared logger wraper
func (l *Logger) SDebug(msg string, keysAndValues ...interface{}) {
	l.zapSugar.Debugw(msg, keysAndValues...)
	l.zapSugarCounter++
	l.trySync()
}

// SInfo zap sugared logger wraper
func (l *Logger) SInfo(msg string, keysAndValues ...interface{}) {
	l.zapSugar.Infow(msg, keysAndValues...)
	l.zapSugarCounter++
	l.trySync()
}

// SWarn zap sugared logger wraper
func (l *Logger) SWarn(msg string, keysAndValues ...interface{}) {
	l.zapSugar.Warnw(msg, keysAndValues...)
	l.zapSugarCounter++
	l.trySync()
}

// SError zap sugared logger wraper
func (l *Logger) SError(msg string, keysAndValues ...interface{}) {
	l.zapSugar.Errorw(msg, keysAndValues...)
	l.zapSugarCounter++
	l.trySync()
}

// SDPanic zap sugared logger wraper
func (l *Logger) SDPanic(msg string, keysAndValues ...interface{}) {
	l.zapSugar.DPanicw(msg, keysAndValues...)
	l.zapSugarCounter++
	l.trySync()
}

// SPanic zap sugared logger wraper
func (l *Logger) SPanic(msg string, keysAndValues ...interface{}) {
	l.zapSugar.Panicw(msg, keysAndValues...)
	l.zapSugarCounter++
	l.trySync()
}

// SFatal zap sugared logger wraper
func (l *Logger) SFatal(msg string, keysAndValues ...interface{}) {
	l.zapSugar.Fatalw(msg, keysAndValues...)
	l.zapSugarCounter++
	l.trySync()
}
