package syncer

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"git.topvdn.com/web/goantelope/lytoken"
)

var (
	responseContent = struct {
		appList      map[string]*App
		lastModified string
		content      []byte
	}{
		map[string]*App{
			"test": &App{
				ID: "test",
				Keys: []string{
					"87938d4d094b883ac644098802f12359",
					"3262029cb9f8f9abd161fd678d012092",
				},
				Partitions: []Partition{
					Partition{Begin: 100041, End: 100051},
					Partition{Begin: 100061, End: 100071},
					Partition{Begin: 100081, End: 100091},
					Partition{Begin: 1000, End: 2000},
				},
			},
			"xybk": &App{
				ID:   "xybk",
				Keys: []string{"824d12de6f5f999bc12c822360b39cdcf"},
				Partitions: []Partition{
					Partition{Begin: 3000, End: 4000},
					Partition{Begin: 200000, End: 300000},
				},
			},
			"": &App{
				ID:   "",
				Keys: []string{"824d12de6f5f999bc12c822360b39cdcf"},
				Partitions: []Partition{
					Partition{Begin: 19, End: 19},
				},
			},
		},
		"1523257200.23",
		[]byte(`
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
                    "end": 300000,
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
    "last_modified": "1523257200.23",
    "request_id": "48042a02dec64749b763fa8de31b7635"
}
			`),
	}
	appData = []struct {
		appID string
		ok    bool
	}{
		{"test", true},
		{"xybk", true},
		{"java", false},
	}

	testLyID  = "test"
	testLyKey = "824d12de6f5f999bc12c822360b39cdcf"
)

func TestAppSyncerLoads(t *testing.T) {
	assert := assert.New(t)

	syncer := NewAppSyncer("appsyncer", "", "", "", 10)
	err := syncer.Loads(responseContent.content)
	assert.Nil(err)
	assert.Equal(responseContent.lastModified, syncer.LastModified)
	assert.Equal(len(responseContent.appList), len(syncer.AppList))

	idxAppMap := map[string]*App{}
	for _, app := range responseContent.appList {
		idxAppMap[app.ID] = app
	}

	for _, app := range syncer.AppList {
		oriApp, ok := idxAppMap[app.ID]
		assert.Equal(true, ok)
		assert.NotNil(oriApp)
		assert.Equal(oriApp.ID, app.ID)

		for _, key := range app.Keys {
			assert.Contains(oriApp.Keys, key)
		}

		beginIdxPartitions := map[uint32]Partition{}
		for _, partition := range oriApp.Partitions {
			beginIdxPartitions[partition.Begin] = partition
		}
		for _, partition := range app.Partitions {
			oriPartition, ok := beginIdxPartitions[partition.Begin]
			assert.Equal(true, ok)
			assert.NotNil(oriPartition)
			assert.Equal(oriPartition.Begin, partition.Begin)
			assert.Equal(oriPartition.End, partition.End)
		}
	}
}

func TestCheckToken(t *testing.T) {
	assert := assert.New(t)

	tokenStr1, _ := lytoken.New(1000, 0, time.Minute).Str([]byte("87938d4d094b883ac644098802f12359"))
	tokenStr2, _ := lytoken.New(100051, 0, time.Minute).Str([]byte("3262029cb9f8f9abd161fd678d012092"))
	tokenStr3, _ := lytoken.New(100065, 0, -time.Minute).Str([]byte("3262029cb9f8f9abd161fd678d012092"))
	tokenStr4, _ := lytoken.New(1, 0, time.Minute).Str([]byte("3262029cb9f8f9abd161fd678d012092"))
	testTokens := []struct {
		tokenStr string
		ok       bool
	}{
		{tokenStr1, true},
		{tokenStr2, true},
		{tokenStr3, false},
		{tokenStr4, false},
	}

	syncer := NewAppSyncer("appsyncer", "", "", "", 10)
	err := syncer.Loads(responseContent.content)
	assert.Nil(err)
	syncer.Update()

	for _, tokenData := range testTokens {
		ok, err := syncer.CheckTokenStr(tokenData.tokenStr)
		assert.Equal(tokenData.ok, ok)
		t.Log(tokenData.tokenStr, ok, err)
	}
}

func TestAppSyncerUpdate(t *testing.T) {
	assert := assert.New(t)
	syncer := &AppSyncer{}
	err := syncer.Loads(responseContent.content)
	assert.Nil(err)

	syncer.Update()
	assert.Equal(
		testUtilGetAppPartitionCnt(syncer),
		testUtilPartitionCnt(syncer.partitions))
	assert.Equal(testUtilGetAppPartitionCnt(syncer), syncer.partitions.tree.Size())
}

func TestAppSyncerFinders(t *testing.T) {
	assert := assert.New(t)
	syncer := NewAppSyncer("appsyncer", "", "", "", 10)
	err := syncer.Loads(responseContent.content)
	assert.Nil(err)
	syncer.Update()

	for _, data := range appData {
		app := syncer.FindByID(data.appID)
		if data.ok {
			assert.NotNil(app)
			assert.Equal(data.appID, app.ID)
		} else {
			assert.Nil(app)
		}
	}

	for _, data := range cidData {
		app := syncer.FindByCID(data.cid)
		if data.ok {
			assert.NotNil(app)
			assert.Equal(data.appID, app.ID)
		} else {
			assert.Nil(app)
		}
	}
}

func TestAppSyncerSync(t *testing.T) {
	assert := assert.New(t)

	url := fmt.Sprintf("http://%v%v?since=", mockServerBind, mockSyncAppPath)
	urlFmt := url + "%v"
	syncer := NewAppSyncer("appsyncer", testLyID, testLyKey, urlFmt, 1)

	// 清除 ctx, 测试重建
	syncer.ctx = nil
	// 清除 http client, 测试重新创建
	syncer.client = nil

	Run(syncer)
	syncer.LastModified = ""
	time.Sleep(8 * time.Second)
	// 清除请求路径, 测试异常情况
	syncer.urlFmt = "http://%v"
	time.Sleep(2 * time.Second)

	syncer.Stop()
	time.Sleep(2 * time.Second)

	for _, data := range cidData {
		app := syncer.FindByCID(data.cid)
		if data.ok {
			assert.NotNil(app)
			assert.Equal(data.appID, app.ID)
		} else {
			assert.Nil(app)
		}
	}

	// 检查是否更新了 CID 段数据
	cid := lastChangePartitionEnd
	appID := "xybk"
	t.Log(cid)
	t.Log(syncer.partitions)
	app := syncer.FindByCID(uint32(cid))
	assert.NotNil(app)
	assert.Equal(appID, app.ID)

	// 错误执行
	syncer.cancelFunc = nil
	syncer.Stop()
}

func testUtilGetAppPartitionCnt(syncer *AppSyncer) int {
	cnt := 0
	for _, app := range syncer.AppList {
		cnt += len(app.Partitions)
	}
	return cnt
}

func init() {
	// 启动测试服务器
	go startMockServer()

}
