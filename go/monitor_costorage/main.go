package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
)

const ConfFile = "/opt/COStorage/storage.conf"

var port int

type RecurlyXml struct {
	Config Partitions `xml:"partitions"`
}

type Partitions struct {
	Partitions []Partition `xml:"partition"`
}

type Partition struct {
	Enable      int `xml:"enable"`
	Upload_port int `xml:"upload_port"`
}

func init() {
	flag.IntVar(&port, "p", 80, "port")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: %s [Options] <IP>\n\nOptions:\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func parseXml() RecurlyXml {
	content, err := ioutil.ReadFile(ConfFile)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	var parts RecurlyXml
	err = xml.Unmarshal(content, &parts)
	if err != nil {
		fmt.Printf("%#v",err)
	}
	return parts
}

func query()  {
	parts := parseXml()
	data := make(map[string][]map[string]string)
	l := 0
	for _,item := range parts.Config.Partitions {
		if item.Enable == 1{
			l++
		}
	}
	data["data"]=make([]map[string]string,l)
	i := 0
	for _,item := range parts.Config.Partitions {
		if item.Enable == 1{
			tmp := make(map[string]string)
			tmp["{#UPLOAD_PORT}"]=strconv.Itoa(item.Upload_port)
			data["data"][i] = tmp
			i++
		}
	}
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println(0)
	}
	fmt.Println(string(b))

}

func check(port int) {
	host := "127.0.0.1"
	ip := net.ParseIP(host)
	tcpAddr := net.TCPAddr{
		IP:   ip,
		Port: port,
	}
	conn, err := net.DialTCP("tcp", nil, &tcpAddr)
	if err == nil {
		defer conn.Close()
		fmt.Println(1)
	} else {
		fmt.Println(0)
	}
}

func main() {
	args := flag.Args()
	if 1 == len(args) {
		query()
	} else {
		check(port)
	}
}
