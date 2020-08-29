package main

import (
	"dev/storage_cycle_log/common"
	"dev/storage_cycle_log/handle"
	"dev/storage_cycle_log/request"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"
)



func main() {
	cycleMap := make(map[int64]int64)
	cycleMap[1] = 7
	cycleMap[2] = 30
	cycleMap[3] = 90
	cycleMap[4] = 15
	cycleMap[5] = 60
	cycleMap[6] = 180
	cycleMap[7] = 365
	cycleMap[15] = 99999
	conf := common.GetConf()
	logger := common.InitLogger()
	logDir, _ := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if conf.Log.Path != ""{
		logDir = path.Join(logDir,conf.Log.Path)
	}
	common.CleanFile(logDir,conf.Log.Cycle)
	mongoClient, err := handle.MongoClient(conf.Mongodb.Mawar.Url)
	if err != nil {
		logger.Error("mongodb connection failed")
	}
	currentTime := time.Now().Unix()
	mdocs := handle.MawarQuery(mongoClient, logger, conf.Mongodb.Mawar.Fields, conf.Mongodb.Mawar.Db, conf.Mongodb.Mawar.Table)
	var cidList []int64
	var CidInfoList []*common.CidInfo
	for _, doc := range mdocs {
		cidList = append(cidList, doc.CID)
	}
	resp0, err := request.CheckCid(conf.Url.CheckServer, cidList, logger, 0, 1000,conf.Step)
	if err != nil {
		logger.Error(fmt.Sprintf("获取图片存储周期失败. %v", err))
	}
	resp1, err := request.CheckCid(conf.Url.CheckServer, cidList, logger, 1, 1000,conf.Step)
	if err != nil {
		logger.Error(fmt.Sprintf("获取视频存储周期失败. %v", err))
	}
	for _, r0 := range resp0.Cids {
		rcycle := cycleMap[r0.Cycle]
		if r0.Cycle == -1 {
			rcycle = -1
		}
		if len(CidInfoList) > 0 {
			flag := false
			for _, cidInfo := range CidInfoList {
				if cidInfo.Cycle == rcycle && cidInfo.Media == "picture" {
					cidInfo.Count++
					flag = true
					break
				}
			}
			if !flag {
				tmp := common.CidInfo{
					Media: "picture",
					Cycle: rcycle,
					Count: 1,
					Time:  currentTime,
				}
				CidInfoList = append(CidInfoList, &tmp)
			}
		} else {
			tmp := common.CidInfo{
				Media: "picture",
				Cycle: rcycle,
				Count: 1,
				Time:  currentTime,
			}
			CidInfoList = append(CidInfoList, &tmp)
		}

	}
	for _, r1 := range resp1.Cids {
		rcycle := cycleMap[r1.Cycle]
		if r1.Cycle == -1 {
			rcycle = -1
		}
		if len(CidInfoList) > 0 {
			flag := false
			for _, cidInfo := range CidInfoList {
				if cidInfo.Cycle == rcycle && cidInfo.Media == "video" {
					cidInfo.Count++
					flag = true
					break
				}
			}
			if !flag {
				tmp := common.CidInfo{
					Media: "video",
					Cycle: rcycle,
					Count: 1,
					Time:  currentTime,
				}
				CidInfoList = append(CidInfoList, &tmp)
			}
		} else {
			tmp := common.CidInfo{
				Media: "video",
				Cycle: rcycle,
				Count: 1,
				Time:  currentTime,
			}
			CidInfoList = append(CidInfoList, &tmp)
		}

	}
	for _, cidInfo := range CidInfoList {
		st := fmt.Sprintf("{\"time\": %d, \"media\": \"%s\", \"cycle\": %d, \"count\": %d, \"category\": \"media_cycle\", \"datasource\": \"oss\"}\n",cidInfo.Time,cidInfo.Media,cidInfo.Cycle,cidInfo.Count)
		baseDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		monitorFile := path.Join(baseDir,"log",conf.MonitorFile + "-" + time.Now().Format("20060102")+".log")
		common.TraceFile(st, monitorFile)
	}
}
