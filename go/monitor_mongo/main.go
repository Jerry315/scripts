package main

import (
	"dev/monitor_mongo/common"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func query(config common.Config, logger *zap.Logger) {
	data := []map[string]string{}
	for _, instance := range config.Instance {
		session, err := mgo.Dial(instance.Address)
		if err != nil {
			continue
		}
		err = session.Ping()
		if err != nil {
			continue
		}
		cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("netstat -tlnp|grep %d | grep  mongod| awk '{print $7}'", instance.Port))
		result, err := cmd.Output()
		if err != nil {
			logger.Error(fmt.Sprintf("execute command failed. %v", err))
			os.Exit(1)
		}
		pid := strings.Split(string(result), "/")[0]
		cmd1 := exec.Command("/bin/bash", "-c", fmt.Sprintf("ps -ef | grep %s | grep -v grep | awk '{print $10}'", pid))
		result1, err := cmd1.Output()
		if err != nil {
			logger.Error(fmt.Sprintf("execute command failed. %v", err))
			os.Exit(1)
		}
		tmp := map[string]string{}
		tmp["{#PORT}"] = strconv.Itoa(instance.Port)
		tmp["{#CONF_FILE}"] = strings.Trim(string(result1), "\n")
		data = append(data, tmp)
	}
	queryStr := make(map[string][]map[string]string)
	queryStr["data"] = data
	b, err := json.Marshal(queryStr)
	if err != nil {
		logger.Error(fmt.Sprintf("execute command failed. %v", err))
		os.Exit(1)
	}
	fmt.Println(string(b))
}

func activeClient(url, item string, logger *zap.Logger) {
	serverStatus, err := common.GetServerStatus(url, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("get server active clients data failed. %v", err))
	}
	fmt.Println(serverStatus.GlobalLock.ActiveClients.Total)
	fmt.Println(reflect.ValueOf(&serverStatus.GlobalLock.ActiveClients).Elem().FieldByName(item).Int())
}

func asserts(url, item string, logger *zap.Logger) {
	serverStatus, err := common.GetServerStatus(url, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("get server status data failed. %v", err))
	}
	fmt.Println(reflect.ValueOf(&serverStatus.Asserts).Elem().FieldByName(item).Int())
}

func connect(url, item string, logger *zap.Logger) {
	serverStatus, err := common.GetServerStatus(url, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("get server connect data failed. %v", err))
	}
	fmt.Println(reflect.ValueOf(&serverStatus.Connections).Elem().FieldByName(item).String())
}

func extraInfo(url, item string, logger *zap.Logger) {
	serverStatus, err := common.GetServerStatus(url, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("get server extra info data failed. %v", err))
	}
	fmt.Println(reflect.ValueOf(&serverStatus.Extra_info).Elem().FieldByName(item).String())
}

func memory(url, item string, logger *zap.Logger) {
	serverStatus, err := common.GetServerStatus(url, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("get server memery data failed. %v", err))
	}
	fmt.Println(reflect.ValueOf(&serverStatus.Mem).Elem().FieldByName(item).Int())
}

func network(url, item string, logger *zap.Logger) {
	serverStatus, err := common.GetServerStatus(url, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("get server network data failed. %v", err))
	}
	fmt.Println(reflect.ValueOf(&serverStatus.Network).Elem().FieldByName(item).Int())
}

func queue(url, item string, logger *zap.Logger) {
	serverStatus, err := common.GetServerStatus(url, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("get server queue data failed. %v", err))
	}
	fmt.Println(reflect.ValueOf(&serverStatus.GlobalLock.CurrentQueue).Elem().FieldByName(item).Int())
}

func opcounters(url, item string, logger *zap.Logger) {
	serverStatus, err := common.GetServerStatus(url, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("get server opcounters data failed. %v", err))
	}
	fmt.Println(reflect.ValueOf(&serverStatus.Opcounters).Elem().FieldByName(item).Int())
}

func getPid(config string, logger *zap.Logger) {
	cmdStr := fmt.Sprintf("ps -ef | grep %s | grep -v grep | grep -v pid | awk '{print $2}'", config)
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	result, err := cmd.Output()
	if err != nil {
		logger.Error(fmt.Sprintf("execute command failed. %v", err))
		os.Exit(1)
	}
	fmt.Println(strings.Trim(string(result), "\n"))
}

