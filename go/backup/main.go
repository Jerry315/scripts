package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"time"
)

type Config struct {
	Layout string   `yaml:"layout"`
	Expire int64      `yaml:"expire"`
	Files  []string `yaml:"files"`
	Dirs   []struct{
		Dir string `yaml:"dir"`
		Cmd []string `yaml:"cmd"`
	}
	Dbs    []struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	}
	Backup string `yaml:"backup"`
}

func parseConf(confFile string) *Config {
	conf := new(Config)
	confData, err := ioutil.ReadFile(confFile)
	if err != nil {
		fmt.Printf("confFile Get err %#v", err)
	}
	err = yaml.Unmarshal(confData, conf)
	if err != nil {
		fmt.Printf("Unmarshal: %#v", err)
	}
	return conf
}

func backupFiles(conf *Config) {
	err := os.Chdir(conf.Backup)
	if err != nil {
		fmt.Printf("backup dir is not exist. %#v", err)
		os.Exit(1)
	}
	ts := time.Now().Format(conf.Layout)
	for _, file := range conf.Files {
		_, fileName := filepath.Split(file)
		cmdStr := fmt.Sprintf("cp %s %s-%s", file, fileName, ts)
		cmd := exec.Command("/bin/bash", "-c", cmdStr)
		result, err := cmd.Output()
		if err != nil {
			fmt.Printf("%#v", err)
			os.Exit(1)
		}
		fmt.Println(string(result))
	}
}

func backupDirs(conf *Config) {
	err := os.Chdir(conf.Backup)
	if err != nil {
		fmt.Printf("backup dir is not exist. %#v", err)
		os.Exit(1)
	}
	for _, item := range conf.Dirs {
		for _,cmd := range item.Cmd {
			result := exec.Command("/bin/bash", "-c", cmd)
			output, err := result.Output()
			if err != nil {
				fmt.Printf("cmd result error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(output))
		}
	}
}

func backupMysql(conf *Config) {
	err := os.Chdir(conf.Backup)
	if err != nil {
		fmt.Printf("backup dir is not exist. %#v", err)
		os.Exit(1)
	}
	ts := time.Now().Format(conf.Layout)
	for _, dbInfo := range conf.Dbs {
		cmdStr := fmt.Sprintf("mysqldump -u%s -p%s -h %s -P %d --single-transaction %s |gzip > %s.sql-%s.gz",
			dbInfo.Username, dbInfo.Password, dbInfo.Host, dbInfo.Port, dbInfo.Database, dbInfo.Database, ts)
		cmd := exec.Command("/bin/bash", "-c", cmdStr)
		result, err := cmd.Output()
		if err != nil {
			fmt.Printf("%#v", err)
			os.Exit(1)
		}
		fmt.Println(string(result))
	}
}

func clearExpireData(conf *Config) {
	nt := time.Now().Unix()
	et := nt - conf.Expire*86400
	reg := regexp.MustCompile(`\d{8}`)
	files, err := ioutil.ReadDir(conf.Backup)
	if err != nil {
		fmt.Printf("backup dir is not exist. %#v", err)
		os.Exit(1)
	}
	os.Chdir(conf.Backup)
	for _, fileName := range files {
		result := reg.FindAllString(fileName.Name(), -1)
		ft, _ := time.Parse(conf.Layout, result[0])
		if ft.Unix() < et{
			os.Remove(fileName.Name())
		}
	}
}

func main() {
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("The basePath failed: %s\n", err.Error())
	}
	confFile := path.Join(basePath, "backup.yaml")
	conf := parseConf(confFile)
	fmt.Println("start backup dirs")
	backupFiles(conf)
	backupDirs(conf)
	backupMysql(conf)
	clearExpireData(conf)
}
