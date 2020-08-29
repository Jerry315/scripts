package syncer

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	mockServerBind  = "127.0.0.1:8090"
	mockSyncAppPath = "/v2/keys"

	fullContent = `
{
    "apps": [
        {
            "app_id": "test",
            "app_key": [
                "87938d4d094b883ac644098802f12359",
                "3262029cb9f8f9abd161fd678d012092"
            ],
            "partitions": [
                {
                    "end": 100051,
                    "start": 100041
                },
                {
                    "end": 100071,
                    "start": 100061
                },
                {
                    "end": 100091,
                    "start": 100081
                },
                {
                    "end": 2000,
                    "start": 1000
                }
            ],
            "webhook": ""
        },
        {
            "app_id": "xybk",
            "app_key": [
                "824d12de6f5f999bc12c822360b39cdcf"
            ],
            "partitions": [
                {
                    "end": 4000,
                    "start": 3000
                },
                {
                    "end": %v,
                    "start": 200000
                }
            ],
            "webhook": ""
        },
        {
            "app_id": "",
            "app_key": [
                "824d12de6f5f999bc12c822360b39cdcf"
            ],
            "partitions": [
                {
                    "end": 19,
                    "start": 19
                }
            ],
            "webhook": ""
        }
    ],
    "last_modified": "%f",
    "request_id": "48042a02dec64749b763fa8de31b7635"
}
	`
	noChangeContent = `
{
    "apps": [],
	"request_id": "48042a02dec64749b763fa8de31b7635"
}
	`
	wrongContent = "lala not a json"
	ops          = []string{"u", "o", "w", "u", "o", "w", "u", "o"}
	opData       = []struct {
		// op: o -> old, u -> update
		op         string
		statusCode int
	}{
		{"u", http.StatusOK},
		{"o", http.StatusOK},
		{"w", http.StatusOK},
		{"u", http.StatusOK},
		{"o", http.StatusOK},
		{"w", http.StatusBadRequest},
		{"u", http.StatusOK},
		{"o", http.StatusOK},
	}
)

// global vars
var (
	lastModified           = float64(time.Now().Unix()) // 最新更新时间
	lastChangePartitionEnd = 300000                     // 变化的 CID 段 End 上一次值
	partitionStep          = 10                         // 变化的 CID 段增加值
	reqCounter             = 0                          // 请求计数器
)

func startMockServer() {
	mux := http.NewServeMux()
	mux.HandleFunc(mockSyncAppPath, mockServerAppInfoHandler)
	mux.HandleFunc(mockSyncCloudPath, mockServerCloudsHandler)

	server := &http.Server{
		Addr:    mockServerBind,
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}

func mockServerAppInfoHandler(w http.ResponseWriter, req *http.Request) {
	data := opData[reqCounter%len(ops)]

	var content []byte
	switch data.op {
	case "u":
		lastModified = float64(time.Now().Unix())
		lastChangePartitionEnd += partitionStep
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
		if sinceF < lastModified {
			content = []byte(fmt.Sprintf(fullContent, lastChangePartitionEnd, lastModified))
		} else {
			content = []byte(noChangeContent)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(data.statusCode)
	_, err := w.Write(content)
	if err != nil {
		log.Printf("mock server: write response content err: %v\n", err)
	}

	reqCounter++
}
