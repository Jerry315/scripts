package request

import (
	"bytes"
	"dev/converge_alert_mail/converge_alert_mail_agent/common"
	"dev/converge_alert_mail/converge_alert_mail_agent/handle"
	"encoding/json"
	"fmt"
	"github.com/beevik/etree"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var config common.Config = common.GetConf()
var logger *zap.Logger = common.InitLogger()

func GetToken(url, secretid, secretkey string) (token common.Token, err error) {
	// 获取token
	tokenUrl := url + "/ops/alert/v1/login"
	payload := strings.NewReader(fmt.Sprintf("{\n\t\"secretid\": \"%s\",\n\t\"secretkey\": \"%s\"\n}", secretid, secretkey))
	req, err := http.NewRequest("POST", tokenUrl, payload)
	if err != nil {
		logger.Error(fmt.Sprintf("[GetToken] create post request failed. %v", err))
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("[GetToken] send post request failed. %v", err))
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		logger.Error(fmt.Sprintf("[GetToken] json parse requst body failed. %v", err))
	}
	return
}

func UpData(postUrl string, data interface{}) (err error) {
	// 更新数据
	dataStr, err := json.Marshal(data)
	if err != nil {
		logger.Error(fmt.Sprintf("[UpData] jsom parse data failed. %v", err))
	}
	payload := strings.NewReader(string(dataStr))
	req, err := http.NewRequest("POST", postUrl, payload)
	if err != nil {
		logger.Error(fmt.Sprintf("[UpData] create post request failed. %v", err))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("[UpData] send post request failed. %v", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		err = nil
	} else {
		err = common.NewError(400, "请求不合法")
	}
	return
}

func GetRelayCid(conf common.Config, logger *zap.Logger) (cids []int64) {
	//通过请求转发的/stat接口获取数据，对数据处理，过滤出推流(bw_in > 100kb)并且
	//推流时间(time > 2min 且 time <= 10min)的 cid 列表和推流(bw_in > 100kb)并且
	//推流时间(time > 10min)的 cid 列表
	var username = conf.Relay.Username
	var password = conf.Relay.Password
	var timeout = conf.Limit.Timeout
	var whiteList = conf.Whitelist.Relay
	for _, relayUrl := range conf.Relay.Urls {
		httpClient := &http.Client{}
		req, err := http.NewRequest("GET", relayUrl, nil)
		if err != nil {
			logger.Error(fmt.Sprintf("[GetCid] something wrong with create request. relay url: %s %v", relayUrl, err))
			continue
		}
		//如果转发开启了HTTPbaseAuth，则需要使用到用户名和密码
		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			logger.Error(fmt.Sprintf("[GetCid] something wrong with get request. relay url: %s %v", relayUrl, err))
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Error(fmt.Sprintf("[GetCid] parse request body failed. relay url: %s %v", relayUrl, err))
				continue
			}
			doc := etree.NewDocument()
			doc.ReadFromBytes(respBody)
			rtmp := doc.SelectElement("rtmp")
			server := rtmp.SelectElement("server")
			applications := server.SelectElements("application")
			for _, application := range applications {
				name := application.SelectElement("name")
				if name.Text() != "live" {
					continue
				}
				live := application.SelectElement("live")
				streams := live.SelectElements("stream")
				for _, stream := range streams {
					name := stream.SelectElement("name")
					bwIn := stream.SelectElement("bw_in")
					onLineTime := stream.SelectElement("time")
					bw, _ := strconv.Atoi(bwIn.Text())
					costTime, _ := strconv.Atoi(onLineTime.Text())
					if bw/1024 > 100 && costTime/60000 >= timeout {
						relayCid, _ := strconv.Atoi(name.Text())
						flag := true
						if len(whiteList) > 0 {
							for _, cid := range whiteList {
								if cid == int64(relayCid) {
									flag = false
									break
								}
							}
						}
						if flag {
							cids = append(cids, int64(relayCid))
						}
					}
				}
			}
		}
	}
	return
}

func PostFormData(cids []int64, url string, logger *zap.Logger) (resp *http.Response, err error) {
	//通过form-data方式发送请求
	tmp := make(map[string][]int64)
	tmp["cidlist"] = cids
	query, err := json.Marshal(tmp)
	var (
		dataContentType string
		messageBuffer   = &bytes.Buffer{}
		buffLen         int
		ioReaderSum     io.Reader
	)
	bodyBuff := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuff)
	_, err = bodyWriter.CreateFormField("cidlist")
	if err != nil {
		logger.Error(fmt.Sprintf("[PostFormData] create form-data failed. %v", err))
		return
	}
	buffLen = len(bodyBuff.Bytes())
	headerBuffer := &bytes.Buffer{}

	headerBuffer.Write(bodyBuff.Bytes()[0:buffLen])
	bodyWriter.Close()
	messageBuffer.Write(query)
	tailBuff := &bytes.Buffer{}
	tailBuff.Write(bodyBuff.Bytes()[buffLen:])
	dataContentType = bodyWriter.FormDataContentType()
	ioReaderSum = io.MultiReader(headerBuffer, messageBuffer, tailBuff)
	resp, err = http.Post(url, dataContentType, ioReaderSum)
	if err != nil {
		logger.Error(fmt.Sprintf("[PostFormData] create request handle failed. %v", err))
		return
	}
	if resp.StatusCode != 200 {
		fmt.Printf("[PostFormData] 非法请求. 请求状态码%d\n", resp.StatusCode)
		err = common.NewError(400, "参数不合法")
		return
	}
	return
}

