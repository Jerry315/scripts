package main

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
	"strings"
	"time"
)

type Config struct {
	Dirs []struct {
		Dir       string   `yaml:"dir"`
		Cycle     int      `yaml:"cycle"`
		Recursion bool     `yaml:"recursion"`
		Suffix    []string `yaml:"suffix"`
	}
	Whitelist []string `yaml:"whitelist"`
	Log       struct {
		Level    string `yaml:"level"`
		Path     string `yaml:"path"`
		FileName string `yaml:"fileName"`
	}
}

func GetConf() Config {
	conf := new(Config)
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("The basePath failed: %s\n", err.Error())
	}
	confFile := path.Join(basePath, "clean_log_file.yaml")
	confData, err := ioutil.ReadFile(confFile)
	if err != nil {
		fmt.Printf("confFile Get err %#v", err)
	}
	err = yaml.Unmarshal(confData, conf)
	if err != nil {
		fmt.Printf("Unmarshal: %#v", err)
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

func cleanAllFileRecursion(path string, whitelist, suffix []string, cycle int, recursion bool, logger *zap.Logger) {
	_, err := os.Stat(path)
	if err != nil {
		logger.Error(fmt.Sprintf("path: %s is not exist.", path))
		return
	}
	if path != "" {
		nt := time.Now().Unix()
		et := int(nt) - cycle*86400
		files, err := ioutil.ReadDir(path)
		if err != nil {
			logger.Error(fmt.Sprintf("read %s failed", path))
		}
		for _, f := range files {
			if f.IsDir() {
				if recursion {
					flag := false
					rPath := filepath.Join(path, f.Name())
					for _, w := range whitelist {
						if w == rPath {
							flag = true
							break
						}
					}
					if flag {
						continue
					}
					rFiles, err := ioutil.ReadDir(rPath)
					if err != nil {
						logger.Error(fmt.Sprintf("read %s failed", rPath))
					}
					if len(rFiles) == 0 {
						err := os.Remove(rPath)
						if err != nil {
							logger.Error(fmt.Sprintf("remove %s failed", rPath))
						} else {
							logger.Info(fmt.Sprintf("remove %s success", rPath))
						}
					} else {
						cleanAllFileRecursion(rPath, whitelist, suffix, cycle, recursion, logger)
					}
				}

			} else {
				if int(f.ModTime().Unix()) <= et {
					rFile := filepath.Join(path, f.Name())
					flag := true

					for _, s := range suffix {
						if strings.HasSuffix(rFile, s) {
							flag = false
							break
						}
					}
					if len(suffix) == 0{
						flag = false
					}
					if flag {
						continue
					}
					err := os.Remove(rFile)
					if err != nil {
						logger.Error(fmt.Sprintf("remove %s failed", filepath.Join(path, f.Name())))
					} else {
						logger.Info(fmt.Sprintf("remove %s success", filepath.Join(path, f.Name())))
					}
				}
			}
		}
	}

}

func main() {
	conf := GetConf()
	var logFile string
	basePath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	if conf.Log.Path == "" {
		logFile = path.Join(basePath, conf.Log.FileName)
	} else {
		logFile = path.Join(conf.Log.Path, conf.Log.FileName)
	}
	logger := InitLogger(logFile, conf.Log.Level)
	for _, d := range conf.Dirs {
		cycle := d.Cycle
		if cycle == 0 {
			cycle = 30
		}
		flag := false
		for _, w := range conf.Whitelist {
			if d.Dir == w {
				flag = true
			}
		}
		if flag {
			continue
		}
		cleanAllFileRecursion(d.Dir, conf.Whitelist, d.Suffix, cycle, d.Recursion, logger)
	}

}
