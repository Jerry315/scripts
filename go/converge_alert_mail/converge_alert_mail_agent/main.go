package main

import (
	"dev/converge_alert_mail/converge_alert_mail_agent/common"
	"dev/converge_alert_mail/converge_alert_mail_agent/request"
	"fmt"
	"os"
)

func main() {
	conf := common.GetConf()
	logger := common.InitLogger()
	token, err := request.GetToken(conf.Server, conf.SecretId, conf.SecretKey)
	if err != nil {
		logger.Error(fmt.Sprintf("get token failed. %v", err))
		os.Exit(0)
	}
	cycleData, err := request.GetCycleDevice(conf, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("get cycle data failed. %v", err))
		os.Exit(0)
	}
	timeOutData, err := request.GetTimeOutDevice(conf, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("get timeout data failed. %v", err))
		os.Exit(0)
	}
	cycleUrl := conf.Server + fmt.Sprintf("/ops/alert/v1/device_cycle?token=%s", token.Token)
	timeOutUrl := conf.Server + fmt.Sprintf("/ops/alert/v1/device_timeout?token=%s", token.Token)
	err = request.UpData(cycleUrl, cycleData)
	if err != nil {
		logger.Error(fmt.Sprintf("存储周期一致性校验数据上传失败"))
	}
	err = request.UpData(timeOutUrl, timeOutData)
	if err != nil {
		logger.Error(fmt.Sprintf("cid超时检测数据上传失败"))
	}

}
