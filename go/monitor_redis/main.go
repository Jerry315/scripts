package main

import (
	"dev/monitor_redis/common"
	"dev/monitor_redis/handler"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"net"
	"os"
	"strconv"
	"strings"
)

func query(config common.Config, logger *zap.Logger) error {
	data := make(map[string][]map[string]string)
	var addresses []string
	for _, item := range config.Instance {
		address := strings.Split(item.Address, ":")
		ip := net.ParseIP(address[0])
		port, err := strconv.Atoi(address[1])
		if err != nil {
			logger.Error("strconv port string to int failed")
			return err
		}
		tcpAddr := net.TCPAddr{
			IP:   ip,
			Port: port,
		}
		_, err = net.DialTCP("tcp", nil, &tcpAddr)
		if err == nil {
			addresses = append(addresses, item.Address)
		}
	}
	fmt.Println(addresses)
	data["data"] = make([]map[string]string, len(addresses))
	for i, address := range addresses {
		tmp := make(map[string]string)
		tmp["{#ADDRESS}"] = address
		data["data"][i] = tmp
	}
	queryStr,err := json.Marshal(data)
	if err != nil{
		logger.Error("parse query string failed")
		return err
	}
	fmt.Println(string(queryStr))
	return nil
}

func main() {
	conf := common.GetConf()
	logger := common.InitLogger(conf.Log.FileName, conf.Log.Level)
	app := cli.NewApp()
	app.Name = "monitor redis db info."
	flag := false
	app.Commands = []cli.Command{
		{
			Name: "query",
			Action: func(c *cli.Context) {
				err := query(conf,logger)
				if err != nil{
					fmt.Println("error")
				}
			},
		},
		{
			Name: "cpu",
			Action: func(c *cli.Context) {
				address := c.String("address")
				for _,item := range conf.Instance{
					if item.Address == address{
						flag = true
						conn,err := handler.Handle(address,item.Password,item.Db,logger)
						if err != nil{
							fmt.Println(-1)
							os.Exit(-1)
						}
						err = handler.GetCpu(conn,logger)
						if err != nil{
							fmt.Println(-1)
							os.Exit(-1)
						}
					}
				}
				if flag == false{
					fmt.Println(-1)
				}
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "address,a",
					Usage: "redis address",
					Value: "",
				},
			},
		},
		{
			Name: "db",
			Action: func(c *cli.Context) {
				address := c.String("address")
				db := c.Int("db")
				for _,item := range conf.Instance{
					if item.Address == address{
						flag = true
						conn,err := handler.Handle(address,item.Password,item.Db,logger)
						if err != nil{
							fmt.Println(-1)
							os.Exit(-1)
						}
						err = handler.GetDb(conn,db,logger)
						if err != nil{
							fmt.Println(-1)
							os.Exit(-1)
						}
					}
				}
				if flag == false {
					fmt.Println(-1)
				}
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "address,a",
					Usage: "redis address",
					Value: "",
				},
				cli.IntFlag{
					Name: "db,d",
					Usage: "redis database number",
					Value: 0,
				},
			},
		},
		{
			Name: "memory",
			Action: func(c *cli.Context) {
				address := c.String("address")
				for _,item := range conf.Instance{
					if item.Address == address{
						flag = true
						conn,err := handler.Handle(address,item.Password,item.Db,logger)
						if err != nil{
							fmt.Println(-1)
							os.Exit(0)
						}
						err = handler.GetMemory(conn,logger)
						if err != nil{
							fmt.Println(-1)
							os.Exit(-1)
						}
					}
				}
				if flag == false {
					fmt.Println(-1)
				}
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "address,a",
					Usage: "redis address",
					Value: "",
				},
			},
		},
	}
	app.Run(os.Args)
}
