package redis_test

import (
	"fmt"

	"git.topvdn.com/web/goantelope/redis"
)

func ExampleClientWrapper() {
	url := "redis://:@127.0.0.1:6379/0"
	urls := []string{
		"redis://:123456@127.0.0.1:7000/0",
		"redis://:123456@127.0.0.1:7001/0",
		"redis://:123456@127.0.0.1:7002/0",
		"redis://:123456@127.0.0.1:7003/0",
		"redis://:123456@127.0.0.1:7004/0",
		"redis://:123456@127.0.0.1:7005/0",
	}

	key1 := "test_key1"
	value1 := "test_value1"
	key2 := "test_key2"
	value2 := "test_value2"

	// 连接单点的 Redis 服务
	standaloneClient, err := redis.NewClient(url)
	if err != nil {
		panic(err)
	}
	_, err = standaloneClient.Do("set", key1, value1).Result()
	if err != nil {
		panic(err)
	}
	get_value1, err := standaloneClient.Do("get", key1).String()
	fmt.Println(get_value1)
	_, err = standaloneClient.Del(key1).Result()
	if err != nil {
		panic(err)
	}

	// 连接集群的 Redis 服务
	clusterClient, err := redis.NewClient(urls...)
	if err != nil {
		panic(err)
	}
	_, err = clusterClient.Do("set", key2, value2).Result()
	if err != nil {
		panic(err)
	}
	get_value2, err := clusterClient.Do("get", key2).String()
	fmt.Println(get_value2)
	_, err = clusterClient.Del(key2).Result()
	if err != nil {
		panic(err)
	}

	// Output:
	// test_value1
	// test_value2
}
