package request

import (
	"bytes"
	"dev/storage_cycle_log/common"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
)

func GetResponse(url string, cids []int64, logger *zap.Logger, mode, cycle int) (reponse common.Response, err error) {
	/*
	根据传入cid的mode和周期返回对应的数据
	mode: 0: 获取图片最新存储周期 1：获取视频最新存储周期
	cycle: 0 不存, 1 7,2 30,3 90,4 15,5 60,6 180,7 365,15 99999
	发送cid列表使用form-data方法，这里需要注意
	*/
	tmp := make(map[string][]int64)
	tmp["cidlist"] = cids
	query, err := json.Marshal(tmp)
	if err != nil {
		logger.Error(fmt.Sprintf("[CheckCid] create cid list string failed. %v", err))
		return
	}
	var (
		dataContentType string
		messageBuffer   = &bytes.Buffer{}
		buffLen         int
	)
	bodyBuff := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuff)
	_, err = bodyWriter.CreateFormField("cidlist")
	if err != nil {
		logger.Error(fmt.Sprintf("[CheckCid] create form-data failed. %v", err))
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
	var ioReaderSum io.Reader
	ioReaderSum = io.MultiReader(headerBuffer, messageBuffer, tailBuff)

	curl := url + fmt.Sprintf("/stat/check_cids_cycle?checkmode=%d&check_cycle=%d", mode, cycle)
	req, err := http.Post(curl, dataContentType, ioReaderSum)
	if err != nil {
		logger.Error(fmt.Sprintf("[CheckCid] create request handle failed. %v", err))
		return
	}
	defer req.Body.Close()
	respBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("[CheckCid] get request body failed. %v", err))
		return
	}
	if req.StatusCode == 200 {
		err = json.Unmarshal(respBody, &reponse)
		if err != nil {
			logger.Error(fmt.Sprintf("[CheckCid] parse response body failed. %v", err))
			return
		}
	} else {
		logger.Error(fmt.Sprintf("[CheckCid] response is err. status code %d, msg %s.", req.StatusCode, req.Status))
		err = common.NewError(req.StatusCode, "获取对象存储cid存储周期异常")
	}
	return
}

func CheckCid(url string, cids []int64, logger *zap.Logger, mode, cycle, step int) (response common.Response, err error) {
	if len(cids) > step {
		z := int(len(cids) / step)
		m := math.Mod(float64(len(cids)), float64(step))
		if m > 0 {
			z++
		}
		for i := 0; i < z; i++ {
			start := i * step
			end := (i + 1) * step
			if end > len(cids) {
				end = len(cids)
			}
			newCids := cids[start:end]
			resp, err := GetResponse(url, newCids, logger, mode, cycle)
			if err == nil {
				for _, cid := range resp.Cids {
					response.Cids = append(response.Cids, cid)
				}
			}
		}
	} else {
		response, err = GetResponse(url, cids, logger, mode, cycle)
	}
	return
}
