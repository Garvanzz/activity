package main

import (
	"activity/global"
	"activity/logic"
)

type player struct {
	data map[int32]interface{}
}

func (p *player) GetActivityData(id int32) interface{} {
	return p.data[id]
}

func (p *player) SetActivityData(id int32, data interface{}) {
	p.data[id] = data
}

func main() {
	mgr := logic.GetInstance()
	mgr.Create()

	obj := new(player)
	obj.data = make(map[int32]interface{}, 0)

	mgr.OnEvent(&global.CEvent{Obj: obj, Type: global.Event_Type_ActivityEvent, Content: map[string]interface{}{
		"key": "test",
	}})

	mgr.Stop()
}
