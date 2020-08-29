package main

import (
	"dev/converge_alert_mail/converge_alert_mail_server/common"
	"dev/converge_alert_mail/converge_alert_mail_server/mongo"
	"dev/converge_alert_mail/converge_alert_mail_server/router"
	"fmt"
	"github.com/urfave/cli"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var config = common.GetConf()
var logger = common.InitLogger()
var ExFile = path.Join(common.BaseDir, config.Log.Path, config.Log.Exfile)

func deviceCycleReport() {
	// 发送存储周期不一致性报告
	t := time.Now().Unix()
	reportXls := path.Join(common.BaseDir, config.Log.Path, config.Log.Exfile)
	if _, err := os.Stat(reportXls); err != nil {
		os.Create(reportXls)
	}
	collection, err := mongo.MongoClient(config.Mongodb.Url, config.Mongodb.Db, config.Mongodb.Table)
	if err != nil {
		logger.Error(fmt.Sprintf("mongodb init failed. %v", err))
		os.Exit(1)
	}
	var cycleData, data []common.DeviceCycleDoc
	var unusualMsg []string
	err = mongo.MgoQuery(collection, config.DeviceCycle.Name, t, &cycleData, logger)
	if err == nil {
		for _, item := range cycleData {
			flag := true
			if len(item.Data) == 0 {
				unusualMsg = append(unusualMsg, fmt.Sprintf("%s，存储周期一致性检测结果：%s.", item.Zname, item.Msg))
				continue
			}
			for _, d := range data {
				if item.Project == d.Project {
					flag = false
					break
				}
			}
			if flag {
				data = append(data, item)
			}
			if !item.Status {
				unusualMsg = append(unusualMsg, fmt.Sprintf("%s，存储周期一致性检测结果：%s.", item.Zname, item.Msg))
			}
		}
		common.WriteExcel(data, config, logger)
		cleanData := common.CleanCycleData(data)
		err = common.SendReport(config, cleanData, reportXls, config.DeviceCycle.Template, config.DeviceCycle.Subject, true, logger)
		if err != nil {
			logger.Error(fmt.Sprintf("发送存储周期一致性校验邮件失败. %v", err))
			err = common.SendMail(config, "发送存储周期一致性校验邮件失败", config.DeviceCycle.Subject)
			if err != nil {
				logger.Error(fmt.Sprintf("邮件服务不可用. %v", err))
			}
		} else {
			if len(unusualMsg) > 0 {
				common.SendMail(config, strings.Join(unusualMsg, "\n\r"), config.DeviceCycle.Subject)
			}
		}
	} else {
		logger.Error(fmt.Sprintf("获取存储周期一致性校验数据失败. %v", err))
		err = common.SendMail(config, "获取存储周期一致性校验数据失败", config.DeviceCycle.Subject)
		if err != nil {
			logger.Error(fmt.Sprintf("邮件服务不可用. %v", err))
		}
	}
}

func deviceTimeOutReport() {
	// 发送cid超时未推流和未上传图片的报告
	t := time.Now().Unix()
	collection, err := mongo.MongoClient(config.Mongodb.Url, config.Mongodb.Db, config.Mongodb.Table)
	if err != nil {
		logger.Error(fmt.Sprintf("mongodb init failed. %v", err))
		os.Exit(1)
	}
	var timeOUtData, data []common.DeviceTimeOutDoc
	err = mongo.MgoQuery(collection, config.DeviceTimeOut.Name, t, &timeOUtData, logger)
	if err == nil {
		for _, item := range timeOUtData {
			flag := true
			if len(item.Data) == 0 {
				continue
			}
			for _, d := range data {
				if item.Project == d.Project {
					flag = false
					break
				}
			}
			if flag {
				data = append(data, item)
			}
		}
		err = common.SendReport(config, data, "", config.DeviceTimeOut.Template, config.DeviceTimeOut.Subject, false, logger)
		if err != nil {
			logger.Error(fmt.Sprintf("发送cid超时邮件失败. %v", err))
			err = common.SendMail(config, "发送cid超时邮件失败", config.DeviceTimeOut.Subject)
			if err != nil {
				logger.Error(fmt.Sprintf("邮件服务不可用. %v", err))
			}
		}
	} else {
		logger.Error(fmt.Sprintf("获取cid超时数据失败. %v", err))
		err = common.SendMail(config, "获取cid超时数据失败", config.DeviceTimeOut.Subject)
		if err != nil {
			logger.Error(fmt.Sprintf("邮件服务不可用. %v", err))
		}
	}

}

func view() {
	// 视图函数，提供数据上报接口
	http.HandleFunc("/ops/alert/v1/login", router.LoginHandler)
	http.HandleFunc("/ops/alert/v1/device_cycle", router.DeviceCycleHandler)
	http.HandleFunc("/ops/alert/v1/device_timeout", router.DeviceTimeOutHandler)
	err := http.ListenAndServe(config.Bind+":"+config.Port, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("ListenAndServer error: %v", err))
	}
}

func main() {
	app := cli.NewApp()
	app.Version = "v0.0.1"
	app.Name = "converge_alert_mail_server"
	app.Commands = []cli.Command{
		{
			Name:        "dcr",
			Usage:       "发送全量对比通配数据库中cid和对象存储的cid图片和视频存储周期是否一致",
			Description: "发送全量对比通配数据库中cid和对象存储的cid图片和视频存储周期是否一致",
			Action: func(c *cli.Context) {
				deviceCycleReport()
			},
		},
		{
			Name:        "dtr",
			Usage:       "发送检测cid超时未推流和未上传图片",
			Description: "发送检测cid超时未推流和未上传图片",
			Action: func(c *cli.Context) {
				deviceTimeOutReport()
			},
		},
		{
			Name:        "server",
			Aliases:     []string{"s"},
			Usage:       "数据上报服务，开启后提供数据上报接口",
			Description: "数据上报服务，开启后提供数据上报接口",
			Action: func(c *cli.Context) {
				view()
			},
		},
	}
	app.Run(os.Args)
}
