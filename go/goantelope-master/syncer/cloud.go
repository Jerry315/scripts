package syncer

import (
	"encoding/json"
	"sync"
)

const (
	// UpCloud 上级云
	UpCloud = "upcloud"
	// DownCloud 下级云
	DownCloud = "downcloud"
	// CurrentCloud 本级云
	CurrentCloud = "currentcloud"
)

// Cloud 从羚羊云中心同步的云部署信息
type Cloud struct {
	APPID        string `json:"app_id"`
	CentralHost  string `json:"central_host"`
	TongHost     string `json:"tong_host"`
	OSSHost      string `json:"oss_host"`
	Location     string `json:"location"`
	LocationCode string `json:"location_code"`
	CloudType    string `json:"cloud_type"`
	SyncURL      string `json:"sync_url"`
}

// CloudSyncer 羚羊云云部署信息同步器
type CloudSyncer struct {
	*Common
	*LyHTTPCommon `json:",inline"`

	Clouds []*Cloud `json:"clouds"`

	upClouds     []*Cloud
	downClouds   []*Cloud
	currentCloud *Cloud
	idxClouds    map[string]*Cloud

	mutex sync.RWMutex
}

// NewCloudSyncer 创建云部署信息同步器
func NewCloudSyncer(name, lyid, lykey, urlFmt string, interval int) *CloudSyncer {
	syncer := &CloudSyncer{
		Common:       &Common{},
		LyHTTPCommon: &LyHTTPCommon{},

		Clouds: []*Cloud{},

		upClouds:   []*Cloud{},
		downClouds: []*Cloud{},
		idxClouds:  map[string]*Cloud{},

		mutex: sync.RWMutex{},
	}
	syncer.Common.Init(name, interval)
	syncer.LyHTTPCommon.Init(lyid, lykey, urlFmt)
	return syncer
}

// Run CloudSyncer
func (syncer *CloudSyncer) Run() {
	Run(syncer)
}

// CurrentCloud 当前云部署信息
func (syncer *CloudSyncer) CurrentCloud() *Cloud {
	return syncer.currentCloud
}

// UpClouds 上级云列表
func (syncer *CloudSyncer) UpClouds() []*Cloud {
	return syncer.upClouds
}

// DownClouds 下级云列表
func (syncer *CloudSyncer) DownClouds() []*Cloud {
	return syncer.downClouds
}

// FindByAPPID 根据 APPID 查找部署信息
func (syncer *CloudSyncer) FindByAPPID(appID string) *Cloud {
	syncer.mutex.RLock()
	cloud, ok := syncer.idxClouds[appID]
	syncer.mutex.RUnlock()
	if !ok {
		return nil
	}
	return cloud
}

// Loads 解码加载请求数据
func (syncer *CloudSyncer) Loads(content []byte) error {
	resp := &LyHTTPResp{}
	resp.Data = syncer
	return json.Unmarshal(content, resp)
}

// NeedUpdate 检查是否需要更新
func (syncer *CloudSyncer) NeedUpdate() bool {
	return len(syncer.Clouds) > 0
}

// Update 更新请求到的数据到同步器
func (syncer *CloudSyncer) Update() {

	upClouds := []*Cloud{}
	downClouds := []*Cloud{}
	idxClouds := map[string]*Cloud{}
	for _, cloud := range syncer.Clouds {
		switch cloud.CloudType {
		case UpCloud:
			upClouds = append(upClouds, cloud)
		case DownCloud:
			downClouds = append(downClouds, cloud)
		case CurrentCloud:
			syncer.currentCloud = cloud
		}
		idxClouds[cloud.APPID] = cloud
	}

	syncer.mutex.Lock()
	syncer.upClouds = upClouds
	syncer.downClouds = downClouds
	syncer.idxClouds = idxClouds
	syncer.mutex.Unlock()
}
