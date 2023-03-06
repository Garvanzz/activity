package global

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

type CEvent struct {
	Obj     IPlayer
	Type    int
	Content interface{}
}

type ActivityManager interface {
	Create() bool                                   //创建
	Stop() bool                                     //停止
	OnEvent(event *CEvent)                          //事件回调
	GetAward(obj IPlayer, actId int32, index int32) // 领取活动奖励
	GetActivityData(activeId int32, obj IPlayer)    // 获取单个活动数据
	GetActivityDataList(obj IPlayer)                // 获取活动数据列表
	GetActivityStatus(obj IPlayer)                  // 获取活动状态
	UnlockWarOrder(obj IPlayer, id int32)           // 解锁高级战令
	StopActivity(id int32) bool                     // 暂停活动
	RecoverActivity(id int32) bool                  // 重启活动
	DelActivity(id int32) bool                      // 删除活动
	//OnRet(ret *dbengine.CDBRet)  //db返回回调
}
