package redis

import (
	"github.com/gomodule/redigo/redis"
	"time"
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
