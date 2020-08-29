package request

import (
	"bytes"
	"dev/chain_check/common"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func GetToken(conf common.Config, logger *zap.Logger) (token string, err error) {
	cid := conf.Cid
	apiUrl := conf.ApiUrl + "/v2/devices/tokens"
	appId := conf.AppId
	appKey := conf.AppKey
	tokenInfo := common.TokenInfo{}
	payload := strings.NewReader("{\n\t\"cids\": [" + strconv.Itoa(cid) + "]\n}")
	req, _ := http.NewRequest("POST", apiUrl, payload)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-App-Id", appId)
	req.Header.Add("X-App-Key", appKey)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("[GetToken] send request failed. %#v", err))
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("[GetToken] parse request body failed. %#v", err))
		return
	}
	err = json.Unmarshal(body, &tokenInfo)
	if err != nil {
		logger.Error(fmt.Sprintf("[GetToken] parse request body to json format failed. %#v", err))
		return
	}
	token = tokenInfo.Tokens[0].Token
	return
}

func Upload2(config common.Config, token, fileName string, logger *zap.Logger, key bool) (response common.ResponseUpload2,err error) {
	response = common.ResponseUpload2{}
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload2] upload file is not exist. %v", err))
		return
	}
	fileSize := fileInfo.Size()
	endPrefix := fmt.Sprintf("/oss/v1/%d/uploadObject?size=%d&expiretype=%d&client_token=%s",
		config.Cid, fileSize, config.ExpireType, token)
	if key {
		u, _ := uuid.NewV4()
		endPrefix = fmt.Sprintf("/oss/v1/%d/uploadObject?size=%d&expiretype=%d&client_token=%s&key=/%s",
			config.Cid, fileSize, config.ExpireType, token, u)
	}
	ossUrl := config.OssUrl + endPrefix
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(fileName))
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload2] create form file failed %v", err))
		return
	}
	file, err := os.Open(fileName)
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload2] open file failed. %v", err))
		return
	}
	defer file.Close()
	_, err = io.Copy(part, file)
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload2] copy file content to form failed %v", err))
		return
	}
	err = writer.Close()
	if err != nil {
		logger.Fatal(fmt.Sprintf("[Upload2] %v", err))
		return
	}
	request, err := http.NewRequest("POST", ossUrl, body)
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload2] create request failed %v", err))
		return
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload2] execute request failed %v", err))
		return
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload2] reade request content failed %v", err))
		return
	}
	if resp.StatusCode == 200 {
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			logger.Error(fmt.Sprintf("[Upload2] json parse request content failed. %v", err))
			return
		}
		logger.Info("[Upload2] upload file success.")

	} else {
		err = common.CustomError{}
		logger.Error(fmt.Sprintf("[Upload2] upload file failed, status code %d, request body %s", resp.StatusCode, string(respBody)))
	}
	return
}

func Download(url string, logger *zap.Logger) (flag bool) {
	flag = false
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("[Download] create request failed. %s", err))
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("[Download] execute request failed. %s", err))
		return
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("[Download] read request body failed. %s", err))
		return
	}
	if resp.StatusCode == 200 {
		flag = true
		logger.Info("[Download] get object success.")
	} else {
		logger.Error(fmt.Sprintf("[Download] get object failed, reuqest code: %d", resp.StatusCode))
	}
	return
}

func DownLoadObjId(cid int, ossUrl, token, objId string, logger *zap.Logger) (flag bool) {
	url := ossUrl
	reg := regexp.MustCompile("client_token")
	e := reg.FindAllString(url,-1)
	if len(e) == 0{
		url = ossUrl + fmt.Sprintf("/oss/v1/%d/objects/%s?client_token=%s", cid, objId, token)
	}

	flag =Download(url, logger)
	return
}

func DownloadKey(cid int, ossUrl, token, key string, logger *zap.Logger) (flag bool) {
	url := ossUrl + fmt.Sprintf("/oss/v1/%d/key?client_token=%s&key=%s", cid, token, key)
	flag =Download(url, logger)
	return
}

func Check2(config common.Config, token, fileName string, logger *zap.Logger) (result int) {
	response,err := Upload2(config, token, fileName, logger, false)
	if err != nil{
		return
	}
	cid := config.Cid
	ossUrl := config.OssUrl
	flag := DownLoadObjId(cid, ossUrl, token, response.Obj_id, logger)
	if flag{
		result = 1
	}
	return result
}

func Check2Key(config common.Config, token, fileName string, logger *zap.Logger) (result int) {
	response,err := Upload2(config, token, fileName, logger, true)
	if err != nil{
		return
	}
	flag := DownloadKey(config.Cid, config.OssUrl, token, response.Key, logger)
	if flag{
		result = 1
	}
	return
}

