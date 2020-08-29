package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Bind  string `yaml:"bind"`
	Port  int    `yaml:"port"`
	Https int    `yaml:"https"`
}

func GetConf() Config {
	conf := new(Config)
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("The basePath failed: %s\n", err.Error())
	}
	confFile := path.Join(basePath, "parse_lvs.yaml")
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

func getRs(w http.ResponseWriter, r *http.Request) {
	config := GetConf()
	r.ParseForm()
	host := r.Form.Get("host")
	cmd := exec.Command("/bin/bash", "-c", "ipvsadm -Ln")
	//创建获取命令输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error can not obtain stdout pipe for command")
		return
	}

	//执行命令
	if err := cmd.Start(); err != nil {
		fmt.Println("Error the command is err,", err)
		return
	}

	//读取所有输出
	b, err := ioutil.ReadAll(stdout)
	if err != nil {
		fmt.Println("ReadAll Stdout: ", err.Error())
		return
	}
	if err := cmd.Wait(); err != nil {
		fmt.Println("wait: ", err.Error())
		return
	}
	flag := false
	hosts := []string{}
	for _, line := range strings.Split(string(b), "\n") {
		newLine := strings.TrimSpace(line)
		if strings.HasPrefix(newLine, "TCP") {
			flag = false
		}
		if strings.HasPrefix(newLine, "TCP") && strings.Contains(newLine, host+":"+strconv.Itoa(config.Https)) {
			flag = true
		}
		if flag {
			if strings.HasPrefix(newLine, "TCP") {
				hosts = append(hosts, strings.TrimSpace(strings.Split(strings.Split(newLine, "  ")[1], " ")[0]))
			} else {
				hosts = append(hosts, strings.TrimSpace(strings.Split(newLine, " ")[1]))
			}
		}
	}
	fmt.Fprintf(w, strings.Join(hosts, ";"))
}

func main() {
	http.HandleFunc("/getrs/", getRs)
	config := GetConf()
	err := http.ListenAndServe(config.Bind+":"+strconv.Itoa(config.Port), nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}
}
