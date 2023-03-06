package impl

import (
	"activity/global"
	"activity/logic"
	"activity/logic/config"
	"github.com/golang/protobuf/proto"
	"time"
)

// 活跃活动
type ActivityTask struct {
	BaseActivity
}

type TaskPD struct {
	PdType      global.PDType       `json:"pd_type"`
	TaskProcess map[int32]*TaskInfo `json:"task_process"`
	GetList     map[int32]int32     `json:"get_list"`    // 阶段奖励和单个奖励放一起
	UpdateTime  *time.Time          `json:"update_time"` // 登录任务 需要记录一个时间 判断是否同一天
}

type TaskInfo struct {
	Condition1 int32 `json:"condition1"` // 目标分数
	Condition2 int32 `json:"condition2"` // 额外参数
	TaskType   int32 `json:"taskType"`   // 任务类型
	Score      int32 `json:"score"`      // 当前分数
}

func (a *ActivityTask) Format(obj global.IPlayer) proto.Message {
	pd := a.getPlayerData(obj)
	pd.loadTask(a.GetCfgId())
	//taskProcess := make(map[int32]*proto_activity.TaskInfo)
	//for id, task := range pd.TaskProcess {
	//	taskProcess[id] = &proto_activity.TaskInfo{
	//		Target: task.Condition1,
	//		Score:  task.Score,
	//	}
	//}
	//
	//return &proto_activity.ActivityTask{
	//	GetList:     pd.GetList,
	//	TaskProcess: taskProcess,
	//}

	return nil
}

func (a *ActivityTask) OnInit() {}

func (a *ActivityTask) OnStart() {}

func (a *ActivityTask) OnEvent(key string, obj global.IPlayer, content map[string]interface{}) {
	switch key {
	case "task":
		//taskType, _ := keyInt32("task_type", content)
		//extraCondition, _ := keyInt32("extra_condition", content)
		//taskCount, _ := keyInt32("task_count", content)
		//accumulate, _ := keyBool("accumulate", content)
		//pd := a.getPlayerData(obj)
		//pd.loadTask(a.ConfigId)

		// 登录任务特殊处理 TODO:object.TASK_LOGIN_X_TIMES 先写死 避免循环引用
		//if taskType == 601 {
		//	now := time.Now()
		//	if pd.UpdateTime != nil && utils.CheckIsSameDay(&now, pd.UpdateTime, 0) {
		//		return
		//	} else {
		//		pd.UpdateTime = &now
		//	}
		//}
		//
		//for _, task := range pd.TaskProcess {
		//	if task.TaskType == taskType && task.Condition2 == extraCondition {
		//		if accumulate {
		//			task.Score += taskCount
		//		} else {
		//			if taskCount > task.Score {
		//				task.Score = taskCount
		//			}
		//		}
		//	}
		//}

		// 红点推送
		//obj.GetConnection().Send(&proto_activity.ResponseActivityScoreChange{
		//	Id:       a.ActId,
		//	ConfigId: a.ConfigId,
		//	RedDot:   a.RedDot(pd),
		//})
	default:
	}
}

func (a *ActivityTask) GetAward(obj global.IPlayer, index int32) {
	//dConf := getDataConf(a.ConfigId)
	//if dConf == nil {
	//	obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityConfNotFound))
	//	obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
	//	return
	//}
	//data := dConf.(global.ConfActivityTask)
	//
	//// 没有该奖励配置
	//awards, ok := data.Rewards[index]
	//if !ok {
	//	obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityNoRewardToRequest))
	//	obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
	//	return
	//}
	//
	//pd := a.getPlayerData(obj)
	//// 已经领取过奖励
	//_, ok = pd.GetList[index]
	//if ok {
	//	obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityAlreadyRequestReward))
	//	obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
	//	return
	//}
	//
	//// 最终奖励
	//if index == data.FinalId {
	//	for _, task := range pd.TaskProcess {
	//		if task.Score < task.Condition1 {
	//			obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityConsumeNotEnough))
	//			obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
	//			return
	//		}
	//	}
	//} else {
	//	// 检查是否是任务奖励
	//	task, ok := pd.TaskProcess[index]
	//	if ok {
	//		if task.Score < task.Condition1 {
	//			obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityConsumeNotEnough))
	//			obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
	//			return
	//		}
	//	} else {
	//		// 检查是否是阶段奖励
	//		conditions, ok := data.ExtreCondition[index]
	//		if !ok {
	//			obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityNoRewardToRequest))
	//			obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
	//			return
	//		}
	//
	//		// 判断任务是否全部完成
	//		for _, id := range conditions {
	//			task := pd.TaskProcess[id]
	//			if task.Score < task.Condition1 {
	//				obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityNoRewardToRequest))
	//				obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
	//				return
	//			}
	//		}
	//	}
	//}
	//
	//m := make(map[int32]uint32)
	//for _, v := range awards {
	//	m[v.ItemID] = v.Num
	//}
	//obj.AddItems(m, false)
	//pd.GetList[index] = 1
	//
	//obj.GetConnection().Send(&proto_item.PushPopReward{Items: m}) // 推送奖励弹框
	//obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: true})
}

func (a *ActivityTask) getPlayerData(obj global.IPlayer) *TaskPD {
	pd := obj.GetActivityData(a.GetId())
	if pd == nil {
		data := &TaskPD{
			PdType:  global.TaskPD,
			GetList: make(map[int32]int32),
		}
		obj.SetActivityData(a.GetId(), data)
		return data
	}
	return pd.(*TaskPD)
}

func (a *ActivityTask) OnClose() {}

func (a *ActivityTask) RedDot(pd *TaskPD) bool {
	for id, task := range pd.TaskProcess {
		if task.Score >= task.Condition1 {
			_, ok := pd.GetList[id]
			if !ok {
				return true
			}
		}
	}
	return false
}

// 加载玩家任务
func (pd *TaskPD) loadTask(cfgId int32) {
	if len(pd.TaskProcess) > 0 {
		return
	}

	dataConf := logic.GetDataConf(cfgId)
	conf := dataConf.(config.ConfActivityTask)

	taskList := make(map[int32]*TaskInfo)
	for id, task := range conf.Tasks {
		taskList[id] = &TaskInfo{
			Condition1: int32(task.Condition1),
			Condition2: int32(task.Condition2),
			TaskType:   int32(task.TaskType),
		}
	}

	pd.TaskProcess = taskList
}
