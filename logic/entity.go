package logic

import (
	"activity/global"
	"activity/logic/data"
	"activity/tools/log"
	"time"
)

type entity struct {
	Id        int32            `json:"id"`
	CfgId     int32            `json:"cfg_id"`
	Type      string           `json:"type"`
	State     string           `json:"state"`
	StartTime int64            `json:"start_time"`
	EndTime   int64            `json:"end_time"`
	TimeType  int              `json:"time_type"`
	handler   global.IActivity `json:"-"`
}

func (e *entity) isActive() bool {
	return e.State == StateRunning
}

func (e *entity) GetCfgId() int32 { return e.CfgId }
func (e *entity) GetType() string { return e.Type }
func (e *entity) GetId() int32    { return e.Id }

// 加载游戏数据
func (e *entity) load() {
	d := data.LoadData(e.Id)
	if d != nil {
		if err := e.handler.UnMarshal(d); err != nil {
			log.Error("activity handler unmarshal error")
		}
	}
}

// 保存游戏数据
func (e *entity) save() {
	if e.handler == nil {
		return
	}

	if v, err := e.handler.Marshal(); err != nil {
		log.Error("activity handler marshal error")
	} else {
		log.Debug("entity save id:%v,cfgId:%v,data:%v", e.Id, e.CfgId, v)
		if v != "" {
			data.SaveData(e.Id, v)
		}
	}
}

// 检查活动状态
func (e *entity) checkState() (event string) {
	event = EventNone

	now := time.Now().Unix()
	switch e.State {
	case StateWaitting:
		if (now >= e.StartTime && now < e.EndTime) || e.TimeType == global.ActTime_AlwaysOpen {
			event = EventStart
		} else if now >= e.EndTime {
			event = EventClose
		}
	case StateRunning:
		if e.TimeType == global.ActTime_Close {
			event = EventClose
		} else if e.TimeType == global.ActTime_CheckTime {
			if now < e.StartTime || now > e.EndTime {
				event = EventClose
			}
		}
	case StateClosed:
		if (now >= e.StartTime && now < e.EndTime) || e.TimeType == global.ActTime_AlwaysOpen {
			event = EventRestart
		}
	}

	return
}

// 检查配置表变化
func (e *entity) checkConfig() (event string) {
	event = EventNone

	conf := global.GetConf(e.CfgId)
	if conf == nil {
		log.Error("activity config error:%v", e.CfgId)
		return
	}

	// set time type
	e.TimeType = conf.ActTime

	switch e.TimeType {
	case global.ActTime_AlwaysOpen: // 常驻活动
		e.StartTime = 0
		e.EndTime = 0

		if e.State == StateWaitting {
			event = EventStart
		} else if e.State == StateClosed {
			event = EventRestart
		}
	case global.ActTime_CheckTime: // 检查活动配置表
		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", Trim(conf.StartTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse startTime err:%v", err)
			return
		}

		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", Trim(conf.EndTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse endTime err:%v", err)
			return
		}

		if startTime.Unix() >= endTime.Unix() {
			log.Error("checkCfg startTime>=endTime err")
			return
		}

		e.StartTime = startTime.Unix()
		e.EndTime = endTime.Unix()
	case global.ActTime_Close: // 关闭活动
		e.StartTime = 0
		e.EndTime = 0

		if e.State == StateRunning || e.State == StateWaitting {
			event = EventClose
		}
	default:
		log.Error("checkCfg ActTime error:%v", conf.ActTime)
	}

	return
}
