package request

import (
	"bytes"
	"dev/monitor_oss_interface/common"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func PicUpload(size, cid, cycle, retry int, timeout time.Duration, ossUrl, token, fileName string, logger *zap.Logger) (response common.Response, err error) {
	//普通对象上传到对象存储系统
	endPrefix := fmt.Sprintf("/oss/v1/%d/uploadObject?size=%d&expiretype=%d&client_token=%s",
		cid, size, cycle, token)
	url := ossUrl + endPrefix
	//通过formdata方式上传图片
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(fileName))
	if err != nil {
		logger.Error(fmt.Sprintf("imgUpload [PicUpload] create form file failed, error %v. cid: %d.", err, cid))
		if retry > 0 {
			retry--
			response, err = PicUpload(size, cid, cycle, retry, timeout, ossUrl, token, fileName, logger)
		}
		return
	}
	file, err := os.Open(fileName)
	if err != nil {
		logger.Error(fmt.Sprintf("imgUpload [PicUpload] open file failed, error %v. cid: %d.", err, cid))
		if retry > 0 {
			retry--
			response, err = PicUpload(size, cid, cycle, retry, timeout, ossUrl, token, fileName, logger)
		}
		return
	}
	defer file.Close()

	_, err = io.Copy(part, file)
	if err != nil {
		logger.Error(fmt.Sprintf("imgUpload [PicUpload] copy file content to form failed, error %v. cid: %d.", err, cid))
		if retry > 0 {
			retry--
			response, err = PicUpload(size, cid, cycle, retry, timeout, ossUrl, token, fileName, logger)
		}
		return
	}
	err = writer.Close()
	if err != nil {
		logger.Error(fmt.Sprintf("imgUpload [PicUpload] error %v. cid: %d.", err, cid))
		if retry > 0 {
			retry--
			response, err = PicUpload(size, cid, cycle, retry, timeout, ossUrl, token, fileName, logger)
		}
		return
	}
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		logger.Error(fmt.Sprintf("imgUpload [PicUpload] create request failed, error %v. cid: %d.", err, cid))
		if retry > 0 {
			retry--
			response, err = PicUpload(size, cid, cycle, retry, timeout, ossUrl, token, fileName, logger)
		}
		return
	}
	//在请求头部添加formdata数据
	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := http.Client{Timeout: time.Second * timeout}
	//client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logger.Error(fmt.Sprintf("imgUpload [PicUpload] execute request failed, error %v. cid: %d.", err, cid))
		if retry > 0 {
			retry--
			response, err = PicUpload(size, cid, cycle, retry, timeout, ossUrl, token, fileName, logger)
		}
		return
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("imgUpload [PicUpload] reade request content failed, error %v. cid: %d.", err, cid))
		if retry > 0 {
			retry--
			response, err = PicUpload(size, cid, cycle, retry, timeout, ossUrl, token, fileName, logger)
		}
		return
	}
	if resp.StatusCode == 200 {
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			logger.Error(fmt.Sprintf("imgUpload [PicUpload] json parse request content failed, error %v. cid: %d.", err, cid))
			if retry > 0 {
				retry--
				response, err = PicUpload(size, cid, cycle, retry, timeout, ossUrl, token, fileName, logger)
			}
			return
		}
		logger.Info(fmt.Sprintf("imgUpload [PicUpload] upload file: %s success. cid: %d.", fileName, cid))

	} else {
		err = common.CustomError{}
		logger.Error(fmt.Sprintf("imgUpload [PicUpload] upload file failed, status code %d, request body %s. cid: %d.", resp.StatusCode, string(respBody), cid))
	}
	if err != nil && retry > 0 {
		retry--
		response, err = PicUpload(size, cid, cycle, retry, timeout, ossUrl, token, fileName, logger)
	}
	return
}

