package logic

import (
	"activity/global"
	"activity/logic/impl"
	"activity/tools/log"
)

func getActivityHandler(entity *entity) (global.IActivity, bool) {
	var handler global.IActivity
	switch entity.Type {
	case global.ActivityType_Cousume:
		h := new(impl.ActivityConsume)
		h.BaseInfo = entity
		handler = h
	case global.ActivityType_Task:
		h := new(impl.ActivityTask)
		h.BaseInfo = entity
		handler = h
	default:
		log.Error("unknown activity type:%v", entity.Type)
		return nil, false
	}

	return handler, true
}

// func setProtoByType(actType string, msg *proto_activity.ResponseActivityData, data proto.Message) {
// 	switch actType {
// 	case global.ActivityType_Cousume, global.ActivityType_Recharge:
// 		nd := data.(*proto_activity.ActivityConsume)
// 		msg.ActivityConsume = nd
// 	case global.ActivityType_SpecialGift, global.ActivityType_GrowGift:
// 		nd := data.(*proto_activity.ActivityGift)
// 		msg.ActivityGift = nd
// 	case global.ActivityType_WarOrder, global.ActivityType_PveWarOrder, global.ActivityType_PvpWarOrder:
// 		nd := data.(*proto_activity.ActivityWarOrder)
// 		msg.ActivityWarOrder = nd
// 	case global.ActivityType_Newcomer:
// 		nd := data.(*proto_activity.ActivityNewcomer)
// 		msg.ActivityNewcomer = nd
// 	case global.ActivityType_Privilege:
// 		nd := data.(*proto_activity.ActivityPrivilege)
// 		msg.ActivityPrivilege = nd
// 	case global.ActivityType_CdKey:
// 		nd := data.(*proto_activity.ActivityCdKey)
// 		msg.ActivityCdKey = nd
// 	case global.ActivityType_Task:
// 		nd := data.(*proto_activity.ActivityTask)
// 		msg.ActivityTask = nd
// 	default:
// 		log.Error("set proto by type error:%v", actType)
// 	}
// }
