package stdlog_test

import (
	"fmt"
	"sync"

	"git.topvdn.com/web/goantelope/stdlog"
)

func ExampleLogger() {
	options := stdlog.NewOptions()
	options.WithPath("test.log").WithLevel(stdlog.Info).WithMaxSize(10)
	basicLogger, err := stdlog.NewBasic(options)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		logType := "test_basic_log"
		basicLogger.Debug(logType, "key1", 12, "key2", "debughello", "name", "tony")
		basicLogger.Info(logType, "key1", 12, "key2", "infohello", "name", "tony")
		basicLogger.Warn(logType, "key1", 12, "key2", "warnhello", "name", "tony")
		basicLogger.Error(logType, "key1", 12, "key2", "errhello", "name", "tony")
		// basicLogger.Fatal(logType, "key1", 12, "key2", "fatalhello", "name", "tony")
	}()

	dataLogger, err := stdlog.NewData(options)
	if err != nil {
		panic(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logType := "test_data_log"
		dataLogger.Debug(logType, "key1", 12, "key2", "debughello", "name", "tony")
		dataLogger.Info(logType, "key1", 12, "key2", "infohello", "name", "tony")
		dataLogger.Warn(logType, "key1", 12, "key2", "warnhello", "name", "tony")
		dataLogger.Error(logType, "key1", 12, "key2", "errhello", "name", "tony")
	}()

	detailLogger, err := stdlog.NewDetail(options)
	if err != nil {
		panic(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logType := "test_detail_log"
		detailLogger.Debug(logType, "key1", 12, "key2", "debughello", "name", "tony")
		detailLogger.Info(logType, "key1", 12, "key2", "infohello", "name", "tony")
		detailLogger.Warn(logType, "key1", 12, "key2", "warnhello", "name", "tony")
		detailLogger.Error(logType, "key1", 12, "key2", "errhello", "name", "tony")
	}()

	wg.Wait()

	// placeholder code to run example in `go test -v`
	fmt.Println("ok")
	// Output:
	// ok
}
