package main

import (
	"dev/monitor_oss_interface/common"
	"dev/monitor_oss_interface/request"
	"dev/monitor_oss_interface/stream"
	"fmt"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"time"
)

func imgUpload(baseDir string, conf common.Config, logger *zap.Logger) (flag bool) {
	flag = true
	rand.Seed(time.Now().Unix())
	tokenMap := make(map[int]string)
	currentTime := time.Now().Unix()
	records := []common.Record{}
	imgDir := path.Join(baseDir, "img")
	imgs := common.ListFile(imgDir)
	recordFile := path.Join(baseDir, conf.Log.Path, "records")
	_, err := os.Stat(recordFile)
	if err == nil {
		records = common.CheckHistory(recordFile)
		records = common.CleanExpireRecord(records, currentTime)
	}

	for _, item := range conf.Picture {
		token, ok := tokenMap[item.Cid]
		if !ok {
			token = common.GetToken(item.Cid, item.Cycle, conf.AppKey)
			tokenMap[item.Cid] = token
		}
		img := path.Join(imgDir, imgs[rand.Intn(len(imgs))])
		imgObj, err := os.Stat(img)
		if err != nil {
			logger.Error(fmt.Sprintf("imgUpLoad [imgUpLoad] upload file is not exist, error %v.", err))
			flag = false
			continue
		}
		size := imgObj.Size()
		response, err := request.PicUpload(int(size), item.Cid, item.Cycle, conf.Retry, conf.Timeout, conf.OssUrl, token, img, logger)
		if err != nil {
			logger.Error(fmt.Sprintf("imgUpLoad [imgUpLoad] upload file is failed, error %v.", err))
			flag = false
			continue
		}
		record := common.Record{}
		record.Media = 0
		record.Cid = item.Cid
		record.Cycle = item.Cycle
		record.ObjectId = response.Obj_id
		record.Size = int(size)
		record.Timestamp = int(currentTime)
		records = append(records, record)
	}
	common.WriteFile(records, path.Join(baseDir, conf.Log.Path, "records"), logger)
	return
}

func imgDownLoad(baseDir string, conf common.Config, logger *zap.Logger) (flag bool) {
	flag = true
	tokenMap := make(map[int]string)
	currentTime := time.Now().Unix()
	recordFile := path.Join(baseDir, conf.Log.Path, "records")
	records := []common.Record{}
	_, err := os.Stat(recordFile)
	if err != nil {
		logger.Error(fmt.Sprintf("imgDownLoad [imgDownLoad] record file is not exists. error %v.", err))
		flag = false
		return
	} else {
		records = common.CheckHistory(recordFile)
		records = common.CleanExpireRecord(records, currentTime)
	}
	//循环请求所有的记录，全部都返回true，结果t1为true，否则结果t1为false
	t1 := true
	for _, record := range records {
		if record.Media != 0 {
			continue
		}
		token, ok := tokenMap[record.Cid]
		if !ok {
			token = common.GetToken(record.Cid, record.Cycle, conf.AppKey)
			tokenMap[record.Cid] = token
		}
		t2 := request.PicDownload(record.Cid, record.Size, record.Timestamp, conf.Retry, conf.Timeout, token, record.ObjectId, conf.OssUrl, logger)
		if !t2 {
			t1 = false
		}
	}
	if !t1 {
		flag = false
	}
	return
}

func videoUpload(baseDir string, conf common.Config, logger *zap.Logger) (flag bool) {
	flag = true
	rand.Seed(time.Now().Unix())
	tokenMap := make(map[int]string)
	videoDir := path.Join(baseDir, "video")
	videos := common.ListFile(videoDir)
	for _, item := range conf.Video {
		token, ok := tokenMap[item.Cid]
		if !ok {
			token = common.GetToken(item.Cid, item.Cycle, conf.AppKey)
			tokenMap[item.Cid] = token
		}
		err := request.UpdateDevice(item.Cid, item.Cycle, conf.Timeout, conf.ApiUrl, conf.AppKey, logger)
		if err != nil {
			flag = false
		}
		video := path.Join(videoDir, videos[rand.Intn(len(videos))])
		_, err = os.Stat(video)
		if err != nil {
			logger.Error(fmt.Sprintf("videoUpload [videoUpload] upload file is not exist, error %v. cid: %d.", err, item.Cid))
			flag = false
			continue
		}
		rtmpUrl := conf.RtmpUrl + fmt.Sprintf("/%s", token)
		err = stream.PushStream(video, rtmpUrl, conf.Timeout, logger)
		if err != nil {
			logger.Error(fmt.Sprintf("videoUpload [videoUpload] push rtmp stream failed, error %v. cid: %d.", err, item.Cid))
			flag = false
			continue
		}

	}
	return
}

