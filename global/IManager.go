package global

import "time"

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
	Update(now time.Time, elspNanoSecond int64)     // tick

	//OnRet(ret *dbengine.CDBRet)  //db返回回调
}
