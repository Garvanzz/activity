package logic

import (
	"activity/global"
	"activity/logic/data"
	"activity/tools/log"
)

type Entity struct {
	Id        int32            `json:"id"`
	CfgId     int32            `json:"cfg_id"`
	Type      string           `json:"type"`
	State     string           `json:"state"`
	handler   global.IActivity `json:"-"`
	StartTime int64            `json:"start_time"`
	EndTime   int64            `json:"end_time"`
	TimeType  int              `json:"time_type"`
}

func (e *Entity) isActive() bool {
	return e.State == StateRunning
}

func (e *Entity) GetCfgId() int32 { return e.CfgId }
func (e *Entity) GetType() string { return e.Type }
func (e *Entity) GetId() int32    { return e.Id }

// 加载游戏数据
func (e *Entity) load() {
	data := data.LoadData(e.Id)
	if data != "" {
		if err := e.handler.UnMarshal(data); err != nil {
			log.Error("activity handler unmarshal error")
		}
	}
}

// 保存游戏数据
func (e *Entity) save() {
	if v, err := e.handler.Marshal(); err != nil {
		log.Error("")
	} else {
		data.SaveData(e.Id, v)
	}
}
