package main

import (
	"dev/check_storage_cycle/common"
	"dev/check_storage_cycle/handle"
	"dev/check_storage_cycle/request"
	"dev/check_storage_cycle/stream"
	"fmt"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

func sample(conf common.Config, logger *zap.Logger, cycleMap map[int64]int64) {
	/*
	抽样检测cid的视频和图片存储周期是否匹配
	*/
	var picCids []int64
	var videoCids []int64
	uncids := make(map[string][]int64)
	flag := 0
	basePath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	if conf.Log.Path != "" {
		basePath = path.Join(basePath, conf.Log.Path)
	}
	cidFile := path.Join(basePath, "cid")
	cids := make(map[string][]int64)
	//检测上一次抽样检测是否异常，并保存异常cid，如果存在，本次则继续使用上次异常cid
	_, err := os.Stat(cidFile)
	if err == nil {
		cids = common.CheckHistory(cidFile)
		//获取异常cid后删除cid文件，避免后续继续使用该文件
		os.Remove(cidFile)
	}
	mongoClient, err := handle.MongoClient(conf.Mongodb.Devices.Url)
	if err != nil {
		logger.Error(fmt.Sprintf("数据库链接失败. %v", err))
		flag = 1
	}
	pCids := cids["picture"]
	vCids := cids["video"]
	if len(pCids) == 0 {
		pCids, err = stream.GetCid(conf, logger, mongoClient, conf.Whitelist.Picture)
		if err != nil {
			flag = 1
		}
	}
	if len(vCids) == 0 {
		vCids, err = stream.GetCid(conf, logger, mongoClient, conf.Whitelist.Video)
		if err != nil {
			flag = 1
		}
	}
	if len(pCids) > 0 {
		pDocs := handle.DevicesQuery(mongoClient, logger, pCids, conf.Mongodb.Devices.Fields, conf.Mongodb.Devices.Db, conf.Mongodb.Devices.Table)
		resp0, err := request.CheckCid(conf.Url.CheckServer, pCids, logger, 0, 1000, conf.Limit.Step)
		if err != nil {
			//记录异常cid到文件中
			common.RecordCid(cids, cidFile, logger)
			logger.Error(fmt.Sprintf("获取图片存储周期异常。%v", err))
			flag = 1
		}
		for _, resp := range resp0.Cids {
			for _, doc := range pDocs {
				if resp.CID == doc.CID {
					//从未上传过图片的设备认定为正常
					if resp.Time == 0 && resp.Cycle == -1 {
						break
					}
					PicStorage, _ := strconv.Atoi(doc.PicStorage)
					cycle := cycleMap[resp.Cycle]
					logger.Info(fmt.Sprintf("check cid %d oss picture storage cycle is %d, mawar picture storage cycle is %d", resp.CID, cycle, PicStorage))
					if cycle != int64(PicStorage) {
						picCids = append(picCids, resp.CID)
						logger.Warn(fmt.Sprintf("check cid %d picture storage cycle is not match", resp.CID))
						flag = 1
					}
					break
				}
			}
		}
	}

	if len(vCids) > 0 {
		vDocs := handle.DevicesQuery(mongoClient, logger, vCids, conf.Mongodb.Devices.Fields, conf.Mongodb.Devices.Db, conf.Mongodb.Devices.Table)
		resp1, err := request.CheckCid(conf.Url.CheckServer, vCids, logger, 1, 1000, conf.Limit.Step)
		if err != nil {
			//记录异常cid到文件中
			common.RecordCid(cids, cidFile, logger)
			logger.Error(fmt.Sprintf("获取视频存储周期异常。%v", err))
			flag = 1
		}
		for _, resp := range resp1.Cids {
			for _, doc := range vDocs {
				if resp.CID == doc.CID {
					//从未推过流的设备认定为正常
					if resp.Time == 0 && resp.Cycle == -1 {
						break
					}
					VideoStorage, _ := strconv.Atoi(doc.VideoStorage)
					cycle := cycleMap[resp.Cycle]
					logger.Info(fmt.Sprintf("check cid %d oss video storage cycle is %d, mawar video storage cycle is %d", resp.CID, cycle, VideoStorage))
					if cycle != int64(VideoStorage) {
						videoCids = append(videoCids, resp.CID)
						logger.Info(fmt.Sprintf("check cid %d video storage cycle is not match", resp.CID))
						flag = 1
					}
					break
				}
			}
		}
	}

	uncids["picture"] = picCids
	uncids["video"] = videoCids
	if len(uncids["picture"]) > 0 || len(uncids["video"]) > 0 {
		//记录异常cid到文件中
		common.RecordCid(uncids, cidFile, logger)
	}
	fmt.Println(flag)
}

func full(conf common.Config, logger *zap.Logger, cycleMap map[int64]int64) {
	//全量校验cid
	result := 0
	DevicesMongoClient, err := handle.MongoClient(conf.Mongodb.Devices.Url)
	if err != nil {
		logger.Error(fmt.Sprintf("数据库链接失败. %v", err))
		result = 1
	}
	MawarAppMongoClient, err := handle.MongoClient(conf.Mongodb.MawarApp.Url)
	if err != nil {
		logger.Error(fmt.Sprintf("数据库链接失败. %v", err))
		result = 1
	}
	CameraMongoClient, err := handle.MongoClient(conf.Mongodb.Camera.Url)
	if err != nil {
		logger.Error(fmt.Sprintf("数据库链接失败. %v", err))
		result = 1
	}
	cids := make([]int64, 0)
	//需要全量查询的cid列表
	var pcids []int64
	var vcids [] int64
	//两个数据库取交集之后的cid信息
	var qdocs []common.MawarDoc
	//从通配中获取全部cid
	deviceDocs := handle.DevicesQuery(DevicesMongoClient, logger, cids, conf.Mongodb.Devices.Fields, conf.Mongodb.Devices.Db, conf.Mongodb.Devices.Table)
	mawarAppDocs := handle.MawarAppQuery(MawarAppMongoClient, logger, cids, conf.Mongodb.MawarApp.Fields, conf.Mongodb.MawarApp.Db, conf.Mongodb.MawarApp.Table)
	// 数据合并
	var mdocs []common.MawarDoc
	for _, ddc := range deviceDocs {
		var tmp common.MawarDoc
		for _, mac := range mawarAppDocs {
			if ddc.CID == mac.CID {
				tmp.CID = ddc.CID
				tmp.Name = ddc.Name
				tmp.SoftwareBuild = ddc.SoftwareBuild
				tmp.SoftwareVersion = ddc.SoftwareVersion
				tmp.Model = ddc.Model
				tmp.Brand = ddc.Brand
				tmp.SN = ddc.SN
				tmp.Group = mac.Group
				tmp.PicStorage = ddc.PicStorage
				tmp.VideoStorage = ddc.VideoStorage
				mdocs = append(mdocs, tmp)
				break
			}
		}
	}
	//从camera表中获取push_state=4，心跳时间（message_timestamp）350内更新更新过的cid
	cdocs := handle.CameraQuery(CameraMongoClient, logger, cids, conf.Limit.Message_timestamp, conf.Mongodb.Camera.Fields, conf.Mongodb.Camera.Db, conf.Mongodb.Camera.Table)
	//取通配数据库和camera数据中cid的交集
	all := conf.Whitelist.All
	picture := conf.Whitelist.Picture
	video := conf.Whitelist.Video
	all, picture, video = common.CompareCid(all, picture, video)
	for _, cd := range cdocs {
		f := true
		if len(all) > 0 {
			for _, cid := range all {
				if cd.CID == cid {
					f = false
					break
				}
			}
		}
		if !f {
			continue
		}
		for _, md := range mdocs {
			if cd.CID == md.CID {
				p := true
				v := true
				if len(picture) > 0 {
					for _, cid := range picture {
						if cd.CID == cid {
							p = false
							break
						}
					}
				}
				if len(video) > 0 {
					for _, cid := range video {
						if cd.CID == cid {
							v = false
							break
						}
					}
				}
				qdocs = append(qdocs, md)
				if p == true {
					pcids = append(pcids, cd.CID)

				}
				if v == true {
					vcids = append(vcids, cd.CID)
				}
				break
			}
		}
	}
	//判断是否存在通配数据库中的cid全都不在camera数据库中
	if len(pcids) == 0 || len(vcids) == 0 || len(qdocs) == 0 {
		logger.Error("from mawar database all cid are not mach camera database cids")
		err := common.SendMail(conf, "全量校验cid视频和图片存储周期异常，通配和中央暂无交集cid")
		if err != nil {
			logger.Error(fmt.Sprintf("邮件服务异常，发送邮件失败。%v", err))
		}

		fmt.Println(1)
		os.Exit(1)
	}
	//从对象存储中获取指定cid的图片存储周期信息
	resp0, err := request.CheckCid(conf.Url.CheckServer, pcids, logger, 0, 1000, conf.Limit.Step)
	if err != nil {
		logger.Error(fmt.Sprintf("获取图片存储周期异常。%v", err))
		err := common.SendMail(conf, "全量校验cid图片存储周期失败")
		if err != nil {
			logger.Error(fmt.Sprintf("邮件服务异常，发送邮件失败。%v", err))
		}
		result = 1
	}

	//从对象存储中获取指定cid的视频存储周期信息
	resp1, err := request.CheckCid(conf.Url.CheckServer, vcids, logger, 1, 1000, conf.Limit.Step)
	if err != nil {
		logger.Error(fmt.Sprintf("获取视频存储周期异常。%v", err))
		err = common.SendMail(conf, "全量校验cid视频存储周期失败")
		if err != nil {
			logger.Error(fmt.Sprintf("邮件服务异常，发送邮件失败。%v", err))
		}

		result = 1
	}
	var uncids []common.UnCid
	//校验cid的图片存储周期
	for _, resp := range resp0.Cids {
		for _, doc := range qdocs {
			if resp.CID == doc.CID {
				PicStorage, _ := strconv.Atoi(doc.PicStorage)
				cycle := cycleMap[resp.Cycle]
				if cycle != int64(PicStorage) {
					if cycle != -1 {
						result = 1
					}
					logger.Warn(fmt.Sprintf("check cid %d oss picture storage cycle is %d, mawar picture storage cycle is %d", resp.CID, cycle, PicStorage))
					if len(uncids) > 0 {
						flag := false
						for _, uncid := range uncids {
							//如果cid已存在异常cid列表中只更新cid的图片存储周期，否则就新增到异常cid列表
							if uncid.CID == doc.CID {
								uncid.MPIC = int64(PicStorage)
								uncid.OPIC = cycle
								flag = true
								break
							}
						}
						if !flag {
							VideoStorage, _ := strconv.Atoi(doc.VideoStorage)
							ov := int64(VideoStorage)
							for _, r1 := range resp1.Cids {
								if r1.CID == doc.CID {
									ov = cycleMap[r1.Cycle]
								}
							}
							tmp := common.UnCid{
								CID:             doc.CID,
								MPIC:            int64(PicStorage),
								OPIC:            cycle,
								MVideo:          int64(VideoStorage),
								OVideo:          ov,
								SN:              doc.SN,
								Name:            doc.Name,
								Brand:           doc.Brand,
								Group:           doc.Group,
								Model:           doc.Model,
								SoftwareVersion: doc.SoftwareVersion,
								SoftwareBuild:   doc.SoftwareBuild,
							}
							uncids = append(uncids, tmp)
						}
					} else {
						VideoStorage, _ := strconv.Atoi(doc.VideoStorage)
						ov := int64(VideoStorage)
						for _, r1 := range resp1.Cids {
							if r1.CID == doc.CID {
								ov = cycleMap[r1.Cycle]
							}
						}
						tmp := common.UnCid{
							CID:             doc.CID,
							MPIC:            int64(PicStorage),
							OPIC:            cycle,
							MVideo:          int64(VideoStorage),
							OVideo:          ov,
							SN:              doc.SN,
							Name:            doc.Name,
							Brand:           doc.Brand,
							Group:           doc.Group,
							Model:           doc.Model,
							SoftwareVersion: doc.SoftwareVersion,
							SoftwareBuild:   doc.SoftwareBuild,
						}
						uncids = append(uncids, tmp)
					}
				}

				break
			}
		}
	}
	//对cid的视频周期进行校验
	for _, resp := range resp1.Cids {
		for _, doc := range qdocs {
			if resp.CID == doc.CID {
				VideoStorage, _ := strconv.Atoi(doc.VideoStorage)
				cycle := cycleMap[resp.Cycle]
				if cycle != int64(VideoStorage) {
					if cycle != -1 {
						result = 1
					}
					logger.Warn(fmt.Sprintf("check cid %d oss video storage cycle is %d, mawar video storage cycle is %d", resp.CID, cycle, VideoStorage))
					if len(uncids) > 0 {
						//标记cid是否异常
						flag := false
						for _, uncid := range uncids {
							//如果cid已存在异常cid列表中只更新cid的视频存储周期，否则就新增到异常cid列表
							if uncid.CID == doc.CID {
								uncid.MVideo = int64(VideoStorage)
								uncid.OVideo = cycle
								flag = true
								break
							}
						}
						if !flag {
							PicStorage, _ := strconv.Atoi(doc.PicStorage)
							op := int64(PicStorage)
							for _, r0 := range resp0.Cids {
								if r0.CID == doc.CID {
									op = cycleMap[r0.Cycle]
								}
							}
							tmp := common.UnCid{
								CID:             doc.CID,
								MPIC:            int64(PicStorage),
								OPIC:            op,
								MVideo:          int64(VideoStorage),
								OVideo:          cycle,
								SN:              doc.SN,
								Name:            doc.Name,
								Brand:           doc.Brand,
								Group:           doc.Group,
								Model:           doc.Model,
								SoftwareVersion: doc.SoftwareVersion,
								SoftwareBuild:   doc.SoftwareBuild,
							}
							uncids = append(uncids, tmp)
						}

					} else {
						PicStorage, _ := strconv.Atoi(doc.PicStorage)
						op := int64(PicStorage)
						for _, r0 := range resp0.Cids {
							if r0.CID == doc.CID {
								op = cycleMap[r0.Cycle]
							}
						}
						tmp := common.UnCid{
							CID:             doc.CID,
							MPIC:            int64(PicStorage),
							OPIC:            op,
							MVideo:          int64(VideoStorage),
							OVideo:          cycle,
							SN:              doc.SN,
							Name:            doc.Name,
							Brand:           doc.Brand,
							Group:           doc.Group,
							Model:           doc.Model,
							SoftwareVersion: doc.SoftwareVersion,
							SoftwareBuild:   doc.SoftwareBuild,
						}
						uncids = append(uncids, tmp)
					}
				}
				break
			}
		}
	}
	if len(uncids) > 0 {
		//有异常cid的情况，打印1，并发送邮件报告
		//result = 1

		logDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		if conf.Log.Path != "" {
			logDir = path.Join(logDir, conf.Log.Path)
		}
		exFile := path.Join(logDir, conf.Log.Exfile)
		handle.WriteExcel(exFile, uncids, logger)
		err := common.SendReport(conf, uncids, exFile, logger)
		if err != nil {
			logger.Error(fmt.Sprintf("发送全量校验cid视频和存储周期报告邮件失败。%v", err))
			err = common.SendMail(conf, "发送全量校验cid视频和图片存储周期报告邮件失败")
			if err != nil {
				logger.Error(fmt.Sprintf("邮件服务异常，发送邮件失败。%v", err))
			}

			result = 1
		}
	} else {
		if result == 0 {
			err = common.SendMail(conf, "全量校验cid视频和图片存储周期正常")
			if err != nil {
				logger.Error(fmt.Sprintf("邮件服务异常，发送邮件失败。%v", err))
			}

		}

	}
	fmt.Println(result)
}

func main() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(1)
		}
	}()
	//对象存储的存储周期使用的是1-9，需要映射真实周期数字
	cycleMap := make(map[int64]int64)
	cycleMap[-1] = -1
	cycleMap[0] = 0
	cycleMap[1] = 7
	cycleMap[2] = 30
	cycleMap[3] = 90
	cycleMap[4] = 15
	cycleMap[5] = 60
	cycleMap[6] = 180
	cycleMap[7] = 365
	cycleMap[15] = 99999
	logger := common.InitLogger()
	conf := common.GetConf()
	app := cli.NewApp()
	app.Version = "v0.0.6"
	app.Name = "check_storage_cycle"
	app.Commands = []cli.Command{
		{
			Name:        "full",
			Aliases:     []string{"f"},
			Usage:       "全量对比通配数据库中cid和对象存储的cid图片和视频存储周期是否一致",
			Description: "全量对比通配数据库中cid和对象存储的cid图片和视频存储周期是否一致",
			Action: func(c *cli.Context) {
				full(conf, logger, cycleMap)
			},
		},
		{
			Name:        "sample",
			Aliases:     []string{"s"},
			Usage:       "抽样对比通配数据库中cid和对象存储的cid图片和视频存储周期是否一致",
			Description: "抽样对比通配数据库中cid和对象存储的cid图片和视频存储周期是否一致",
			Action: func(c *cli.Context) {
				sample(conf, logger, cycleMap)
			},
		},
	}
	app.Run(os.Args)

}
