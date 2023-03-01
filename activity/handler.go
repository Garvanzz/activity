package activity

func getActivityHandler(actType string) IActivity {
	var handler IActivity
	//switch actType {
	//case ActivityType_Cousume, ActivityType_Recharge:
	//	handler = new(ActivityConsume)
	//case ActivityType_SpecialGift:
	//	handler = new(ActivityGift)
	//case ActivityType_GrowGift:
	//	handler = new(ActivityGrowGift)
	//case ActivityType_WarOrder, global.ActivityType_PveWarOrder, global.ActivityType_PvpWarOrder:
	//	handler = new(ActivityWarOrder)
	//case ActivityType_Newcomer:
	//	handler = new(ActivityNewcomer)
	//case ActivityType_Privilege:
	//	handler = new(ActivityPrivilege)
	//case ActivityType_CdKey:
	//	handler = new(ActivityCdkey)
	//case ActivityType_Task:
	//	handler = new(ActivityTask)
	//default:
	//	log.Error("get activity handler err:activityType=%v", actType)
	//}
	return handler
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