func SplitCids(cids []int64, step int) (cidIndex [][]int) {
	if len(cids) > step {
		z := int(len(cids) / step)
		m := math.Mod(float64(len(cids)), float64(step))
		if m > 0 {
			z++
		}
		for i := 0; i < z; i++ {
			var tmp []int
			start := i * step
			end := (i + 1) * step
			if end > len(cids) {
				end = len(cids)
			}
			tmp = append(tmp, start, end)
			cidIndex = append(cidIndex, tmp)
		}
	} else {
		tmp := []int{0, len(cids)}
		cidIndex = append(cidIndex, tmp)
	}
	return
}

func GetCycleDevice(conf common.Config, logger *zap.Logger) (response common.CycleResponse, err error) {
	// 将中央数据库和通配数据取交集，过滤白名单中的cid，返回视频和图片对应的cid
	// 通配数据库客户端
	defer func() {
		err := recover()
		if err != nil {
			logger.Error(fmt.Sprintf("[GetCycleDevice] %v", err))
		}
	}()
	response.Project = config.Project
	response.Module = "deviceCycle"
	response.CreateTime = time.Now().Unix()
	response.Zname = config.Zname
	response.Status = false
	DevicesMongoClient, err := handle.MongoClient(conf.Mongodb.Devices.Url)
	if err != nil {
		response.Msg = "连接通配device数据库失败"
		logger.Error(fmt.Sprintf("[GetCycleDevice] connect to mawar device db failed. %v", err))
		return
	}
	MawarAppMongoClient, err := handle.MongoClient(conf.Mongodb.MawarApp.Url)
	if err != nil {
		response.Msg = "连接通配app数据库失败"
		logger.Error(fmt.Sprintf("[GetCycleDevice] connect to mawar app db failed. %v", err))
		return
	}
	// 中央数据库客户端
	CameraMongoClient, err := handle.MongoClient(conf.Mongodb.Camera.Url)
	if err != nil {
		response.Msg = "连接中央数据库失败"
		logger.Error(fmt.Sprintf("[GetCycleDevice] connect to camera device db failed. %v", err))
		return
	}
	//两个数据库取交集之后的cid信息
	var cycleMap = common.CycleMap
	var qdocs []common.MawarDoc
	var pcids []int64
	var vcids []  int64
	var uncids []common.UnCid
	//从通配中获取全部cid
	ddcs := handle.DevicesQuery(DevicesMongoClient, logger, conf.Mongodb.Devices.Fields, conf.Mongodb.Devices.Db, conf.Mongodb.Devices.Table)
	macs := handle.MawarAppQuery(MawarAppMongoClient, logger, conf.Mongodb.MawarApp.Fields, conf.Mongodb.MawarApp.Db, conf.Mongodb.MawarApp.Table)
	var mdocs []common.MawarDoc
	for _, ddc := range ddcs {
		var tmp common.MawarDoc
		for _, mac := range macs {
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
	cdocs := handle.CameraQuery(CameraMongoClient, logger, conf.Limit.MessageTimestamp, conf.Mongodb.Camera.Fields, conf.Mongodb.Camera.Db, conf.Mongodb.Camera.Table)
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
		logger.Warn("[GetCycleDevice] from mawar database all cid are not mach camera database cids")
		response.Status = true
		response.Msg = "没有可校验的cid"
		return
	}
	//从对象存储中获取指定cid的图片存储周期信息
	resp0, err := CheckDeviceCycle(conf.CheckServer, pcids, logger, 0, 1000, conf.Limit.Step)
	if err != nil {
		response.Msg = fmt.Sprintf("[GetCycleDevice] 获取图片存储周期异常。%v", err)
		logger.Error(fmt.Sprintf("[GetCycleDevice] 获取图片存储周期异常。%v", err))
	}

	//从对象存储中获取指定cid的视频存储周期信息
	resp1, err := CheckDeviceCycle(conf.CheckServer, vcids, logger, 1, 1000, conf.Limit.Step)
	if err != nil {
		response.Msg = fmt.Sprintf("[GetCycleDevice] 获取视频存储周期异常。%v", err)
		logger.Error(fmt.Sprintf("[GetCycleDevice] 获取视频存储周期异常。%v", err))
	}

	//校验cid的图片存储周期
	for _, resp := range resp0.Cids {
		for _, doc := range qdocs {
			if resp.CID == doc.CID {
				PicStorage, _ := strconv.Atoi(doc.PicStorage)
				cycle := cycleMap[resp.Cycle]
				if cycle != int64(PicStorage) {
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
	response.Data = uncids
	response.Status = true
	todayDate := time.Now().Format(conf.Log.Layout)
	name := conf.Log.CycleFile + "-" + todayDate + ".txt"
	fileName := path.Join(common.BaseDir, conf.Log.Path, name)
	if _, err := os.Stat(fileName); err != nil {
		os.Mkdir(path.Join(common.BaseDir, conf.Log.Path), 0644)
		os.Create(fileName)
	}
	common.RecordCid(uncids, fileName, logger)
	common.ClearExpireData(conf)
	return
}

func GetDeviceCycleResponse(url string, cids []int64, logger *zap.Logger, mode, cycle int) (response common.Response, err error) {
	/*
	根据传入cid的mode和周期返回对应的数据
	mode: 0: 获取图片最新存储周期 1：获取视频最新存储周期
	cycle: 0 不存, 1 7,2 30,3 90,4 15,5 60,6 180,7 365,15 99999
	发送cid列表使用from-data方法，这里需要注意
	*/
	curl := url + fmt.Sprintf("/stat/check_cids_cycle?checkmode=%d&check_cycle=%d", mode, cycle)
	req, err := PostFormData(cids, curl, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("[GetDeviceCycleResponse] create request handle failed. %v", err))
		return
	}
	defer req.Body.Close()
	if req.StatusCode == 200 {
		respBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("[GetDeviceCycleResponse] get request body failed. %v", err))
		}
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			logger.Error(fmt.Sprintf("[GetDeviceCycleResponse] parse response body failed. %v", err))
		}
	} else {
		logger.Error(fmt.Sprintf("[GetDeviceCycleResponse] response is err. status code %d, msg %s.", req.StatusCode, req.Status))
		err = common.NewError(req.StatusCode, "请求参数不合法")
	}
	return
}

func GetDeviceTimeOutResponse0(url string, cids []int64, Timeout int64, logger *zap.Logger) (response *common.CheckServerResp0, err error) {
	/*
	根据传入cid的mode和周期返回对应的数据
	mode: 0: 获取图片最新存储周期
	*/
	curl := url + fmt.Sprintf("/stat/timeout_cids?timeout=%d&checkmode=0", Timeout)
	req, err := PostFormData(cids, curl, logger)
	defer req.Body.Close()
	if err != nil {
		logger.Error(fmt.Sprintf("[GetDeviceTimeOutResponse0] create request handle failed. %v", err))
		return
	}
	if req.StatusCode == 200 {
		respBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("[GetDeviceTimeOutResponse0] get request body failed. %v", err))
		}
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			logger.Error(fmt.Sprintf("[GetDeviceTimeOutResponse0] parse response body failed. %v", err))
		}
	} else {
		logger.Error(fmt.Sprintf("[GetDeviceTimeOutResponse0] response is err. status code %d, msg %s.", req.StatusCode, req.Status))
		err = common.NewError(req.StatusCode, "请求参数不合法")
	}
	return
}

