package impl

import (
	"activity/global"
	"encoding/json"
	"time"

	"github.com/golang/protobuf/proto"
)

// 累计消费活动
type ActivityConsume struct {
	BaseActivity
	data *ActivityConsumeData
}

type ActivityConsumeData struct {
	StartTime time.Time
}

type ConsumePD struct {
	PdType  global.PDType   `json:"pd_type"`
	Score   int32           `json:"score"`
	GetList map[int32]int32 `json:"get_list"`
}

func (a *ActivityConsume) Format(obj global.IPlayer) proto.Message {
	// pd := a.getPlayerData(obj)

	// 如果是充值活动就每次获取的时候更新数值
	//if a.GetType() == global.ActivityType_Recharge {
	//	rechargeMoney := obj.GetRechargedMoney(2, a.data.StartTime.Unix(), time.Now().Unix())
	//	pd.Score = int32(rechargeMoney)
	//}

	//return &proto_activity.ActivityConsume{
	//Score:   pd.Score,
	//GetList: pd.GetList,
	//}

	return nil
}

func (a *ActivityConsume) OnInit() {
	if a.data == nil {
		a.data = &ActivityConsumeData{
			StartTime: time.Now(), // TODO:是否是要在这里初始化时间
		}
	}
}

func (a *ActivityConsume) OnStart() {}

func (a *ActivityConsume) OnEvent(key string, obj global.IPlayer, content map[string]interface{}) {
	switch key {
	//case "consume":
	//	score, ok := keyInt32("score", content)
	//	scoreType, ok1 := keyString("type", content)
	//	if ok && ok1 {
	//		// 钻石消费 和 玩家消费两种
	//		if scoreType == "diamond" && a.ActType == global.ActivityType_Cousume {
	//			pd := a.getPlayerData(obj)
	//			pd.Score += score
	//
	//			// 推送客户端活动积分变化
	//			obj.GetConnection().Send(&proto_activity.ResponseActivityScoreChange{
	//				Id:       a.ActId,
	//				ConfigId: a.ConfigId,
	//				RedDot:   a.RedDot(pd),
	//			})
	//		}
	//	}
	//case "recharge":
	//	if a.ActType == global.ActivityType_Recharge {
	//		pd := a.getPlayerData(obj)
	//		rechargeMoney := obj.GetRechargedMoney(2, a.data.StartTime.Unix(), time.Now().Unix())
	//		pd.Score = int32(rechargeMoney)
	//
	//		// 推送客户端活动积分变化
	//		obj.GetConnection().Send(&proto_activity.ResponseActivityScoreChange{
	//			Id:       a.ActId,
	//			ConfigId: a.ConfigId,
	//			RedDot:   a.RedDot(pd),
	//		})
	//	}
	default:
	}
}

// 目标分数变成了key 所以index=实际领取奖励需要的分数
func (a *ActivityConsume) GetAward(obj global.IPlayer, index int32) {
	conf := global.GetDataConf(a.GetCfgId())
	if conf == nil {
		//obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityConfNotFound))
		//obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
		return
	}
	data := conf.(global.ConfActivityConsume)

	// 没有该奖励配置
	awards, ok := data.Reward[index]
	if !ok {
		//obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityNoRewardToRequest))
		//obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
		return
	}

	// 没有达到目标分数
	pd := a.getPlayerData(obj)
	if pd.Score < index {
		//obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityConsumeNotEnough))
		//obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
		return
	}

	// 已经领取过奖励
	_, ok = pd.GetList[index]
	if ok {
		//obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityAlreadyRequestReward))
		//obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
		return
	}

	m := make(map[int32]uint32)
	for _, v := range awards {
		m[v.ItemID] = v.Num
	}
	//obj.AddItems(m, false)
	pd.GetList[index] = 1

	//obj.GetConnection().Send(&proto_item.PushPopReward{Items: m}) // 推送奖励弹框
	//obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: true})
}

func (a *ActivityConsume) OnClose() {}

func (a *ActivityConsume) Marshal() (string, error) {
	if data, err := json.Marshal(a.data); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

func (a *ActivityConsume) UnMarshal(data string) error {
	return json.Unmarshal([]byte(data), a.data)
}

func (a *ActivityConsume) RedDot(pd *ConsumePD) bool {
	if conf := global.GetDataConf(a.GetCfgId()); conf != nil {
		data := conf.(global.ConfActivityConsume)

		for target := range data.Reward {
			if pd.Score >= target {
				_, ok := pd.GetList[target]
				if !ok {
					return true
				}
			}
		}
	}

	return false
}

func (a *ActivityConsume) getPlayerData(obj global.IPlayer) *ConsumePD {
	pd := obj.GetActivityData(a.GetId())
	if pd == nil {
		data := &ConsumePD{
			PdType:  global.ConsumePD,
			GetList: make(map[int32]int32),
		}

		obj.SetActivityData(a.GetId(), data)
		return data
	}
	return pd.(*ConsumePD)
}
