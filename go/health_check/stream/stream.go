package stream

import (
	"dev/health_check/common"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

func GetCid(conf common.Yaml, url, username, password string, logger *zap.Logger) [2]int {
	client := &http.Client{}
	var c [2]int
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("[GetCid] something wrong with create request. %#v", err))
		return [2]int{0, 0}
	}
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("[GetCid] something wrong with get request. #%v", err))
		return [2]int{0, 0}
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("[GetCid] parse request body failed #%v", err))
			return [2]int{0, 0}
		}
		streams := doc.Find("live stream")

		var data []map[string]string

		streams.Each(func(i int, selection *goquery.Selection) {
			bw_in, _ := strconv.Atoi(selection.ChildrenFiltered("bw_in").Text())
			cost_time, _ := strconv.Atoi(selection.ChildrenFiltered("time").Text())
			if bw_in/1024 > 100 && cost_time/60000 > 2 {
				tag := true
				cid, _ := strconv.Atoi(selection.ChildrenFiltered("name").Text())
				if conf.Ignorecid != nil {
					for _, c := range conf.Ignorecid {
						if cid == c {
							tag = false
							break
						}
					}
				}
				if tag != false {
					cidIfo := make(map[string]string)
					cidIfo["cid"] = selection.ChildrenFiltered("name").Text()
					cidIfo["bw_in"] = selection.ChildrenFiltered("bw_in").Text()
					cidIfo["cost_time"] = selection.ChildrenFiltered("time").Text()
					data = append(data, cidIfo)
				}

			}
		})

		f1 := rand.Intn(len(data))
		f2 := rand.Intn(len(data))
		for i, item := range data {
			if i == f1 || i == f2 {
				cid, _ := strconv.Atoi(item["cid"])
				if c[0] == 0 {
					c[0] = cid
				} else {
					c[1] = cid
				}
			}
		}
		logger.Info("[GetCid] get cid success")
	} else {
		logger.Error(fmt.Sprintf("[GetCid] request failed"), zap.Int("StatusCode", resp.StatusCode))
	}

	return c
}

func GetToken(cids [2]int, url, username, password string, logger *zap.Logger) (tokens []map[int]string) {
	for _, cid := range cids {
		tokenInfo := common.Tokens{}
		tokenUrl := url + "/internal/devices/token?cid=" + strconv.Itoa(cid)
		req, err := http.NewRequest("GET", tokenUrl, nil)
		if err != nil {
			logger.Error(fmt.Sprintf("[GetToken] get token failed #%v", err))
			return
		}
		req.Header.Add("Content-Type", "application/json")
		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}
		r1, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.Error(fmt.Sprintf("[GetToken] get token failed #%v", err))
			return
		}
		defer r1.Body.Close()
		if r1.StatusCode == 200 {
			token := make(map[int]string)
			body, _ := ioutil.ReadAll(r1.Body)
			json.Unmarshal(body, &tokenInfo)
			token[cid] = tokenInfo.Token
			tokens = append(tokens,token)
			logger.Info("[GetToken] get token success", zap.Int("cid", cid))
		} else {
			logger.Error(fmt.Sprintf("[GetToken] request failed"), zap.Int("cid", cid))
			return
		}

	}
	return
}

func GetRecord(tokens []map[int]string, url string, logger *zap.Logger) (n int) {
	//检测cid 的时间轴，同比当前时间相差是否大于1分钟
	defer func() {
		if err := recover(); err != nil {
			logger.Error(fmt.Sprintf("#%v", err))
		}
	}()
	n = 1
	timeLines := common.TimeLines{}
	end := int(time.Now().Unix())
	begin := end - 3600
	for _, t := range tokens {
		for cid, token := range t {
			recordUrl := url + "/oss/v1/" + strconv.Itoa(cid) + "/record/timeline?begin=" + strconv.Itoa(begin) + "&end=" + strconv.Itoa(end) + "&client_token=" + token
			req, err := http.NewRequest("GET", recordUrl, nil)
			if err != nil {
				logger.Error(fmt.Sprintf("[GetRecord] create request %s failed #%v", recordUrl, err))
				n = 1
				continue
			}
			req.Header.Add("Content-Type", "application/json")
			r2, err := http.DefaultClient.Do(req)
			if err != nil {
				logger.Error(fmt.Sprintf("[GetRecord] get request body failed #%v", err), zap.Int("cid", cid))
				n = 1
				continue
			}
			defer r2.Body.Close()

			if r2.StatusCode == 200 {
				body, err := ioutil.ReadAll(r2.Body)
				if err != nil {
					logger.Error(fmt.Sprintf("[GetRecord] read request body failed #%v", err), zap.Int("cid", cid))
					n = 1
					continue
				}
				err = json.Unmarshal(body, &timeLines)
				if err != nil {
					logger.Error(fmt.Sprintf("[GetRecord] parse request body failed #%v", err), zap.Int("cid", cid))
					n = 1
					continue
				}
				if timeLines.Timelines == nil {
					logger.Error(fmt.Sprintf("[GetRecord] record time shaft is empty"), zap.Int("cid", cid))
					n = 1
					continue
				}
				periodTime := timeLines.Timelines[len(timeLines.Timelines)-1].End
				if int(end)-periodTime >= 120 {
					logger.Error("[GetRecord] The current time is more than two minute longer than the video record last end time", zap.Int("cid", cid))
					n = 1
				} else {
					if GetM3u8(token, url, cid, begin, end, 2, logger) {
						n = 0
					} else {
						n = 1
					}
				}
			} else {
				logger.Error(fmt.Sprintf("[GetRecord] requests failed #%v", err), zap.Int("StatusCode", r2.StatusCode), zap.Int("cid", cid))
				n = 1
				continue
			}

		}
	}
	return
}

func GetM3u8(token, url string, cid, begin, end, retry int, logger *zap.Logger) bool {
	//获取m3u8文件
	tag := false
	defer func() {
		if err := recover(); err != nil {
			logger.Error(fmt.Sprintf("#%v", err))
		}
	}()
	retry--
	if retry < 1 {
		return tag
	}

	m3u8_url := url + "/oss/v1/" + strconv.Itoa(cid) + "/record/m3u8/" + strconv.Itoa(begin) + "_" + strconv.Itoa(end) + ".m3u8?client_token=" + token
	req, err := http.NewRequest("GET", m3u8_url, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("[GetM3u8] create request failed #%v", err), zap.Int("cid", cid))
		return tag
	}
	req.Header.Add("Content-Type", "application/json")
	r3, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("[GetM3u8] get request body failed #%v", err), zap.Int("cid", cid))
		return tag
	}
	defer r3.Body.Close()
	if r3.StatusCode == 200 {
		body, _ := ioutil.ReadAll(r3.Body)
		b, err := regexp.Match("record/ts", body)
		if err != nil {
			logger.Error(fmt.Sprintf("[GetM3u8] get record ts none #%v", err), zap.Int("cid", cid))
			tag = GetM3u8(token, url, cid, begin, end, retry, logger)
		}
		if b {
			tag = true
			logger.Info("[GetM3u8] get record ts success", zap.Int("cid", cid))
		}
	} else {
		tag = false
		logger.Error(fmt.Sprintf("[GetM3u8] requests failed"), zap.Int("StatusCode", r3.StatusCode), zap.Int("cid", cid))
	}
	return tag
}