func GetDeviceTimeOutResponse1(url string, cids []int64, Timeout int64, logger *zap.Logger) (response *common.CheckServerResp1, err error) {
	/*
	根据传入cid的mode和周期返回对应的数据
	mode: 1：检测多久未推流
	*/
	curl := url + fmt.Sprintf("/stat/timeout_cids?timeout=%d&checkmode=1", Timeout)
	req, err := PostFormData(cids, curl, logger)
	defer req.Body.Close()
	if err != nil {
		logger.Error(fmt.Sprintf("[GetDeviceTimeOutResponse1] create request handle failed. %v", err))
		return
	}

	if req.StatusCode == 200 {
		respBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("[GetDeviceTimeOutResponse1] get request body failed. %v", err))
		}
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			logger.Error(fmt.Sprintf("[GetDeviceTimeOutResponse1] parse response body failed. %v", err))
		}
	} else {
		logger.Error(fmt.Sprintf("[GetDeviceTimeOutResponse1] response is err. status code %d, msg %s.", req.StatusCode, req.Status))
		err = common.NewError(req.StatusCode, "请求参数不合法")
	}
	return
}

func CheckDeviceCycle(url string, cids []int64, logger *zap.Logger, mode, cycle, step int) (response common.Response, err error) {
	defer func() {
		err := recover()
		if err != nil {
			logger.Error(fmt.Sprintf("[CheckDeviceCycle] 检测cid设备存储周期失败, %v", err))
		}
	}()
	for _, item := range SplitCids(cids, step) {
		resp, err := GetDeviceCycleResponse(url, cids[item[0]:item[1]], logger, mode, cycle)
		if err == nil {
			for _, cid := range resp.Cids {
				response.Cids = append(response.Cids, cid)
			}
		}
	}
	return
}

