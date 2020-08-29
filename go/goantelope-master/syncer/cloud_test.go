package syncer

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	upCloud1 = &Cloud{
		APPID:        "oojd",
		CentralHost:  "https://jxsr-api.antelopecloud.cn",
		TongHost:     "https://jxsr-api.antelopecloud.cn",
		OSSHost:      "https://jxsr-oss1.antelopecloud.cn",
		Location:     "南昌市",
		LocationCode: "360100",
		CloudType:    "upcloud",
		SyncURL:      "http://jxsr-api.antelopecloud.cn/sync",
	}
	upCloud2 = &Cloud{
		APPID:        "objl",
		CentralHost:  "https://jxsh-api.antelopecloud.cn",
		TongHost:     "https://jxsh-api.antelopecloud.cn",
		OSSHost:      "https://jxsh-oss1.antelopecloud.cn",
		Location:     "南昌市上",
		LocationCode: "360200",
		CloudType:    "upcloud",
		SyncURL:      "http://jxsh-api.antelopecloud.cn/sync",
	}
	currentCloud = &Cloud{
		APPID:        "djki",
		CentralHost:  "https://ncgx-api.antelopecloud.cn",
		TongHost:     "https://ncgx-api.antelopecloud.cn",
		OSSHost:      "https://ncgx-oss1.antelopecloud.cn",
		Location:     "贵溪市",
		LocationCode: "360681",
		CloudType:    "currentcloud",
		SyncURL:      "",
	}
	downCloud1 = &Cloud{
		APPID:        "abcd",
		CentralHost:  "https://dbck-api.antelopecloud.cn",
		TongHost:     "https://dbck-api.antelopecloud.cn",
		OSSHost:      "https://dbck-oss1.antelopecloud.cn",
		Location:     "贵溪市下",
		LocationCode: "360202",
		CloudType:    "downcloud",
		SyncURL:      "http://dbck-api.antelopecloud.cn/sync",
	}
	downCloud2 = &Cloud{
		APPID:        "efdb",
		CentralHost:  "https://gxsh-api.antelopecloud.cn",
		TongHost:     "https://gxsh-api.antelopecloud.cn",
		OSSHost:      "https://gxsh-oss1.antelopecloud.cn",
		Location:     "贵溪市2",
		LocationCode: "360203",
		CloudType:    "downcloud",
		SyncURL:      "http://gxsh-api.antelopecloud.cn/sync",
	}
	upClouds = map[string]*Cloud{
		upCloud1.APPID: upCloud1,
		upCloud2.APPID: upCloud2,
	}
	downClouds = map[string]*Cloud{
		downCloud1.APPID: downCloud1,
		downCloud2.APPID: downCloud2,
	}
	cloudsResponseContent = struct {
		clouds       map[string]*Cloud
		lastModified string
		content      []byte
	}{
		map[string]*Cloud{
			upCloud1.APPID:     upCloud1,
			upCloud2.APPID:     upCloud2,
			currentCloud.APPID: currentCloud,
			downCloud1.APPID:   downCloud1,
			downCloud2.APPID:   downCloud2,
		},
		"1441676854.32",
		[]byte(`
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
                "location_code": "360681",
                "cloud_type": "currentcloud",
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
                "sync_url": "http://gxsh-api.antelopecloud.cn/sync"
            }
        ],
        "last_modified": "1441676854.32"
    },
    "message": "",
    "request_id": "70507923d1c97dcc990ef9a456ce10140d473ace"
}
		`),
	}
	cloudData = []struct {
		appID       string
		location    string
		centralHost string
		ok          bool
	}{
		{"oojd", "南昌市", "https://jxsr-api.antelopecloud.cn", true},
		{"djki", "贵溪市", "https://ncgx-api.antelopecloud.cn", true},
		{"xxbb", "不知道哪里", "https://bzdnl-api.topvdn.com", false},
		{"efdb", "贵溪市2", "https://gxsh-api.antelopecloud.cn", true},
	}
)

func TestCloudSyncerLoads(t *testing.T) {
	assert := assert.New(t)

	syncer := NewCloudSyncer("cloudsyncer", "", "", "", 10)
	err := syncer.Loads(cloudsResponseContent.content)
	assert.Nil(err)
	assert.Equal(cloudsResponseContent.lastModified, syncer.LastModified)

	for _, cloud := range syncer.Clouds {
		oriCloud, ok := cloudsResponseContent.clouds[cloud.APPID]
		assert.Equal(true, ok)
		assert.NotNil(oriCloud)
		assert.Equal(true, testUtilCompareCloud(oriCloud, cloud))
	}
}

func TestCloudSyncerUpdate(t *testing.T) {
	assert := assert.New(t)
	syncer := &CloudSyncer{}
	err := syncer.Loads(cloudsResponseContent.content)
	assert.Nil(err)

	syncer.Update()
	syncerUpClouds := syncer.UpClouds()
	syncerDownClouds := syncer.DownClouds()
	syncerCurrentCloud := syncer.CurrentCloud()
	assert.Equal(len(upClouds), len(syncerUpClouds))
	assert.Equal(len(downClouds), len(syncerDownClouds))
	assert.NotNil(syncerCurrentCloud)
	assert.Equal(currentCloud.APPID, syncerCurrentCloud.APPID)
	assert.Equal(currentCloud.SyncURL, syncerCurrentCloud.SyncURL)

	for _, cloud := range syncerUpClouds {
		oriCloud, ok := upClouds[cloud.APPID]
		assert.Equal(true, ok)
		assert.NotNil(oriCloud)
		assert.Equal(true, testUtilCompareCloud(oriCloud, cloud))
	}
}

func TestCloudSyncerFindByAPPID(t *testing.T) {
	assert := assert.New(t)
	syncer := NewCloudSyncer("cloudsyncer", "", "", "", 10)
	err := syncer.Loads(cloudsResponseContent.content)
	assert.Nil(err)
	syncer.Update()

	for _, data := range cloudData {
		cloud := syncer.FindByAPPID(data.appID)
		if data.ok {
			assert.NotNil(cloud)
			assert.Equal(data.appID, cloud.APPID)
			assert.Equal(data.centralHost, cloud.CentralHost)
		} else {
			assert.Nil(cloud)
		}
	}
}

func TestCloudSyncerSync(t *testing.T) {
	assert := assert.New(t)

	url := fmt.Sprintf("http://%v%v?since=", mockServerBind, mockSyncCloudPath)
	urlFmt := url + "%v"
	syncer := NewCloudSyncer("cloudsyncer", testLyID, testLyKey, urlFmt, 1)

	Run(syncer)
	time.Sleep(8 * time.Second)

	for _, data := range cloudData {
		cloud := syncer.FindByAPPID(data.appID)
		if data.ok {
			assert.NotNil(cloud)
			assert.Equal(data.appID, cloud.APPID)
			assert.Equal(data.centralHost, cloud.CentralHost)
		} else {
			assert.Nil(cloud)
		}
	}
}

func testUtilCompareCloud(a, b *Cloud) bool {
	if a.APPID != b.APPID {
		return false
	}
	if a.CentralHost != b.CentralHost {
		return false
	}
	if a.TongHost != b.TongHost {
		return false
	}
	if a.OSSHost != b.OSSHost {
		return false
	}
	if a.Location != b.Location {
		return false
	}
	if a.LocationCode != b.LocationCode {
		return false
	}
	if a.CloudType != b.CloudType {
		return false
	}
	if a.SyncURL != b.SyncURL {
		return false
	}
	return true
}
