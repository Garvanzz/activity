package global

import (
	"github.com/golang/protobuf/proto"
	"time"
)

type PDType int32

// 活动玩家数据类型
const (
	ConsumePD PDType = iota + 1
	TaskPD
)

// 配置表时间类型
const (
	ActTime_Close      = iota // 关闭活动
	ActTime_AlwaysOpen        // 常驻活动
	ActTime_CheckTime         // 配置表时间
)

const (
	Event_Type_PlayerOnline = iota + 1
	Event_Type_PlayerOffline
	Event_Type_ActivityEvent
)

// player impl
type IPlayer interface {
	GetActivityData(id int32) interface{}
	SetActivityData(id int32, data interface{})
}

type IActivity interface {
	OnInit()
	OnStart()
	OnClose()
	Marshal() (string, error)
	UnMarshal(data string) error
	OnEvent(key string, obj IPlayer, content map[string]interface{})
	Update(time.Time, int64)
	Format(obj IPlayer) proto.Message
	GetAward(obj IPlayer, index int32)
	OnDayReset()
}
