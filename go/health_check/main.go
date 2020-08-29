package main

import (
	"dev/health_check/common"
	"dev/health_check/stream"
	"fmt"
	"time"

	//"time"
)

func main() {
	conf := common.GetConf()
	logger := common.InitLogger()
	defer func() {
		if err := recover();err != nil{
			logger.Error(fmt.Sprintf("%#v",err))
		}
	}()
	n := 1
	cids := stream.GetCid(conf,conf.Urls.Relay,conf.Httpbasicauth.Stat.Username,conf.Httpbasicauth.Stat.Password,logger)
	if cids[0] == 0 && cids[1] == 0 {
		n = 1
	}else {
		tokenInfo := stream.GetToken(cids,conf.Urls.Api,conf.Httpbasicauth.Api.Username,conf.Httpbasicauth.Api.Password,logger)
		if len(tokenInfo) == 0{
			time.Sleep(2*time.Second)
			tokenInfo = stream.GetToken(cids,conf.Urls.Api,conf.Httpbasicauth.Api.Username,conf.Httpbasicauth.Api.Password,logger)
		}
		if tokenInfo != nil{
			n = stream.GetRecord(tokenInfo,conf.Urls.Oss,logger)
		}
	}
	fmt.Println(n)
}
