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

const (
	ActivityType_Newcomer    = "ConfPlayerGift"  // 新手礼包活动
	ActivityType_Privilege   = "ConfPrivilege"   // 特权活动
	ActivityType_WarOrder    = "ConfWarOrder"    // 最强指挥官
	ActivityType_SpecialGift = "ConfSpecialGift" // 礼包类活动
	ActivityType_Recharge    = "ConfComRecharge" // 充值类活动
	ActivityType_Cousume     = "ConfConsume"     // 累计积分活动
	ActivityType_PveWarOrder = "ConfPveOrder"    // Pve战令
	ActivityType_PvpWarOrder = "ConfPvpOrder"    // Pvp战令
	ActivityType_GrowGift    = "ConfGrowGift"    // 成长礼包
	ActivityType_CdKey       = "ConfCdkeyGift"   // CdKey活动
	ActivityType_Task        = "ConfLiveness"    // 活跃活动
)

// player impl
type IPlayer interface {
	GetActivityData(id int32) interface{}
	SetActivityData(id int32, data interface{})
}

type IActivity interface {
	OnInit()  // 每次加载完成都会调用一次
	OnStart() // 只会调用一次
	OnClose() // 活动结束调用
	Marshal() (string, error)
	UnMarshal(data []byte) error
	OnEvent(key string, obj IPlayer, content map[string]interface{})
	Update(time.Time, int64)
	Format(obj IPlayer) proto.Message
	GetAward(obj IPlayer, index int32)
	OnDayReset()
}
