package impl

import (
	"activity/activity"
	"github.com/golang/protobuf/proto"
	"time"
)

type IActivity interface {
	OnInit()
	OnStart()
	OnClose()
	OnSave()
	OnEvent(key string, obj activity.IPlayer, content map[string]interface{})
	Update(time.Time, int64)
	GetAward(obj activity.IPlayer, index int32)
	Format(obj activity.IPlayer) proto.Message
	OnDayReset()
}

type Def struct {
	entity activity.Entity
}

var _ IActivity = &Def{}

func (act *Def) OnInit() {}

func (act *Def) OnStart() {}

func (act *Def) OnClose() {}

func (act *Def) OnSave() {}

func (act *Def) OnEvent(key string, obj activity.IPlayer, content map[string]interface{}) {}

func (act *Def) Update(now time.Time, elspNanoSecond int64) {}

func (act *Def) GetAward(obj activity.IPlayer, index int32) {}

func (act *Def) Format(obj activity.IPlayer) proto.Message { return nil }

func (act *Def) GetCfgId() int32 { return act.entity.GetCfgId() }

func (act *Def) GetType() string { return act.entity.GetType() }

func (act *Def) GetId() int32 { return act.entity.GetId() }

func (act *Def) OnDayReset() {}