func Upload3(config common.Config, token, fileName string, logger *zap.Logger) (response common.ResponseUpload3,err error) {
	response = common.ResponseUpload3{}
	_, err = os.Stat(fileName)
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload3] upload file is not exist. %v", err))
		return
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	f, err := os.Open(fileName)
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload3] open file failed. %v", err))
		return
	}
	defer f.Close()
	fw, err := writer.CreateFormField("message")
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload3] create form field failed. %v",err))
		return
	}
	message := fmt.Sprintf("{\"topic_id\": 0,\"channel_id\": 0,\"subject\": \"\",\"body\": {},\"delay_time\": 1000,\"attachments\": [{\"form_field\": \"file\",\"key\": \"\",\"area_id\": 0,\"metadata\": {},\"url\": \"\",\"file_name\": \"\",\"expiretype\": %d}]}", config.ExpireType)
	_, err = fw.Write([]byte(message))
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload3]  write form field message failed. %v",err))
		return
	}
	fw, err = writer.CreateFormFile("file", filepath.Base(fileName))
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload3] create form file failed. %v",err))
		return
	}
	if _, err = io.Copy(fw, f); err != nil {
		logger.Error(fmt.Sprintf("[Upload3] write form file content failed. %v",err))
		return
	}
	err = writer.Close()
	if err != nil {
		logger.Fatal(fmt.Sprintf("[Upload3] %v", err))
		return
	}
	url := config.OssUrl + fmt.Sprintf("/oss/v1/%d/uploadEvent?client_token=%s", config.Cid, token)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload3] create request failed. %v",err))
		return
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload3] execute request failed. %v",err))
		return
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("[Upload3] reade request content failed. %v",err))
		return
	}
	if resp.StatusCode == 200 {
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			logger.Error(fmt.Sprintf("[Upload3] json parse request body failed. %v",err))
			return
		}
		logger.Info("[Upload3] upload file success.")
	} else {
		err = common.CustomError{}
		logger.Error(fmt.Sprintf("[Upload3] upload file failed, request code: %d, request content: %s. %v",
			resp.StatusCode,string(respBody),err))

	}
	return
}

func Check3(config common.Config, token, fileName string, logger *zap.Logger) (result int) {
	response,err := Upload3(config, token, fileName, logger)
	if err != nil{
		return
	}
	cid := config.Cid
	ossUrl := config.OssUrl
	flag := DownLoadObjId(cid, ossUrl, token, response.Attachments[0].Key, logger)
	if flag{
		result = 1
	}
	return
}

func MawarUpload(config common.Config, token, fileName string, logger *zap.Logger) (flag bool) {
	flag = false
	_, err := os.Stat(fileName)
	if err != nil {
		logger.Error(fmt.Sprintf("[MawarUpload] upload file is not exist. %v", err))
		return
	}

	url := config.ApiUrl + fmt.Sprintf("/iermu/uploadImg?deviceID=%s&timeStamp=%d&imageID=%d&access_token=%s",
		config.Sn, time.Now().Unix(), time.Now().Unix()*1000, token)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", filepath.Base(fileName))
	if err != nil {
		logger.Fatal(fmt.Sprintf("[MawarUpload] %v", err))
		return
	}
	file, err := os.Open(fileName)
	if err != nil {
		logger.Fatal(fmt.Sprintf("[MawarUpload] %v", err))
		return
	}
	defer file.Close()
	_, err = io.Copy(part, file)
	if err != nil {
		logger.Fatal(fmt.Sprintf("[MawarUpload] %v", err))
		return
	}
	err = writer.Close()
	if err != nil {
		logger.Fatal(fmt.Sprintf("[MawarUpload] %v", err))
		return
	}
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		logger.Fatal(fmt.Sprintf("[MawarUpload] %v", err))
		return
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logger.Fatal(fmt.Sprintf("[MawarUpload] %v", err))
		return
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatal(fmt.Sprintf("[MawarUpload] %v", err))
		return
	}
	if resp.StatusCode == 200 {
		flag = true
		logger.Info("[MawarUpload] upload file success")
	} else {
		logger.Error(fmt.Sprintf("[Upload2] upload file failed, status code %d, request body %s", resp.StatusCode, respBody))
	}
	return
}

func MawarDownload(config common.Config, token string, logger *zap.Logger) (flag bool) {
	url := config.OssUrl + "/fileinfo/last_objs?type=3&count=1&access_token=" + token
	flag = false
	response := &common.ResponseMawar{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("[MawarDownload] create request failed. %v", err))
		return
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logger.Error(fmt.Sprintf("[MawarDownload] get request body failed. %v", err))
		return
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("[MawarDownload] parse request body failed. %v", err))
		return
	}
	if resp.StatusCode == 200 {
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			logger.Error(fmt.Sprintf("[MawarDownload] json parse request body failed. %#v", err))
			return
		}
		for _, obj := range response.Obj_infos {
			if (int(time.Now().Unix()) - obj.Upload_time) <= 30 {
				flag = true
				logger.Info("[MawarDownload] download mawar upload object success")
			}
		}
	}
	return
}

func MawarCheck(config common.Config, token, fileName string, logger *zap.Logger) (flag int) {
	uflag := MawarUpload(config, token, fileName, logger)
	if uflag {
		dflag := MawarDownload(config, token, logger)
		if dflag {
			flag = 1
		}
	}
	return flag

}
