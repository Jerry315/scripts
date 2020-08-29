package redis

import (
	"context"
	"errors"

	"github.com/go-redis/redis"
)

// Redis 服务类型
const (
	ServerTypeStandalone = "standalone" // 单点 Redis
	ServerTypeCluster    = "cluster"    // 集群 Redis
)

// 错误变量
var (
	ErrInvalidServerType           = errors.New("redis: invalid server type")           // 不正确的服务类型
	ErrInconsistentClusterPassword = errors.New("redis: inconsistent cluster password") // Redis 集群 URL 列表密码不一致
)

// ClientWrapper Redis 客户端封装接口类型
type ClientWrapper interface {
	ServerType() string
	RawClient() interface{}

	Context() context.Context
	Do(args ...interface{}) *redis.Cmd
	PSubscribe(channels ...string) *redis.PubSub
	Process(cmd redis.Cmder) error
	ProcessContext(ctx context.Context, cmd redis.Cmder) error
	Subscribe(channels ...string) *redis.PubSub
	Watch(fn func(*redis.Tx) error, keys ...string) error
	WatchContext(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error

	redis.Cmdable
}

// NewClient 创建 Redis 客户端, 兼容封装连接单点及集群
func NewClient(urls ...string) (ClientWrapper, error) {
	serverType := ServerTypeStandalone
	if len(urls) > 1 {
		serverType = ServerTypeCluster
	}

	srvOptions := []*redis.Options{}
	for _, url := range urls {
		options, err := redis.ParseURL(url)
		if err != nil {
			return nil, err
		}
		srvOptions = append(srvOptions, options)
	}
	// 取第一个使用, 集群使用第一个 url 的密码
	options := srvOptions[0]

	switch serverType {
	case ServerTypeStandalone:
		return &standaloneClient{redis.NewClient(options)}, nil
	case ServerTypeCluster:
		addrs := []string{}
		for _, ops := range srvOptions {
			if ops.Password != options.Password {
				return nil, ErrInconsistentClusterPassword
			}
			addrs = append(addrs, ops.Addr)
		}
		return &clusterClient{redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addrs,
			Password: options.Password,
		})}, nil
	}
	return nil, ErrInvalidServerType
}

// standaloneClient 连接单点 Redis 的客户端封装
type standaloneClient struct {
	*redis.Client
}

// ServerType 返回连接的 Redis 服务类型, 单点
func (sc *standaloneClient) ServerType() string {
	return ServerTypeStandalone
}

// RawClient 返回原始的 redis 客户端, *redis.Client
func (sc *standaloneClient) RawClient() interface{} {
	return sc.Client
}

// clusterClient 连接 Redis 集群的客户端封装
type clusterClient struct {
	*redis.ClusterClient
}

// ServerType 返回连接的 Redis 服务类型, 集群
func (cc *clusterClient) ServerType() string {
	return ServerTypeCluster
}

// RawClient 返回原始的 redis 客户端, *redis.ClusterClient
func (cc *clusterClient) RawClient() interface{} {
	return cc.ClusterClient
}
