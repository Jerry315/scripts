package stream

import (
	"dev/check_storage_cycle/common"
	"dev/check_storage_cycle/handle"
	"fmt"
	"github.com/beevik/etree"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"net/http"
	"strconv"
)

func GetCid(conf common.Config, logger *zap.Logger, client *mgo.Session, cids []int64) (c []int64, err error) {
	/*
	通过请求转发的/stat接口获取数据，对数据处理，过滤出推流(bw_in > 100kb)并且
	推流时间(time > 2min 且 time <= 10min)的 cid 列表和推流(bw_in > 100kb)并且
	推流时间(time > 10min)的 cid 列表,从上述满⾜条件的数据中随机抽选16个cid进行
	检查(2-10min占比6成取整，大于10min占比4成取整，数量不够以实际数量为准)，
	且cid⼀一定存在通配数据库中
	*/
	var sixData []map[string]string
	var fourData []map[string]string
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", conf.Url.Relay.Url, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("[GetCid] something wrong with create request. %#v", err))
		return
	}
	//如果转发开启了HTTPbaseAuth，则需要使用到用户名和密码
	if conf.Url.Relay.Username != "" && conf.Url.Relay.Password != "" {
		req.SetBasicAuth(conf.Url.Relay.Username, conf.Url.Relay.Password)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("[GetCid] something wrong with get request. #%v", err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("[GetCid] parse request body failed #%v", err))
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
				if bw/1024 > 100 && costTime/60000 >= conf.Limit.Timeout {
					relayCid, _ := strconv.Atoi(name.Text())
					flag := true
					if len(cids) > 0 {
						for _, cid := range cids {
							if cid == int64(relayCid) {
								flag = false
								break
							}
						}
					}
					if flag {
						cidInfo := make(map[string]string)
						cidInfo["cid"] = name.Text()
						cidInfo["bw_in"] = bwIn.Text()
						cidInfo["cost_time"] = onLineTime.Text()
						if costTime/60000 > 10 {
							fourData = append(fourData, cidInfo)
						} else {
							sixData = append(sixData, cidInfo)
						}
					}
				}
			}
		}


		fields := []string{"_id"}
		cids := make([]int64, 0)
		/*
		以下操作剔除cid不在通配数据库
		*/
		ddcs := handle.DevicesQuery(client, logger, cids, fields, conf.Mongodb.Devices.Db, conf.Mongodb.Devices.Table)
		newSixData := confirmCid(sixData, ddcs)
		newFourData := confirmCid(fourData, ddcs)
		sixCidNum := common.FTI(conf.Limit.Cid_num, 0.6)
		fourCidNum := common.FTI(conf.Limit.Cid_num, 0.4)
		if len(newSixData) > sixCidNum {
			index := common.RandSlice(len(newSixData), sixCidNum)
			for _, i := range index {
				cid, _ := strconv.Atoi(newSixData[i]["cid"])
				c = append(c, int64(cid))
			}
		} else {
			for _, doc := range newSixData {
				cid, _ := strconv.Atoi(doc["cid"])
				c = append(c, int64(cid))
			}
		}
		if len(newFourData) > fourCidNum {
			index := common.RandSlice(len(newFourData), fourCidNum)
			for _, i := range index {
				cid, _ := strconv.Atoi(newFourData[i]["cid"])
				c = append(c, int64(cid))
			}
		} else {
			for _, doc := range newFourData {
				cid, _ := strconv.Atoi(doc["cid"])
				c = append(c, int64(cid))
			}
		}
		logger.Info("[GetCid] get cid success")
	} else {
		logger.Error(fmt.Sprintf("[GetCid] request failed"), zap.Int("StatusCode", resp.StatusCode))
	}

	return
}

func confirmCid(docs []map[string]string, mdocs []common.DevicesDoc) (cdocs []map[string]string) {
	/*
	比对数据，如果cid不在通配数据库中，就剔除
	*/
	for _, doc := range docs {
		for _, md := range mdocs {
			cid, _ := strconv.Atoi(doc["cid"])
			if int64(cid) == md.CID {
				cdocs = append(cdocs, doc)
				break
			}
		}
	}
	return
}
