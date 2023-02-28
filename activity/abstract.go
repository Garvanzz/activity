package activity

import (
	"time"

	"google.golang.org/protobuf/proto"
)

// player impl
type IPlayer interface {
}

type IActivity interface {
	OnInit()
	OnStart()
	OnClose()
	OnSave()
	OnEvent(key string, obj IPlayer, content map[string]interface{})
	Update(time.Time, int64)

	GetAward(obj IPlayer, index int32)
	Format(obj IPlayer) proto.Message

	OnDayReset()

	SetActId(id int32)
	SetActType(t string)
	SetConfigId(id int64)
	SetRound(round int32)
}

var _ IActivity = &DefActivity{}

type DefActivity struct {
	ActId    int32
	Round    int32
	ActType  string
	ConfigId int64
}

func (act *DefActivity) OnInit() {}

func (act *DefActivity) OnStart() {}

func (act *DefActivity) OnClose() {}

func (act *DefActivity) OnSave() {}

func (act *DefActivity) OnEvent(key string, obj IPlayer, content map[string]interface{}) {}

func (act *DefActivity) Update(now time.Time, elspNanoSecond int64) {}

func (act *DefActivity) GetAward(obj IPlayer, index int32) {}

func (act *DefActivity) Format(obj IPlayer) proto.Message { return nil }

func (act *DefActivity) SetActId(id int32) { act.ActId = id }

func (act *DefActivity) SetActType(t string) { act.ActType = t }

func (act *DefActivity) SetConfigId(id int64) { act.ConfigId = id }

func (act *DefActivity) SetRound(round int32) { act.Round = round }
