package request

import (
	"bytes"
	"crypto/md5"
	"dev/backup_oss/common"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func readeStoreFile(fileName string, logger *zap.Logger) (storeData *common.StoreData) {
	storeData = &common.StoreData{}
	_, err := os.Stat(fileName)
	if err != nil {
		return
	}
	readePtr, err := os.Open(fileName)
	defer readePtr.Close()
	if err != nil {
		fmt.Printf("%#v", err)
		logger.Fatal(fmt.Sprintf("%#v", err))
		return
	}
	decoder := json.NewDecoder(readePtr)
	err = decoder.Decode(storeData)
	if err != nil {
		fmt.Println("解码失败，err=", err)
		logger.Fatal(fmt.Sprintf("解码失败，err=", err))
		return
	} else {
		logger.Info(fmt.Sprintf("解码成功：%#v\n", storeData))
	}
	return
}

func writeStoreFile(fileName string, storeData *common.StoreData, logger *zap.Logger) {
	writerPrt, _ := os.Create(fileName)
	defer writerPrt.Close()
	encoder := json.NewEncoder(writerPrt)
	err := encoder.Encode(storeData)
	if err != nil {
		logger.Fatal(fmt.Sprintf("编码失败，err=%#v", err))
	} else {
		logger.Info("编码成功")
	}
}

func getMd5Sum(fileName string, logger *zap.Logger) string {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("%#v", err)
		logger.Warn(fmt.Sprintf("[getMd5Sum] %s file is not exist", fileName))
		return ""
	}
	defer file.Close()
	md5h := md5.New()
	io.Copy(md5h, file)
	md5Str := hex.EncodeToString(md5h.Sum(nil))
	return md5Str
}

func GetToken(conf common.Yaml, logger *zap.Logger) string {
	cid := conf.Cid
	apiUrl := conf.ApiUrl
	appId := conf.AppId
	appKey := conf.AppKey
	tokenInfo := common.TokenInfo{}
	payload := strings.NewReader("{\n\t\"cids\": [" + strconv.Itoa(cid) + "],\n\t\"duration\": 3600\n}")
	req, _ := http.NewRequest("POST", apiUrl, payload)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-App-Id", appId)
	req.Header.Add("X-App-Key", appKey)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[GetToken] send request failed. %#v", err)
		logger.Error(fmt.Sprintf("[GetToken] send request failed. %#v", err))
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("[GetToken] parse request body failed. %#v", err)
		logger.Error(fmt.Sprintf("[GetToken] parse request body failed. %#v", err))
	}
	err = json.Unmarshal(body, &tokenInfo)
	if err != nil {
		fmt.Printf("[GetToken] parse request body to json format failed. %#v", err)
		logger.Error(fmt.Sprintf("[GetToken] parse request body to json format failed. %#v", err))
	}
	return tokenInfo.Tokens[0].Token
}

