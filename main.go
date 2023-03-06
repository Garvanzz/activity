package main

import (
	"activity/logic/config"
	"activity/tools/log"
	"activity/tools/redis"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
)

type player struct {
	Data map[int32]interface{}
}

func (p *player) GetActivityData(id int32) interface{} {
	return p.Data[id]
}

func (p *player) SetActivityData(id int32, data interface{}) {
	p.Data[id] = data
}

func (p *player) load() {
	p.Data = make(map[int32]interface{}, 0)

	reply, err := redis.RedisExec("GET", "PlayerTestData")
	if err != nil {
		log.Error("load player data from redis error:%v", err)
		return
	}

	if reply != nil {
		err := json.Unmarshal(reply.([]byte), &p.Data)
		if err != nil {
			log.Error("unmarshal player data error:", err)
			return
		}
	}
}

func (p *player) save() {
	b, err := json.Marshal(p.Data)
	if err != nil {
		log.Error("save player data error:%v", err)
		return
	}

	redis.RedisExec("SET", "PlayerTestData", string(b))
}

func main() {
	log.Debug("start-----------------")
	//mgr := logic.GetInstance()
	//mgr.Create()
	//
	//obj := new(player)
	//obj.load()

	//mgr.OnEvent(&global.CEvent{Obj: obj, Type: global.Event_Type_ActivityEvent, Content: map[string]interface{}{
	//	"key": "test",
	//}})

	//mgr.Stop()
	//
	//obj.save()

	// 加载配置
	config.Init()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
	<-c

	log.Debug("end-----------------")
}