func replHealth(url string, logger *zap.Logger) {
	replData, err := common.GetReplStatus(url, logger)
	flag := 1
	if err != nil {
		logger.Error(fmt.Sprintf("get repl data failed. %v", err))
		flag = 0
	}
	for _, member := range replData.Members {
		if member.Health != 1 {
			flag = 0
		}
	}
	fmt.Println(flag)
}

func replRelay(url string, logger *zap.Logger) {
	replData, err := common.GetReplStatus(url, logger)
	var primaryOptime, secondaryOptime time.Time
	if err != nil {
		logger.Error(fmt.Sprintf("get repl data failed. %v", err))
	}
	for _, member := range replData.Members {
		if member.StateStr == "PRIMARY" {
			primaryOptime = member.OptimeDate
		} else if member.StateStr == "SECONDARY" {
			secondaryOptime = member.OptimeDate
		}
	}

	sl := primaryOptime.Unix() - secondaryOptime.Unix()
	fmt.Println(sl)
}

func uptime(url string, logger *zap.Logger) {
	serverStatus, err := common.GetServerStatus(url, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("get server status data failed. %v", err))
	}
	fmt.Println(serverStatus.Uptime)
}

func main() {
	config := common.GetConf()
	logger := common.InitLogger(config.Log.FileName, config.Log.Level)
	app := cli.NewApp()
	app.Name = "monitor_mongo"
	app.Commands = []cli.Command{
		{
			Name:        "query",
			Aliases:     []string{"q"},
			Usage:       "query mongodb info",
			Description: "query mongodb info",
			Action: func(c *cli.Context) {
				query(config, logger)
			},
		},
		{
			Name:        "pid",
			Aliases:     []string{"pid"},
			Usage:       "get mongodb process id",
			Description: "get mongodb process id",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config,c",
					Value: "",
					Usage: "mongodb config file name",
				},
			},
			Action: func(c *cli.Context) {
				confFile := c.String("config")
				getPid(confFile, logger)
			},
		},
		{
			Name:        "status",
			Aliases:     []string{"s"},
			Usage:       "check mongodb cluster health status",
			Description: "check mongodb cluster health status",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Value: -1,
					Usage: "mongodb listen port",
				},
			},
			Action: func(c *cli.Context) {
				port := c.Int("port")
				if port == -1 {
					port = 27017
				}
				for _, instance := range config.Instance {
					if instance.Port == port {
						replHealth(instance.Address, logger)
					}
				}
			},
		},
		{
			Name:        "replrelay",
			Aliases:     []string{"replrelay"},
			Usage:       "get mongodb repl relay time",
			Description: "get mongodb repl relay time",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Value: -1,
					Usage: "mongodb listen port",
				},
			},
			Action: func(c *cli.Context) {
				port := c.Int("port")
				if port == -1 {
					port = 27017
				}
				for _, instance := range config.Instance {
					if instance.Port == port {
						replRelay(instance.Address, logger)
					}
				}
			},
		},
		{
			Name:        "uptime",
			Aliases:     []string{"u"},
			Usage:       "check mongodb uptime",
			Description: "check mongodb uptime",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Value: -1,
					Usage: "mongodb listen port",
				},
			},
			Action: func(c *cli.Context) {
				port := c.Int("port")
				if port == -1 {
					port = 27017
				}
				for _, instance := range config.Instance {
					if instance.Port == port {
						uptime(instance.Address, logger)
					}
				}
			},
		},
		{
			Name:        "asserts",
			Aliases:     []string{"a"},
			Usage:       "check mongodb asserts",
			Description: "check mongodb asserts",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Value: -1,
					Usage: "mongodb listen port",
				},
				cli.StringFlag{
					Name:  "item,i",
					Value: "",
					Usage: "asserts item",
				},
			},
			Action: func(c *cli.Context) {
				port := c.Int("port")
				if port == -1 {
					port = 27017
				}
				item := c.String("item")
				for _, instance := range config.Instance {
					if instance.Port == port {
						asserts(instance.Address, item, logger)
					}
				}
			},
		},
		{
			Name:        "connect",
			Aliases:     []string{"c"},
			Usage:       "check mongodb connect",
			Description: "check mongodb connect",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Value: -1,
					Usage: "mongodb listen port",
				},
				cli.StringFlag{
					Name:  "item,i",
					Value: "",
					Usage: "connect info item",
				},
			},
			Action: func(c *cli.Context) {
				port := c.Int("port")
				if port == -1 {
					port = 27017
				}
				item := c.String("item")
				for _, instance := range config.Instance {
					if instance.Port == port {
						connect(instance.Address, item, logger)
					}
				}
			},
		},
		{
			Name:        "memory",
			Aliases:     []string{"m"},
			Usage:       "check mongodb memory info",
			Description: "check mongodb memory info",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Value: -1,
					Usage: "mongodb listen port",
				},
				cli.StringFlag{
					Name:  "item,i",
					Value: "",
					Usage: "memory info item",
				},
			},
			Action: func(c *cli.Context) {
				port := c.Int("port")
				if port == -1 {
					port = 27017
				}
				item := c.String("item")
				for _, instance := range config.Instance {
					if instance.Port == port {
						memory(instance.Address, item, logger)
					}
				}
			},
		},
		{
			Name:        "extraInfo",
			Aliases:     []string{"e"},
			Usage:       "check mongodb extraInfo",
			Description: "check mongodb extraInfo",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Value: -1,
					Usage: "mongodb listen port",
				},
				cli.StringFlag{
					Name:  "item,i",
					Value: "",
					Usage: "extraInfo item",
				},
			},
			Action: func(c *cli.Context) {
				port := c.Int("port")
				if port == -1 {
					port = 27017
				}
				item := c.String("item")
				for _, instance := range config.Instance {
					if instance.Port == port {
						extraInfo(instance.Address, item, logger)
					}
				}
			},
		},
		{
			Name:        "network",
			Aliases:     []string{"n"},
			Usage:       "check mongodb network info",
			Description: "check mongodb network info",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Value: -1,
					Usage: "mongodb listen port",
				},
				cli.StringFlag{
					Name:  "item,i",
					Value: "",
					Usage: "network info item",
				},
			},
			Action: func(c *cli.Context) {
				port := c.Int("port")
				if port == -1 {
					port = 27017
				}
				item := c.String("item")
				for _, instance := range config.Instance {
					if instance.Port == port {
						network(instance.Address, item, logger)
					}
				}
			},
		},
		{
			Name:        "query",
			Aliases:     []string{"query"},
			Usage:       "check mongodb current queue info",
			Description: "check mongodb current queue info",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Value: -1,
					Usage: "mongodb listen port",
				},
				cli.StringFlag{
					Name:  "item,i",
					Value: "",
					Usage: "current queue info item",
				},
			},
			Action: func(c *cli.Context) {
				port := c.Int("port")
				if port == -1 {
					port = 27017
				}
				item := c.String("item")
				for _, instance := range config.Instance {
					if instance.Port == port {
						queue(instance.Address, item, logger)
					}
				}
			},
		},
		{
			Name:        "activeClient",
			Usage:       "check mongodb current active clients info",
			Description: "check mongodb current active clients info",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Value: -1,
					Usage: "mongodb listen port",
				},
				cli.StringFlag{
					Name:  "item,i",
					Value: "",
					Usage: "current active clients info item",
				},
			},
			Action: func(c *cli.Context) {
				port := c.Int("port")
				if port == -1 {
					port = 27017
				}
				item := c.String("item")
				for _, instance := range config.Instance {
					if instance.Port == port {
						activeClient(instance.Address, item, logger)
					}
				}
			},
		},
		{
			Name:        "opcounters",
			Usage:       "check mongodb opcounters info",
			Description: "check mongodb opcounters info",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Value: -1,
					Usage: "mongodb listen port",
				},
				cli.StringFlag{
					Name:  "item,i",
					Value: "",
					Usage: "opcounters info item",
				},
			},
			Action: func(c *cli.Context) {
				port := c.Int("port")
				if port == -1 {
					port = 27017
				}
				item := c.String("item")
				for _, instance := range config.Instance {
					if instance.Port == port {
						opcounters(instance.Address, item, logger)
					}
				}
			},
		},
	}
	app.Run(os.Args)
}