func UploadFile(conf common.Yaml, token, paramName, fileName string, logger *zap.Logger) error {
	fmt.Println("[UploadFile] start upload file")
	basePath := conf.Log.Path
	if basePath == "" {
		basePath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	}
	storeFile := path.Join(basePath, conf.DataName)
	storeData := readeStoreFile(storeFile, logger)
	storeFileInfo := common.FileInfo{}
	md5Str := getMd5Sum(fileName, logger)
	for _,item := range storeData.Upload {
		if item.Area == conf.Area && item.Md5Str == md5Str{
			fmt.Printf("[UploadFile] file: %s has been upload.\n", fileName)
			logger.Info(fmt.Sprintf("[UploadFile] file: %s has been upload.\n", fileName))
			return nil
		}
	}
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		fmt.Printf("[UploadFile] upload file is not exist. %#v", err)
		logger.Error(fmt.Sprintf("[UploadFile] upload file is not exist. %#v", err))
		return err
	}
	fileSize := fileInfo.Size()
	ossUrl := conf.OssUrl + fmt.Sprintf("/oss/v1/%d/uploadObject?size=%d&expiretype=%d&client_token=%s",
		conf.Cid, fileSize, conf.Expiretype, token)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(fileName))
	if err != nil {
		fmt.Printf("[UploadFile] %#v", err)
		logger.Fatal(fmt.Sprintf("[UploadFile] %#v", err))
		return err
	}
	file, err := os.Open(fileName)
	if err != nil {
		logger.Fatal(fmt.Sprintf("[UploadFile] %#v", err))
		return err
	}
	defer file.Close()
	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Printf("[UploadFile] %#v", err)
		logger.Fatal(fmt.Sprintf("[UploadFile] %#v", err))
		return err
	}

	err = writer.Close()
	if err != nil {
		fmt.Printf("[UploadFile] %#v", err)
		logger.Fatal(fmt.Sprintf("[UploadFile] %#v", err))
	}
	request, err := http.NewRequest("POST", ossUrl, body)
	if err != nil {
		fmt.Printf("[UploadFile] %#v", err)
		logger.Fatal(fmt.Sprintf("[UploadFile] %#v", err))
		return err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Printf("[UploadFile] %#v", err)
		logger.Fatal(fmt.Sprintf("[UploadFile] %#v", err))
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[UploadFile] %#v", err)
		logger.Fatal(fmt.Sprintf("[UploadFile] %#v", err))
		return err
	}
	response := common.Response{}
	if resp.StatusCode == 200 {
		err = json.Unmarshal(respBody, &response)
		if err != nil {
			fmt.Printf("[UploadFile] %#v", err)
			logger.Fatal(fmt.Sprintf("[UploadFile] %#v", err))
			return err
		}
		storeFileInfo.Area = conf.Area
		storeFileInfo.Md5Str = md5Str
		storeFileInfo.FileName = fileName
		storeFileInfo.FileSize = int(fileSize)
		storeFileInfo.ObjId = response.Obj_id
		storeData.Upload = append(storeData.Upload,&storeFileInfo)
		writeStoreFile(storeFile, storeData, logger)
		fmt.Printf("up load file success. objectId: %s, file_size: %d, file_name: %s",
			response.Obj_id, response.File_size, response.Name)
		logger.Info(fmt.Sprintf("up load file success. objectId: %s, file_size: %d, file_name: %s",
			response.Obj_id, response.File_size, response.Name))
		return nil
	} else {
		fmt.Printf("up load file failed, status code %d, request body %s", resp.StatusCode, respBody)
		logger.Error(fmt.Sprintf("up load file failed, status code %d, request body %s", resp.StatusCode, respBody))
	}
	fmt.Println("up load file finished")
	return err
}

func DownloadFile(conf common.Yaml, token, objId string, logger *zap.Logger) {
	basePath := conf.Log.Path
	if basePath == "" {
		basePath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	}
	storeFile := path.Join(basePath, conf.DataName)
	storeFileInfo := common.FileInfo{}
	ossUrl := conf.OssUrl + fmt.Sprintf("/oss/v1/%d/objects/%s?client_token=%s", conf.Cid, objId, token)
	request, err := http.NewRequest("GET", ossUrl, nil)
	if err != nil {
		fmt.Printf("[DownloadFile] create request failed. %#v", err)
		logger.Error(fmt.Sprintf("[DownloadFile] create request failed. %#v", err))
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Printf("[DownloadFile] get request body failed. %#v", err)
		logger.Error(fmt.Sprintf("[DownloadFile] get request body failed. %#v", err))
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[DownloadFile] %#v", err)
		logger.Error(fmt.Sprintf("[DownloadFile] get request body failed. %#v", err))
	}
	fileName := strings.Trim(strings.Split(resp.Header.Get("Content-Disposition"), "=")[1], "\"")
	downloadPath := conf.Download.Path
	if downloadPath == "" {
		downloadPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	}
	downloadFile := path.Join(downloadPath, fileName)
	err = ioutil.WriteFile(downloadFile, respBody, 0644)
	if err != nil {
		fmt.Printf("[DownloadFile] %#v", err)
		logger.Error(fmt.Sprintf("[DownloadFile] %#v", err))
		return
	}
	md5Str := getMd5Sum(downloadFile, logger)
	storeData := readeStoreFile(storeFile, logger)
	fileInfo, err := os.Stat(downloadFile)
	if err != nil {
		fmt.Printf("[DownloadFile] upload file is not exist. %#v", err)
		logger.Error(fmt.Sprintf("[DownloadFile] upload file is not exist. %#v", err))
		return
	}
	fileSize := fileInfo.Size()
	storeFileInfo.FileSize = int(fileSize)
	storeFileInfo.FileName = fileName
	storeFileInfo.Md5Str = md5Str
	storeFileInfo.Area = conf.Area
	storeFileInfo.ObjId = objId
	storeData.Download = append(storeData.Download,&storeFileInfo)
	writeStoreFile(storeFile, storeData, logger)
}
