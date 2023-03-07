package main

import (
	"activity/global"
	"activity/logic"
	"activity/tools/log"
	"activity/tools/redis"
	timer2 "activity/tools/timer"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"
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

//运行
func run(cb func()) {
	interval := time.Second

	AfterFunc(interval, cb)
}

func AfterFunc(d time.Duration, cb func()) *time.Timer {
	return time.AfterFunc(d, func() {
		cb()
	})
}

func main() {
	log.Debug("start-----------------")

	//加载配置
	global.Init()

	mgr := logic.GetInstance()
	mgr.Create()

	obj := new(player)
	obj.load()

	timer := timer2.New(time.Second, mgr.Update)
	timer.Begin()

	// event notify
	// event()

	time.Sleep(3 * time.Second)

	mgr.GetActivityStatus(nil)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
	<-c

	timer.Stop()
	mgr.Stop()
	obj.save()

	log.Debug("end-----------------")
}

func event(m global.ActivityManager, obj *player) {
	m.OnEvent(&global.CEvent{Obj: obj, Type: global.Event_Type_ActivityEvent, Content: map[string]interface{}{
		"key": "test",
	}})
}
