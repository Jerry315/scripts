package log

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	testLogFile  = "test.log"
	testLogFile2 = "test2.log"
)

func TestLoggerNormalWrite(t *testing.T) {
	assert := assert.New(t)

	cfg := &Config{
		Level:         "debug",
		Path:          testLogFile,
		Development:   false,
		FlushCount:    5,
		FlushInterval: 5,
	}
	logger, err := WithConfig(cfg)
	assert.Nil(err)
	assert.NotNil(logger)
	defer func() {
		err := os.Remove(testLogFile)
		assert.Nil(err)
	}()

	// 测试写入日志
	cnt := 35432
	i := cnt
	for i > 0 {
		logger.Debug("debug", zap.Int("number", i))
		i--
		logger.Info("info", zap.Int("number", i))
		i--
		logger.Warn("warn", zap.Int("number", i))
		i--
		logger.Error("error", zap.Int("number", i))
		i--
		logger.SDebug("sdebug", "number", i)
		i--
		logger.SInfo("sinfo", "number", i)
		i--
		logger.SWarn("swarn", "number", i)
		i--
		logger.SError("serror", "number", i)
		i--
	}
	err = logger.Sync()
	assert.Nil(err)

	// 检查写入的日志
	f, err := os.Open(testLogFile)
	assert.Nil(err)
	defer func() {
		err := f.Close()
		assert.Nil(err)
	}()

	logs := []map[string]interface{}{}
	decoder := json.NewDecoder(bufio.NewReader(f))
	for {
		log := map[string]interface{}{}
		err := decoder.Decode(&log)
		if err == io.EOF {
			break
		}
		assert.Nil(err)

		logs = append(logs, log)
	}

	assert.Equal(cnt, len(logs))

	idxLog := map[int]map[string]interface{}{}
	for _, log := range logs {
		num, ok := log["number"].(float64)
		assert.Equal(true, ok)
		idxLog[int(num)] = log
	}

	for i := cnt; i > 0; i-- {
		_, ok := idxLog[i]
		assert.Equal(true, ok)
	}
}

func TestPanicAndFatal(t *testing.T) {
	assert := assert.New(t)

	logger, err := New("", "info", testLogFile2, 1, 1, true)
	defer func() {
		err := os.Remove(testLogFile2)
		assert.Nil(err)
	}()

	assert.Nil(err)
	assert.NotNil(logger)

	// Panic
	assert.Panics(func() { logger.Panic("panic") })
	assert.Panics(func() { logger.DPanic("dpanic") })
	assert.Panics(func() { logger.SPanic("spanic") })
	assert.Panics(func() { logger.SDPanic("sdpanic") })
}

func TestSyncTicker(t *testing.T) {
	assert := assert.New(t)

	logger, err := New("", "info", testLogFile, 1, 1, false)
	assert.Nil(err)

	defer func() {
		err := os.Remove(testLogFile)
		assert.Nil(err)
	}()

	t.Log(logger)
	logger.RunSyncTicker()
	time.Sleep(time.Second * 3)
	logger.StopSyncTicker()

	t.Log(logger)
	logger.RunSyncTicker()
	time.Sleep(time.Second * 3)
	logger.StopSyncTicker()
	t.Log(logger)
}

func TestNoPath(t *testing.T) {
	assert := assert.New(t)

	logger, err := New("", "debug", "", 1, 1, false)
	assert.Nil(err)
	assert.NotNil(logger)

	logger.Debug("test debug")
	logger.Info("test info")
}

func TestNewWithConfig(t *testing.T) {
	assert := assert.New(t)

	cfg := &Config{
		Level:         "debug",
		Path:          "",
		Development:   false,
		FlushCount:    5,
		FlushInterval: 5,
		EncodeCaller:  "full",
		Encoding:      "console",
	}

	logger, err := NewWithConfig("", cfg, zap.AddCallerSkip(1))
	assert.Nil(err)
	assert.NotNil(logger)

	logger.Debug("test debug")
	logger.Info("test info")
}

func TestNewDefault(t *testing.T) {
	assert := assert.New(t)

	logger, err := NewDefault("", "", true)
	assert.Nil(err)
	assert.NotNil(logger)

	logger.Debug("test debug")
	logger.Info("test info")
}

func TestOutputPaths(t *testing.T) {
	assert := assert.New(t)

	logger, err := NewDefault("", "", true)
	assert.Nil(err)
	assert.NotNil(logger)

	assert.Equal(1, len(logger.cfg.zapCfg().OutputPaths))
}
