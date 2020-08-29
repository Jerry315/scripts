package syncer

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	mockSyncCloudPath = "/v3/clouds"

	fullCloudContent = `
{
    "data":{
        "clouds": [
            {
                "id": "57b43efc11459781779f5978",
                "app_id": "oojd",
                "central_host": "https://jxsr-api.antelopecloud.cn",
                "tong_host": "https://jxsr-api.antelopecloud.cn",
                "oss_host": "https://jxsr-oss1.antelopecloud.cn",
                "location": "南昌市",
                "location_code": "360100",
                "cloud_type": "upcloud",
                "sync_url": "http://jxsr-api.antelopecloud.cn/sync"
            },
            {
                "id": "17b43efc11459781779f5978",
                "app_id": "objl",
                "central_host": "https://jxsh-api.antelopecloud.cn",
                "tong_host": "https://jxsh-api.antelopecloud.cn",
                "oss_host": "https://jxsh-oss1.antelopecloud.cn",
                "location": "南昌市上",
                "location_code": "360200",
                "cloud_type": "upcloud",
                "sync_url": "http://jxsh-api.antelopecloud.cn/sync"
            },
            {
                "id": "57b43efc11459781779f5979",
                "app_id": "djki",
                "central_host": "https://ncgx-api.antelopecloud.cn",
                "tong_host": "https://ncgx-api.antelopecloud.cn",
                "oss_host": "https://ncgx-oss1.antelopecloud.cn",
                "location": "贵溪市",
                "location_code": "%v",
                "cloud_type": "",
                "sync_url": ""
            },
            {
                "id": "27b43efc11459781779f5978",
                "app_id": "abcd",
                "central_host": "https://dbck-api.antelopecloud.cn",
                "tong_host": "https://dbck-api.antelopecloud.cn",
                "oss_host": "https://dbck-oss1.antelopecloud.cn",
                "location": "贵溪市下",
                "location_code": "360202",
                "cloud_type": "downcloud",
                "sync_url": "http://dbck-api.antelopecloud.cn/sync"
            },
            {
                "id": "37b43efc11459781779f5978",
                "app_id": "efdb",
                "central_host": "https://gxsh-api.antelopecloud.cn",
                "tong_host": "https://gxsh-api.antelopecloud.cn",
                "oss_host": "https://gxsh-oss1.antelopecloud.cn",
                "location": "贵溪市2",
                "location_code": "360203",
                "cloud_type": "downcloud",
                "sync_url": "http://dbck-api.antelopecloud.cn/sync"
            }
        ],
        "last_modified": "%f"
    },
    "message": "",
    "request_id": "70507923d1c97dcc990ef9a456ce10140d473ace"
}
	`
	noChangeCloudContent = `
{
    "clouds": [],
	"request_id": "48042a02dec64749b763fa8de31b7635"
}
	`
)

// global vars
var (
	lastCloudModified      = float64(time.Now().Unix()) // 云部署信息最近更新时间
	lastChangeLocationCode = 360681                     // 变化的地点编号段 End 上一次值
	locationCodeStep       = 2                          // 变化的地点编号段增加值
	cloudReqCounter        = 0                          // 云部署信息请求计数器
)

func mockServerCloudsHandler(w http.ResponseWriter, req *http.Request) {
	data := opData[cloudReqCounter%len(ops)]

	var content []byte
	switch data.op {
	case "u":
		lastCloudModified = float64(time.Now().Unix())
		lastChangeLocationCode += locationCodeStep
	case "w":
		content = []byte(wrongContent)
	}

	var sinceF float64
	since := req.FormValue("since")
	if since != "" {
		f, err := strconv.ParseFloat(since, 64)
		if err != nil {
			log.Printf("mock server: %v\n", err)
			sinceF = float64(0)
		} else {
			sinceF = f
		}
	} else {
		sinceF = float64(0)
	}

	log.Printf("since arg %f\n", sinceF)
	if len(content) == 0 {
		if sinceF < lastCloudModified {
			content = []byte(fmt.Sprintf(fullCloudContent, lastChangeLocationCode, lastCloudModified))
		} else {
			content = []byte(noChangeCloudContent)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(data.statusCode)
	_, err := w.Write(content)
	if err != nil {
		log.Printf("mock server: write response content err: %v\n", err)
	}

	cloudReqCounter++
}
