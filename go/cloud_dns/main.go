package main

import (
	"dev/cloud_dns/handle"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

type Yaml struct {
	Aliyun struct {
		AK       string `yaml:"ak"`
		SID      string `yaml:"sid"`
		RID      string `yaml:"rid"`
		ENDPOINT string `yaml:"endpoint"`
		FMT      string `yaml:"fmt"`
		VERSION  string `yaml:"version"`
		PROTOCOL string `yaml:"protocol"`
		METHOD   string `yaml:"method"`
	}
}


func AliClient(conf Yaml) (client *alidns.Client) {
	// 生成调用阿里云接口client
	client, err := alidns.NewClientWithAccessKey(conf.Aliyun.RID, conf.Aliyun.AK, conf.Aliyun.SID)
	if err != nil {
		log.Fatal("something wrong with your client accesskey")
	}
	return client
}

func GetConf() Yaml {
	// 获取配置文件，并返回
	conf := new(Yaml)
	yamlFile, err := ioutil.ReadFile("cloud_dns.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err #%v", err)
	}
	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return *conf
}

func GetArgv() *cli.App {
	// 获取系统传入的参数
	app := cli.NewApp()
	app.Name = "Cloud parsing"
	app.Usage = "create or get domain info"
	app.Flags = []cli.Flag {
		cli.StringFlag{
			Name:"method,m",
			Value: "",
			Usage: "获取指定域名下记录信息：GetDomainRecords\n" +
				"\t获取指定记录的详细信息：GetDomainRecordInfo\n" +
				"\t获取某个固定子域名的所有解析记录：GetSubDomainRecord\n" +
				"\t根据传入参数添加解析记录：AddDomainRecord\n" +
				"\t根据传入参数删除解析记录：DelDomainRecord\n" +
				"\t根据传入参数修改解析记录：UpdateDomainRecord\n" +
				"\t根据传入参数获取设置解析记录状态：SetDomainRecord",
		},
		cli.StringFlag{
			Name:"domain,d",
			Value: "",
			Usage:"根域名",
		},
		cli.StringFlag{
			Name:"subdomain,sd",
			Value:"",
			Usage:"获取某个固定子域名的所有解析记录列表",
		},
		cli.StringFlag{
			Name:"RecordId,rid",
			Value:"",
			Usage:"唯一请求识别码",
		},
		cli.StringFlag{
			Name:"Type,t",
			Value:"A",
			Usage:"解析类型包括(不区分大小写)：A、MX、CNAME、TXT、DIRECT_URL、FORWORD_URL、NS、AAAA、SRV",
		},
		cli.StringFlag{
			Name:"RR,r",
			Value:"",
			Usage:" 主机名，www",
		},
		cli.StringFlag{
			Name:"Value,V",
			Value:"",
			Usage:"	记录值,ip",
		},
		cli.StringFlag{
			Name:"Status,s",
			Value:"Enable",
			Usage:"Enable or Disable",
		},
		cli.IntFlag{
			Name:"PageNumber,pn",
			Value:1,
			Usage:"当前页数，起始值为1，默认为1",
		},
		cli.IntFlag{
			Name:"PageSize,pg",
			Value:20,
			Usage:"分页查询时设置的每页行数，最大值500，默认为20",
		},
	}
	return app
}


func main() {
	conf := GetConf()
	client := AliClient(conf)
	app := GetArgv()
	app.Action = func(c *cli.Context) error {
		method := c.String("method")
		domain := c.String("domain")
		subdomain := c.String("subdomain")
		recordid := c.String("RecordId")
		_type := c.String("Type")
		rr := c.String("RR")
		value := c.String("Value")
		status := c.String("Status")
		pn := c.Int("PageNumber")
		pg := c.Int("PageSize")
		PageNumber := requests.NewInteger(pn)
		PageSize := requests.NewInteger(pg)
		if method == "GetDomainRecords" {
			if domain == ""{
				fmt.Println("domain argument is must")
			}
			handle.GetDomainRecords(client,domain,PageNumber,PageSize)
		} else if method == "GetDomainRecordInfo"{
			if recordid == ""{
				fmt.Println("RecordId argument is must")
			}
			handle.GetDomainRecordInfo(client,recordid)
		} else if method == "GetSubDomainRecord"{
			if subdomain == ""{
				fmt.Println("subdomain argument is must")
			}
			handle.GetSubDomainRecord(client,subdomain,PageNumber,PageSize)
		} else if method == "AddDomainRecord" {
			if domain == "" || rr == "" || value == ""||_type == "" {
				fmt.Println("missing parameter")
			}
			handle.AddDomainRecord(client,domain,rr,value,_type)
		} else if method == "DelDomainRecord" {
			if recordid == ""{
				fmt.Println("RecordId argument is must")
			}
			handle.DelDomainRecord(client,recordid)
		} else if method == "UpdateDomainRecord" {
			if recordid == "" || rr == "" || _type == "" || value == ""{
				fmt.Println("missing parameter")
			}
			handle.UpdateDomainRecord(client,recordid,rr,_type,value)
		} else if method == "SetDomainRecord" {
			if recordid == ""{
				fmt.Println("RecordId argument is must")
			}
			handle.SetDomainRecord(client,recordid,status)
		}
		return nil
	}
	app.Run(os.Args)
}
