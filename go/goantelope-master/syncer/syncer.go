package syncer

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Err 变量
var (
	ErrResponseNotOk = errors.New("http response code is not 200 OK")
)

// Syncer 数据同步类型
type Syncer interface {
	Name() string
	Interval() int
	Ctx() context.Context

	Request() ([]byte, error)
	Loads([]byte) error
	NeedUpdate() bool
	Update()
}

// Common 同步器通用属性类型
type Common struct {
	name       string
	interval   int
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// Name 返回同步器名称
func (c *Common) Name() string {
	return c.name
}

// Interval 返回同步器同步时间间隔, 单位秒
func (c *Common) Interval() int {
	return c.interval
}

// Ctx 返回同步器 context
func (c *Common) Ctx() context.Context {
	if c.ctx == nil || c.cancelFunc == nil {
		ctx, cancelFunc := context.WithCancel(context.Background())
		c.ctx = ctx
		c.cancelFunc = cancelFunc
	}
	return c.ctx
}

// Init 初始化
func (c *Common) Init(name string, interval int) {
	c.name = name
	c.interval = interval
	ctx, cancelFunc := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancelFunc = cancelFunc
}

// Run 启动数据同步器
func Run(syncer Syncer) {
	syncFunc := func() error {
		content, err := syncer.Request()
		if err != nil {
			return err
		}
		err = syncer.Loads(content)
		if err != nil {
			return err
		}

		if !syncer.NeedUpdate() {
			return nil
		}
		syncer.Update()
		return nil
	}

	go func() {
		log.Printf("syncer: %v first sync\n", syncer.Name())
		err := syncFunc()
		if err != nil {
			log.Printf("syncer: %v first sync error: %v\n", syncer.Name(), err)
		}
	}()
	timer := time.NewTicker(time.Duration(syncer.Interval()) * time.Second)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				err := syncFunc()
				if err != nil {
					log.Printf("syncer: %v sync failed: %v\n", syncer.Name(), err)
				}
			}
		}
	}(syncer.Ctx())
}

// LyHTTPCommon 请求羚羊云中心 Web 的通用数据
type LyHTTPCommon struct {
	lyid   string
	lykey  string
	urlFmt string
	client *http.Client

	LastModified string `json:"last_modified"`
}

// LyHTTPResp 羚羊云中心 Web 新的返回值固定类型
type LyHTTPResp struct {
	Data      interface{} `json:"data"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
}

// Init 初始化数据
func (c *LyHTTPCommon) Init(lyid, lykey, urlFmt string) {
	c.lyid = lyid
	c.lykey = lykey
	c.urlFmt = urlFmt
	c.client = &http.Client{}
	c.LastModified = "0"
}

// Request 基于 APPID/Key 授权信息请求羚羊云中心 Web 工具函数
// 需要支持根据 `since` 参数确定响应数据
func (c *LyHTTPCommon) Request() ([]byte, error) {
	if c.client == nil {
		c.client = &http.Client{}
	}
	lastModified := c.LastModified
	if lastModified == "" {
		lastModified = "0"
	}

	var content []byte
	url := fmt.Sprintf(c.urlFmt, lastModified)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("syncer: create request %v failed: %v\n", url, err)
		return content, err
	}
	req.Header.Set("X-APP-ID", c.lyid)
	req.Header.Set("X-APP-Key", c.lykey)
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("syncer: request ly %v failed: %v\n", url, err)
		return content, err
	}

	// 读取响应内容
	content, err = ioutil.ReadAll(resp.Body)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("syncer: close response body io failed: %v\n", err)
		}
	}()
	if err != nil {
		log.Printf("syncer: read response body failed: %v\n", err)
		return content, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("syncer: request %v failed with status code: %v, response: %v\n",
			url, resp.StatusCode, string(content))
		return content, ErrResponseNotOk
	}
	return content, nil
}
