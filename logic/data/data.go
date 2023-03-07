package data

import (
	"activity/tools/log"
	"activity/tools/redis"
	"fmt"
)

const (
	ActivityRedisKey = "ACTIVITYTEST"
)

func LoadData(id int32) []byte {
	reply, err := redis.RedisExec("GET", fmt.Sprintf("%s:%d", ActivityRedisKey, id))
	if err != nil {
		log.Error("load activity data from redis error:%v", err)
		return nil
	}

	if reply == nil {
		return nil
	}

	return reply.([]byte)
}

func SaveData(id int32, data string) {
	redis.Request("SET", fmt.Sprintf("%s:%d", ActivityRedisKey, id), data)
}

func DelData(id int32) {
	redis.Request("DEL", fmt.Sprintf("%s:%d", ActivityRedisKey, id))
}
