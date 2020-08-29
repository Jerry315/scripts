package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
)

const ConfDir = "/opt/codproducer/etc"

type Yaml struct {
	General struct {
		RpcPort string `yaml:"rpcPort"`
		Queryport string `yaml:"queryport"`
	}
}

var port int

func init() {
	flag.IntVar(&port, "p", 80, "port")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: %s [Options] <IP>\n\nOptions:\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func GetConfFiles() []string {
	var confFiles []string
	files, _ := ioutil.ReadDir(ConfDir)
	for _, f := range files {
		confFiles = append(confFiles, path.Join(ConfDir, f.Name()))
	}
	return confFiles
}

func ParseConf(confFile string) Yaml {
	conf := new(Yaml)
	data, err := ioutil.ReadFile(confFile)
	if err != nil {
		fmt.Println(0)
	}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		fmt.Println(0)
	}
	return *conf
}

func Query() {
	confFiles := GetConfFiles()
	result := make(map[string][]map[string]string)
	result["data"] = make([]map[string]string, len(confFiles))
	i := 0
	for _, f := range confFiles {
		conf := ParseConf(f)
		tmp := make(map[string]string)
		tmp["{#CONF}"] = f
		tmp["{#RPCPORT}"] = strings.Split(conf.General.RpcPort,":")[1]
		tmp["{#QUERYPORT}"] = strings.Split(conf.General.Queryport,":")[1]
		result["data"][i] = tmp
		i++
	}
	b, err := json.Marshal(result)
	if err != nil {
		fmt.Println(0)
	}
	fmt.Println(string(b))
}

func Check(port int) {
	cmmand1 := fmt.Sprintf("netstat -anp|grep %d | grep codproducer | awk '{print $4}'", port)
	cmd1 := exec.Command("/bin/bash", "-c", cmmand1)

	var out1 bytes.Buffer
	cmd1.Stdout = &out1

	err := cmd1.Run()
	if err != nil {
		fmt.Println(err)
	}
	result := out1.String()
	host := strings.Split(result, ":")[0]
	if len(host) == 0 {
		host = "127.0.0.1"
	}
	ip := net.ParseIP(host)
	tcpAddr := net.TCPAddr{
		IP:   ip,
		Port: port,
	}
	conn, err := net.DialTCP("tcp", nil, &tcpAddr)
	if err == nil {
		defer conn.Close()
		command2 := fmt.Sprintf("netstat -anp|grep %d | grep codproducer | awk '{print $7}'", port)
		cmd2 := exec.Command("/bin/bash", "-c", command2)

		var out2 bytes.Buffer
		cmd2.Stdout = &out2

		err := cmd2.Run() /**/
		if err != nil {
			fmt.Println(0)
		}
		resp := out2.String()
		fmt.Println(strings.Split(resp, "/")[0])
	} else {
		fmt.Println(0)
	}
}

func main() {
	args := flag.Args()
	if len(args) == 1 {
		Query()
	} else {
		Check(port)
	}
}
