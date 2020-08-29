package handler

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

func Handle(address, password string, db int, logger *zap.Logger) (client redis.Conn, err error) {
	dn := redis.DialDatabase(db)
	if password != "" {
		pd := redis.DialPassword(password)
		client, err = redis.Dial("tcp", address, pd, dn)
	} else {
		client, err = redis.Dial("tcp", address, dn)
	}
	if err != nil {
		logger.Error("create redis handle failed.")
		return nil, err
	}
	return
}

func GetMemory(conn redis.Conn, logger *zap.Logger) (err error){
	memory, err := redis.String(conn.Do("info", "Memory"))
	if err != nil {
		logger.Error("get redis memory info failed")
		return
	}
	items := strings.Split(memory, "\n")
	for _, item := range items {
		if strings.Contains(item, "used_memory:") {
			usedMemory := strings.Split(item, ":")[1]
			fmt.Println(usedMemory)
		}
	}
	return
}

func GetCpu(conn redis.Conn, logger *zap.Logger) (err error){
	cpuInfo, err := redis.String(conn.Do("info", "CPU"))
	if err != nil {
		logger.Error("get redis cpu info failed")
		return
	}
	for _, item := range strings.Split(cpuInfo, "\n") {
		if strings.Contains(item, "used_cpu_sys:") {
			fmt.Println(strings.Split(item, ":")[1])
		}
	}
	return
}

func GetDb(conn redis.Conn, db int, logger *zap.Logger) (err error){
	dbInfo, err := redis.String(conn.Do("info", "Keyspace"))
	if err != nil {
		logger.Error("get redis db info failed")
		return
	}
	for _, item := range strings.Split(dbInfo, "\n") {
		if strings.Contains(item, "db"+strconv.Itoa(db)) {
			fmt.Println(strings.Split(strings.Split(strings.Split(item, ":")[1], ",")[0], "=")[1])
		}
	}
	return
}