func CheckDeviceTimeout0(url string, cids []int64, logger *zap.Logger, step int, Timeout int64) (response common.CheckServerResp0, err error) {
	defer func() {
		err := recover()
		if err != nil {
			logger.Error(fmt.Sprintf("[CheckDeviceTimeout0] 检测cid未上传图片失败, %v", err))
		}
	}()
	for _, item := range SplitCids(cids, step) {
		resp, err := GetDeviceTimeOutResponse0(url, cids[item[0]:item[1]], Timeout, logger)
		if err != nil {
			logger.Error(fmt.Sprintf("%v", err))
		} else {
			for _, cid := range resp.Timeoutcids {
				response.Timeoutcids = append(response.Timeoutcids, cid)
			}
		}
		time.Sleep(time.Second * 10)

	}
	return
}

func CheckDeviceTimeout1(url string, cids []int64, logger *zap.Logger, step int, Timeout int64) (response common.CheckServerResp1, err error) {
	defer func() {
		err := recover()
		if err != nil {
			logger.Error(fmt.Sprintf("[CheckDeviceTimeout1] 检测cid未推流失败, %v", err))
		}
	}()
	for _, item := range SplitCids(cids, step) {
		resp, err := GetDeviceTimeOutResponse1(url, cids[item[0]:item[1]], Timeout, logger)
		if err == nil {
			for _, cid := range resp.Timeoutcids {
				response.Timeoutcids = append(response.Timeoutcids, cid)
			}
		} else {
			err = common.NewError(400, "参数不合法")
		}
		time.Sleep(time.Second * 10)
	}
	return
}

func GetTimeOutDevice(config common.Config, logger *zap.Logger) (response common.TimeOutResponse, err error) {
	defer func() {
		err := recover()
		if err != nil{
			logger.Error(fmt.Sprintf("获取cid未推流和未上传图片数据失败"))
		}
	}()
	cids := GetRelayCid(config, logger)
	response.Project = config.Project
	response.Module = "deviceTimeOut"
	response.CreateTime = time.Now().Unix()
	response.Zname = config.Zname
	response.Total = int64(len(cids))
	response.Status = false
	url := config.CheckServer
	checkServerResp0, err := CheckDeviceTimeout0(url, cids, logger, config.Limit.Step, config.Timeout)
	if err != nil {
		response.Msg = fmt.Sprintf("检测cid未推流失败. %v", err)
		logger.Error(fmt.Sprintf("[GetTimeOutDevice] check device mode 2 failed. %v", err))
	}
	checkServerResp1, err := CheckDeviceTimeout1(url, cids, logger, config.Limit.Step, config.Timeout)
	if err != nil {
		response.Msg = fmt.Sprintf("检测cid未上传图片失败. %v", err)
		logger.Error(fmt.Sprintf("[GetTimeOutDevice] check device mode 1 failed. %v", err))
	}
	currentTime := time.Now().Unix()

	if len(checkServerResp1.Timeoutcids) > 0 {
		var tmpTimeOutCids []common.Timeoutcid
		for _, item1 := range checkServerResp1.Timeoutcids {
			flag := true
			for _, item0 := range checkServerResp0.Timeoutcids {
				if item1.CID == item0.CID {
					item0.LatestImageTime = item1.LatestVideoTime
					flag = false
					break
				}
			}
			if flag {
				tmp := common.Timeoutcid{
					CID:             item1.CID,
					LatestImageTime: currentTime,
					LatestVideoTime: item1.LatestVideoTime,
				}
				tmpTimeOutCids = append(tmpTimeOutCids, tmp)
			}
		}
		for _, item := range tmpTimeOutCids {
			checkServerResp0.Timeoutcids = append(checkServerResp0.Timeoutcids, item)
		}
	}
	response.Data = checkServerResp0.Timeoutcids
	response.Status = true
	todayDate := time.Now().Format(config.Log.Layout)
	name := config.Log.TimeOutFile + "-" + todayDate + ".txt"
	fileName := path.Join(common.BaseDir, config.Log.Path, name)
	if _, err := os.Stat(fileName); err != nil {
		os.Mkdir(path.Join(common.BaseDir, config.Log.Path), 0644)
		os.Create(fileName)
	}
	common.RecordCid(checkServerResp0, fileName, logger)
	common.ClearExpireData(config)
	return
}