func videoDownload(baseDir string, conf common.Config, logger *zap.Logger) (flag bool) {
	flag = true
	rand.Seed(time.Now().Unix())
	tokenMap := make(map[int]string)
	currentTime := time.Now()
	end := currentTime.Unix() - int64(currentTime.Second()) - int64((60 * currentTime.Minute())) - int64(3600*currentTime.Hour()) - 1
	begin := end - 86400 + 1
	records := []common.Record{}
	recordFile := path.Join(baseDir, conf.Log.Path, "records")
	_, err := os.Stat(recordFile)
	if err == nil {
		records = common.CheckHistory(recordFile)
		records = common.CleanExpireRecord(records, currentTime.Unix())
	}
	for _, item := range conf.Video {
		token, ok := tokenMap[item.Cid]
		if !ok {
			token = common.GetToken(item.Cid, item.Cycle, conf.AppKey)
			tokenMap[item.Cid] = token
		}
		objId, err := request.GetVideoTimer(item.Cid, int(begin), int(end), conf.Retry, conf.Timeout, token, conf.OssUrl, logger)
		if err != nil {
			flag = false
			continue
		}
		t := true
		if len(records) > 0 {
			for _, record := range records {
				if record.ObjectId == objId {
					t = false
					break
				}
			}
		}
		if t {
			record := common.Record{}
			record.Media = 1
			record.Cid = item.Cid
			record.Cycle = item.Cycle
			record.ObjectId = objId
			record.Size = 0
			record.Timestamp = int(begin)
			records = append(records, record)
		}
	}
	common.WriteFile(records, path.Join(baseDir, conf.Log.Path, "records"), logger)
	t1 := true
	for _, record := range records {
		if record.Media != 1 {
			continue
		}
		token, ok := tokenMap[record.Cid]
		if !ok {
			token = common.GetToken(record.Cid, record.Cycle, conf.AppKey)
			tokenMap[record.Cid] = token
		}
		//fmt.Println("cid: ", record.Cid, "object id: ", record.ObjectId, "timestamp: ", record.Timestamp, "retry: ", conf.Retry)
		t2 := request.VideoDownTs(record.Cid, record.Timestamp, conf.Retry, conf.Timeout, record.ObjectId, token, conf.OssUrl, logger)
		if !t2 {
			t1 = false
		}
	}
	if !t1 {
		flag = false
	}
	return
}

func main() {
	config := common.GetConf()
	var logFile string
	basePath, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	if config.Log.Path == "" {
		logFile = path.Join(basePath, config.Log.FileName)
	} else {
		logFile = path.Join(config.Log.Path, config.Log.FileName)
	}
	logger := common.InitLogger(logFile, config.Log.Level)
	app := cli.NewApp()
	app.Version = common.Version
	app.Name = "check img or video upload and download chain is health."
	app.Commands = []cli.Command{
		{
			Name:        "imgUpload",
			Usage:       "图片上传",
			Description: "图片上传",
			Action: func(c *cli.Context) {
				result := imgUpload(basePath, config, logger)
				if result {
					fmt.Println(0)
				} else {
					fmt.Println(1)
				}
			},
		},
		{
			Name:        "imgDownload",
			Usage:       "图片下载",
			Description: "图片下载",
			Action: func(c *cli.Context) {
				result := imgDownLoad(basePath, config, logger)
				if result {
					fmt.Println(0)
				} else {
					fmt.Println(1)
				}
			},
		},
		{
			Name:        "videoUpload",
			Usage:       "视频上传",
			Description: "视频上传",
			Action: func(c *cli.Context) {
				result := videoUpload(basePath, config, logger)
				if result {
					fmt.Println(0)
				} else {
					fmt.Println(1)
				}
			},
		},
		{
			Name:        "videoDownload",
			Usage:       "视频下载",
			Description: "视频下载",
			Action: func(c *cli.Context) {
				result := videoDownload(basePath, config, logger)
				if result {
					fmt.Println(0)
				} else {
					fmt.Println(1)
				}
			},
		},
	}
	app.Run(os.Args)
}
