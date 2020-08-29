package handler

import (
	"dev/elastic_tools/common"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func ESClient(esUrl, logFile string) (client *elastic.Client, err error) {
	logFileHandle, _ := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	cfg := []elastic.ClientOptionFunc{
		elastic.SetURL(esUrl),
		elastic.SetSniff(false),
		elastic.SetInfoLog(log.New(logFileHandle, "ES-INFO: ", 0)),
		elastic.SetTraceLog(log.New(logFileHandle, "ES-TRACE: ", 0)),
		elastic.SetErrorLog(log.New(logFileHandle, "ES-ERROR: ", 0)),
	}
	client, err = elastic.NewClient(cfg ...)
	return
}

func CreateIndex(client *elastic.Client, index, mapping string, logger *zap.Logger) {
	//mapping := `{
	//	"settings":{
	//		"number_of_shards":1,
	//       "number_of_replicas": 0
	//	},
	//	"mappings": {
	//		"tweet": {
	//			"properties": {
	//				"tags": {
	//					"type": "string"
	//				},
	//				"location": {
	//					"type": "geo_point"
	//				}
	//			}
	//		}
	//	}
	//}`
	ctx := context.Background()
	createIndex, err := client.CreateIndex(index).BodyString(mapping).Do(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("create index:%s failed. %v", index, err))
		os.Exit(-1)
	}
	if !createIndex.Acknowledged {
		fmt.Println("!createIndex.Acknowledged")
		logger.Warn("!createIndex.Acknowledged")
	} else {
		fmt.Println("createIndex.Acknowledged")
		logger.Info("createIndex.Acknowledged")
	}
}

func DeleteIndex(client *elastic.Client, config common.Config, logger *zap.Logger) {
	ctx := context.Background()
	indiceList := []string{}
	for _, item := range config.Delete {
		if item.Enable {
			for _, index := range item.Index {
				sd, _ := time.ParseDuration(strconv.Itoa(item.DelayDays*(-24)) + "h")
				indiceList = append(indiceList, index+time.Now().Add(sd).Format(item.DateFmt))
			}
		}
	}
	deleteIndex, err := client.DeleteIndex(strings.Join(indiceList, ",")).Do(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("delete indices: %s failed. %v", strings.Join(indiceList, ","), err))
	}
	if !deleteIndex.Acknowledged {
		fmt.Println("!deleteIndex.Acknowledged")
	} else {
		fmt.Println("deleteIndex.Acknowledged")
	}
}

func SnapShotRepository(client *elastic.Client, config common.Config, repository string, logger *zap.Logger) {
	mapping := fmt.Sprintf(`{
           "type": "fs",
           "settings": {
               "location": "%s",
               "max_snapshot_bytes_per_sec": "%s",
               "max_restore_bytes_per_sec": "%s"
           }
	}`, repository, config.MaxSnapshotBytesPerSec, config.MaxRestoreBytesPerSec)
	ctx := context.Background()
	snapShotRepository, err := client.SnapshotCreateRepository(repository).BodyString(mapping).Do(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("create snapshot repository:%s failed. %v", repository, err))
		os.Exit(-1)
	}
	if !snapShotRepository.Acknowledged {
		logger.Warn("!snapShotRepository.Acknowledged")
	} else {
		logger.Info("snapShotRepository.Acknowledged")
	}

}

func SnapShot(client *elastic.Client, config common.Config, repository, snapshot string, logger *zap.Logger) {
	indiceList := []string{}
	for _, item := range config.Snapshot {
		if item.Enable {
			for _, index := range item.Index {
				sd, _ := time.ParseDuration(strconv.Itoa(item.DelayDays*(-24)) + "h")
				indiceList = append(indiceList, index+time.Now().Add(sd).Format(item.DateFmt))
			}
		}
	}
	mapping := fmt.Sprintf(`{
		"indices": "%s",
		"ignore_unavailable": true,
		"include_global_state": false
	}`, strings.Join(indiceList, ","))
	ctx := context.Background()
	snapShotIndex, err := client.SnapshotCreate(repository, snapshot).BodyString(mapping).Do(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("create snapshot:%s, repository:%s failed. %v", snapshot, repository, err))
		os.Exit(-1)
	}
	if !*snapShotIndex.Accepted {
		logger.Warn("!snapShotIndex.Accepted")
	} else {
		logger.Info("snapShotIndex.Accepted")
		response, _ := json.Marshal(snapShotIndex.Snapshot)
		logger.Info(string(response))
	}

}

func GetSnapShot(client *elastic.Client, repository string, logger *zap.Logger) {
	ctx := context.Background()
	gsi, err := client.SnapshotGet(repository).Do(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("get %s snapshot failed. %v", repository, err))
		os.Exit(-1)
	}
	response, err := json.Marshal(gsi)
	if err != nil {
		logger.Error(fmt.Sprintf("json parse get snapshot response failed. %v", err))
		os.Exit(-1)
	}
	fmt.Println(string(response))
}

func DeleteSnapShot(client *elastic.Client, repository, snapshot string, logger *zap.Logger) {
	ctx := context.Background()
	sd, err := client.SnapshotDelete(repository, snapshot).Do(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("delete snapshot:%s failed. %v", snapshot, err))
		os.Exit(-1)
	}
	if !sd.Acknowledged {
		fmt.Println("!sd.Acknowledged")
		logger.Warn("!sd.Acknowledged")
	} else {
		fmt.Println("sd.Acknowledged")
		logger.Info("sd.Acknowledged")
	}
}

func SetTag(client *elastic.Client, config common.Config, logger *zap.Logger) {
	ctx := context.Background()
	for _, item := range config.Settings {
		if item.Enable {
			mapping := fmt.Sprintf(`{"index.routing.allocation.require.tag": "%s"}`, item.Tag)
			for _, index := range item.Index {
				sd, _ := time.ParseDuration(strconv.Itoa(item.DelayDays*(-24)) + "h")
				indice := index + time.Now().Add(sd).Format(item.DateFmt)
				st, err := client.IndexPutSettings(indice).BodyString(mapping).Do(ctx)
				if err != nil {
					logger.Error(fmt.Sprintf("indices: %s set arrguments failed. %v", indice, err))
					os.Exit(-1)
				}
				if !st.Acknowledged {
					fmt.Println("!st.Acknowledged")
					logger.Warn("!st.Acknowledged")
				} else {
					fmt.Println("st.Acknowledged")
					logger.Info("st.Acknowledged")
				}
			}

		}
	}

}

func GetTag(client *elastic.Client, indices string, logger *zap.Logger) {
	ctx := context.Background()
	gs, err := client.IndexGetSettings(indices).Do(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("get indices: %s failed. %v", indices, err))
		os.Exit(-1)
	}
	response, _ := json.Marshal(gs)
	fmt.Println(string(response))
}