func PicDownload(cid, size, timestamp, retry int, timeout time.Duration, token, objId, ossUrl string, logger *zap.Logger) (flag bool) {
	url := ossUrl + fmt.Sprintf("/oss/v1/%d/objects/%s?client_token=%s", cid, objId, token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("imgDownload [PicDownload] create request failed, error %s. obj_time: %s obj_timestamp：%d cid: %d.", err, common.ParseTimeStamp(timestamp), timestamp, cid))
		if retry > 0 {
			retry--
			flag = PicDownload(cid, size, timestamp, retry, timeout, token, objId, ossUrl, logger)
		}
		return
	}
	client := http.Client{Timeout: time.Second * timeout}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("imgDownload [PicDownload] execute request failed, error %v. obj_time: %s obj_timestamp：%d cid: %d.", err, common.ParseTimeStamp(timestamp), timestamp, cid))
		if retry > 0 {
			retry--
			flag = PicDownload(cid, size, timestamp, retry, timeout, token, objId, ossUrl, logger)
		}
		return
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("imgDownload [PicDownload] read request body failed, error %s. obj_time: %s obj_timestamp：%d cid: %d.", err, common.ParseTimeStamp(timestamp), timestamp, cid))
		if retry > 0 {
			retry--
			flag = PicDownload(cid, size, timestamp, retry, timeout, token, objId, ossUrl, logger)
		}
		return
	}
	if resp.StatusCode == 200 {
		if int(resp.ContentLength) == size {
			flag = true
			logger.Info(fmt.Sprintf("imgDownload [PicDownload] download file success. obj_time: %s obj_timestamp：%d cid: %d objectId: %s.", common.ParseTimeStamp(timestamp), timestamp, cid, objId))
		} else {
			logger.Warn(fmt.Sprintf("imgDownload [PicDownload] download file is success, but size is not match. obj_time: %s obj_timestamp：%d cid: %d objectId: %s.", common.ParseTimeStamp(timestamp), timestamp, cid, objId))
		}
	} else {
		logger.Error(fmt.Sprintf("imgDownload [PicDownload] download file failed, reuqest code %d. obj_time: %s obj_timestamp：%d cid: %d objectId: %s.", resp.StatusCode, common.ParseTimeStamp(timestamp), timestamp, cid, objId))
	}
	if !flag && retry > 0 {
		retry--
		flag = PicDownload(cid, size, timestamp, retry, timeout, token, objId, ossUrl, logger)
	}
	return
}

func GetVideoTimer(cid, begin, end, retry int, timeout time.Duration, token, ossUrl string, logger *zap.Logger) (objId string, err error) {
	rs := common.RecordTs{}
	url := ossUrl + fmt.Sprintf("/oss/v1/%d/record/objects?begin=%d&end=%d&type=0&client_token=%s", cid, begin, end, token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("videoDownload [GetVideoTimer] create new request failed, error %v. cid: %d.", err, cid))
		if retry > 0 {
			retry--
			objId, err = GetVideoTimer(cid, begin, end, retry, timeout, token, ossUrl, logger)
		}
		return
	}
	client := http.Client{Timeout: time.Second * timeout}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("videoDownload [GetVideoTimer] get video timer request failed, error %v. cid: %d.", err, cid))
		if retry > 0 {
			retry--
			objId, err = GetVideoTimer(cid, begin, end, retry, timeout, token, ossUrl, logger)
		}
		return
	}
	if resp.StatusCode != 200 {
		logger.Error(fmt.Sprintf("videoDownload [GetVideoTimer] get video timer request failed, request status code %d. error %v. cid: %d.", resp.StatusCode, err, cid))
		if retry > 0 {
			retry--
			objId, err = GetVideoTimer(cid, begin, end, retry, timeout, token, ossUrl, logger)
			err = common.CustomError{"请求状态码异常"}
		}
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("videoDownload [GetVideoTimer] parse video timer response failed, error %v. cid: %d.", err, cid))
		if retry > 0 {
			retry--
			objId, err = GetVideoTimer(cid, begin, end, retry, timeout, token, ossUrl, logger)
		}
		return
	}
	err = json.Unmarshal(body, &rs)
	if err != nil {
		logger.Error(fmt.Sprintf("videoDownload [GetVideoTimer] json parse video timer response failed, error %v. cid: %d.", err, cid))
		if retry > 0 {
			retry--
			objId, err = GetVideoTimer(cid, begin, end, retry, timeout, token, ossUrl, logger)
		}
		return
	}
	if len(rs.TimeList) == 0 {
		logger.Error(fmt.Sprintf("videoDownload [GetVideoTimer] get video timer is empty. cid: %d.", cid))
		err = common.CustomError{"对象列表为空"}
		return
	} else {
		objId = rs.TimeList[0].Oid
	}
	if err != nil && retry > 0 {
		retry--
		objId, err = GetVideoTimer(cid, begin, end, retry, timeout, token, ossUrl, logger)
	}
	return
}

