package data

import (
	"activity/tools/log"
	"encoding/json"
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

func loadData(id int32, bindObj interface{}) bool {
	reply, err := RedisExec("GET", fmt.Sprintf("%s:%d", ActivityRedisKey, id))
	if err != nil {
		log.Error("")
		return false
	}

	if reply == nil {
		log.Error("")
		return false
	}

	err = json.Unmarshal(reply.([]byte), bindObj)
	if err != nil {
		log.Error("")
		return false
	}

	return true
}

func saveData(id int32, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		log.Error("", err)
		return
	}

	Request("SET", fmt.Sprintf("%s:%d", ActivityRedisKey, id), b)
}

func delData(id int32) {
	Request("DEL", fmt.Sprintf("%s:%d", ActivityRedisKey, id))
}
