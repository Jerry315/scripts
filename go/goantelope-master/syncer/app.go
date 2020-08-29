package syncer

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"sync"

	"git.topvdn.com/web/goantelope/lytoken"
)

// 错误变量
var (
	ErrTokenExpired = errors.New("token expired")
	ErrAppNotFound  = errors.New("app not found")
)

// App 从羚羊云中心同步 APP 信息
type App struct {
	ID         string      `json:"app_id"`
	Keys       []string    `json:"app_key"`
	Partitions []Partition `json:"partitions"`
}

// AppSyncer 羚羊云 APP 信息同步器
type AppSyncer struct {
	// 同步器通用属性
	*Common
	*LyHTTPCommon `json:",inline"`

	AppList []*App `json:"apps"`

	partitions *Partitions
	idxApps    map[string]*App

	mutex sync.RWMutex
}

// NewAppSyncer 创建羚羊云 APP 信息同步器
func NewAppSyncer(name, lyid, lykey, urlFmt string, interval int) *AppSyncer {
	syncer := &AppSyncer{
		Common:       &Common{},
		LyHTTPCommon: &LyHTTPCommon{},

		AppList: []*App{},

		partitions: NewPartitions(),
		idxApps:    map[string]*App{},

		mutex: sync.RWMutex{},
	}
	syncer.Common.Init(name, interval)
	syncer.LyHTTPCommon.Init(lyid, lykey, urlFmt)
	return syncer
}

// Run 启动 AppSyncer
func (syncer *AppSyncer) Run() {
	Run(syncer)
}

// CheckAPP 检查 APP ID 及 APP Key 是否正确
func (syncer *AppSyncer) CheckAPP(appID, appKey string) bool {
	app := syncer.FindByID(appID)
	if app == nil {
		return false
	}
	for _, key := range app.Keys {
		rv := subtle.ConstantTimeCompare([]byte(key), []byte(appKey))
		if rv == 1 {
			return true
		}
	}
	return false
}

// CheckToken 检查 Token 是否正确
func (syncer *AppSyncer) CheckToken(token *lytoken.Token) (bool, error) {
	if token.IsExpired() {
		return false, ErrTokenExpired
	}
	app := syncer.FindByCID(token.CID)
	if app == nil {
		return false, ErrAppNotFound
	}

	keysBytes := [][]byte{}
	for _, key := range app.Keys {
		keysBytes = append(keysBytes, []byte(key))
	}
	return token.IsValid(keysBytes), nil
}

// CheckTokenStr 检查 Token 字符串是否正确
func (syncer *AppSyncer) CheckTokenStr(tokenStr string) (bool, error) {
	token, err := lytoken.FromStr(tokenStr)
	if err != nil {
		return false, err
	}
	return syncer.CheckToken(token)
}

// FindByID 根据 APPID 查找 APP 信息
func (syncer *AppSyncer) FindByID(appID string) *App {
	syncer.mutex.RLock()
	app, ok := syncer.idxApps[appID]
	syncer.mutex.RUnlock()
	if !ok {
		return nil
	}
	return app
}

// FindByCID 根据 CID 查找 APP 信息
func (syncer *AppSyncer) FindByCID(cid uint32) *App {
	appID, found := syncer.partitions.Find(cid)
	if !found {
		return nil
	}
	if appID == "" {
		return nil
	}
	return syncer.FindByID(appID)
}

// Loads 解码羚羊云中心 Web 返回的 APP 信息数据
func (syncer *AppSyncer) Loads(content []byte) error {
	return json.Unmarshal(content, syncer)
}

// NeedUpdate 检查是否需要更新 APP 数据
func (syncer *AppSyncer) NeedUpdate() bool {
	return len(syncer.AppList) > 0
}

// Update 更新 APP 信息到同步器中, 应该先执行 Loads
func (syncer *AppSyncer) Update() {
	if syncer.idxApps == nil {
		syncer.idxApps = map[string]*App{}
	}
	if syncer.partitions == nil {
		syncer.partitions = NewPartitions()
	}
	syncer.partitions.Clear()

	idxApps := map[string]*App{}
	for _, app := range syncer.AppList {
		// 更新到 CID 段集合
		syncer.partitions.Update(app.ID, app.Partitions)
		// 更新到临时 ID 索引 Map
		idxApps[app.ID] = app
	}

	// 构造及替换 CID 段树
	syncer.partitions.Build()

	// 替换 ID 索引 Map
	syncer.mutex.Lock()
	syncer.idxApps = idxApps
	syncer.mutex.Unlock()
}

// Stop 停止同步定时器
func (syncer *AppSyncer) Stop() {
	if syncer.cancelFunc == nil {
		return
	}
	syncer.cancelFunc()
}
