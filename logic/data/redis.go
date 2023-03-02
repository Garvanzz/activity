package data

import (
	"activity/tools/log"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

const (
	ActivityRedisKey = "ACTIVITY"
)

var (
	pool *redis.Pool
)

func init() {
	pool = &redis.Pool{
		MaxIdle:     200,
		MaxActive:   2000,
		IdleTimeout: 60 * 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
}

func RedisExec(cmd string, args ...interface{}) (reply interface{}, err error) {
	conn := pool.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
}

func Request(cmd string, args ...interface{}) {
	go func() {
		conn := pool.Get()
		defer conn.Close()
		conn.Do(cmd, args...)
	}()
}

func LoadData(id int32) string {
	reply, err := RedisExec("GET", fmt.Sprintf("%s:%d", ActivityRedisKey, id))
	if err != nil {
		log.Error("load activity data from redis error:%v", err)
		return ""
	}

	if reply == nil {
		return ""
	}

	return reply.(string)
}

func SaveData(id int32, data string) {
	Request("SET", fmt.Sprintf("%s:%d", ActivityRedisKey, id), data)
}

func DelData(id int32) {
	Request("DEL", fmt.Sprintf("%s:%d", ActivityRedisKey, id))
}
