package syncer

import (
	"encoding/json"
	"sync/atomic"
)

// 回调类型
const (
	WebhookTypeInfo  = "device-info"  // 设备信息
	WebhookTypeState = "device-state" // 设备状态
)

// Webhook 回调配置
type Webhook struct {
	ID         string             `json:"id"`
	APPID      string             `json:"app_id"`
	URL        string             `json:"url"`
	Type       string             `json:"type"`
	Partitions []WebhookPartition `json:"partitions"`
}

// NoPartition 判断回调是否有指定 CID 段
func (webhook *Webhook) NoPartition() bool {
	return webhook.Partitions == nil || len(webhook.Partitions) == 0
}

// Is 判断回调是否属于指定 APP 下 CID 的回调
func (webhook *Webhook) Is(cid uint32, appID string) bool {
	if webhook.NoPartition() {
		if webhook.APPID == appID {
			return true
		}
		return false
	}

	for _, partition := range webhook.Partitions {
		if partition.Contains(cid) {
			return true
		}
	}
	return false
}

// WebhookPartition 回调配置的 CID 分段
type WebhookPartition struct {
	Begin uint32 `json:"begin"`
	End   uint32 `json:"end"`
}

// Contains 判断 CID 段是否包含指定的 CID
func (p WebhookPartition) Contains(cid uint32) bool {
	return cid >= p.Begin && cid <= p.End
}

// WebhookSyncer 羚羊云回调配置信息同步
type WebhookSyncer struct {
	*Common
	*LyHTTPCommon `json:",inline"`

	Webhooks        []*Webhook   `json:"webhooks"`
	fronzenWebhooks atomic.Value // map[string][]*Webhook 类型数据
}

// NewWebhookSyncer 创建羚羊云回调配置同步器
func NewWebhookSyncer(name, lyid, lykey, urlFmt string, interval int) *WebhookSyncer {
	syncer := &WebhookSyncer{
		Common:       &Common{},
		LyHTTPCommon: &LyHTTPCommon{},
		Webhooks:     []*Webhook{},
	}
	syncer.fronzenWebhooks.Store(map[string][]*Webhook{})
	syncer.Common.Init(name, interval)
	syncer.LyHTTPCommon.Init(lyid, lykey, urlFmt)
	return syncer
}

// Run 启动 WebhookSyncer
func (syncer *WebhookSyncer) Run() {
	Run(syncer)
}

// GetWebhooks 根据 CID, APPID 获取指定类型的回调列表
func (syncer *WebhookSyncer) GetWebhooks(cid uint32, appID, hookType string) []*Webhook {
	rtWebhooks := []*Webhook{}
	typedWebhooksI := syncer.fronzenWebhooks.Load()
	typedWebhooks, ok := typedWebhooksI.(map[string][]*Webhook)
	if !ok {
		return rtWebhooks
	}
	webhooks, ok := typedWebhooks[hookType]
	if !ok {
		return rtWebhooks
	}

	// 遍历检查指定类型的回调配置
	for _, webhook := range webhooks {
		if ok := webhook.Is(cid, appID); ok {
			rtWebhooks = append(rtWebhooks, webhook)
		}
	}
	return rtWebhooks
}

// Loads 解析并加载羚羊云中心 web 返回的回调配置数据
func (syncer *WebhookSyncer) Loads(content []byte) error {
	resp := &LyHTTPResp{}
	resp.Data = syncer
	return json.Unmarshal(content, resp)
}

// NeedUpdate 检查是否需要更新回调配置数据
func (syncer *WebhookSyncer) NeedUpdate() bool {
	return len(syncer.Webhooks) > 0
}

// Update 更新回调配置信息到同步器中, 应该先执行 Loads
func (syncer *WebhookSyncer) Update() {
	typedWebhooks := map[string][]*Webhook{
		WebhookTypeInfo:  []*Webhook{},
		WebhookTypeState: []*Webhook{},
	}
	for _, webhook := range syncer.Webhooks {
		switch webhook.Type {
		case WebhookTypeInfo:
			typedWebhooks[WebhookTypeInfo] = append(
				typedWebhooks[WebhookTypeInfo], webhook)
		case WebhookTypeState:
			typedWebhooks[WebhookTypeState] = append(
				typedWebhooks[WebhookTypeState], webhook)
		}
	}
	syncer.fronzenWebhooks.Store(typedWebhooks)
}

// Stop 停止 WebhookSyncer 定时器
func (syncer *WebhookSyncer) Stop() {
	if syncer.cancelFunc == nil {
		return
	}
	syncer.cancelFunc()
}
