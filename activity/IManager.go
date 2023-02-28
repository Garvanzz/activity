package activity

const (
	ActivityType_Newcomer    = "ConfPlayerGift"  // 新手礼包活动
	ActivityType_Privilege   = "ConfPrivilege"   // 特权活动
	ActivityType_WarOrder    = "ConfWarOrder"    // 最强指挥官
	ActivityType_SpecialGift = "ConfSpecialGift" // 礼包类活动
	ActivityType_Recharge    = "ConfComRecharge" // 充值类活动
	ActivityType_Cousume     = "ConfConsume"     // 累计积分活动
	ActivityType_PveWarOrder = "ConfPveOrder"    // Pve战令
	ActivityType_PvpWarOrder = "ConfPvpOrder"    // Pvp战令
	ActivityType_GrowGift    = "ConfGrowGift"    //成长礼包
	ActivityType_CdKey       = "ConfCdkeyGift"   // CdKey活动
	ActivityType_Task        = "ConfLiveness"    // 活跃活动
)

type PDType int32

// 活动玩家数据类型
const (
	WarOrderPD PDType = iota + 1
	TaskPD
	GrowGiftPD
	GiftPD
	ConsumePD
)

type Manager interface {
	Create() bool //创建
	Stop() bool   //停止

	OnEvent(event *event.CEvent) //事件回调
	OnRet(ret *dbengine.CDBRet)  //db返回回调

	GetAward(obj IPlayer, actId int32, index int32) // 领取活动奖励
	GetActivityData(activeId int32, obj IPlayer)    // 获取单个活动数据
	GetActivityDataList(obj IPlayer)                // 获取活动数据列表
	GetActivityStatus(obj IPlayer)
	UnlockWarOrder(obj IPlayer, id int32) // 解锁高级战令

	// 改变活动状态
	StopActivity(actId int32)
	RestartActivity(actId int32)
	DelActivity(actId int32)
}