func VideoDownTs(cid, timestamp, retry int, timeout time.Duration, objId, token, ossUrl string, logger *zap.Logger) (flag bool) {
	url := ossUrl + fmt.Sprintf("/oss/v2/%d/record/ts/%s.ts?client_token=%s", cid, objId, token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("videoDownload [VideoDownTs] create new request failed, error %v. obj_time: %s obj_timestamp：%d cid：%d.", err, common.ParseTimeStamp(timestamp), timestamp, cid))
		if retry > 0 {
			retry--
			flag = VideoDownTs(cid, timestamp, retry, timeout, objId, token, ossUrl, logger)
		}
		return
	}
	client := http.Client{Timeout: time.Second * timeout}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("videoDownload [VideoDownTs] get video ts file request failed, error %v. obj_time: %s obj_timestamp：%d cid: %d.", err, common.ParseTimeStamp(timestamp), timestamp, cid))
		if retry > 0 {
			retry--
			flag = VideoDownTs(cid, timestamp, retry, timeout, objId, token, ossUrl, logger)
		}
		return
	}
	if resp.StatusCode == 200 {
		logger.Info(fmt.Sprintf("videoDownload [VideoDownTs] get video ts file success. obj_time: %s obj_timestamp：%d cid: %d objectId: %s.", common.ParseTimeStamp(timestamp), timestamp, cid, objId))
		flag = true
	} else {
		logger.Error(fmt.Sprintf("videoDownload [VideoDownTs] get video ts file request failed, request status code is %d. obj_time: %s obj_timestamp：%d cid：%d objectId: %s.", resp.StatusCode, common.ParseTimeStamp(timestamp), timestamp, cid, objId))
	}
	if !flag && retry > 0 {
		retry--
		flag = VideoDownTs(cid, timestamp, retry, timeout, objId, token, ossUrl, logger)
	}
	return
}

func UpdateDevice(cid, cycle int, timeout time.Duration, apiUrl, appKey string, logger *zap.Logger) (err error) {
	token := common.GetToken(cid, cycle, appKey)
	url := apiUrl + fmt.Sprintf("/v2/devices/%d/info?client_token=%s", cid, token)
	payload := strings.NewReader("{\n\t\"info\": {\n\t\t\"deleted\": 1 \n\t}\n}")
	req, err := http.NewRequest("PUT", url, payload)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		logger.Error(fmt.Sprintf("updateDevice [UpdateDevice] create update device request failed, error %v. cid: %d.", err, cid))
		return
	}
	client := http.Client{Timeout: time.Second * timeout}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("updateDevice [UpdateDevice] update device failed, error %v. cid: %d.", err, cid))
		return
	}
	if resp.StatusCode == 200 {
		logger.Info(fmt.Sprintf("updateDevice [UpdateDevice] update device success. cid: %d.", cid))
	} else {
		logger.Error(fmt.Sprintf("updateDevice [UpdateDevice] update device failed request status code is %d. cid: %d.", resp.StatusCode, cid))
		err = common.CustomError{}
	}
	return
}
