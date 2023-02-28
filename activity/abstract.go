package activity

type IActivity interface {
	OnDayReset()
	OnInit()
}


type IActivity interface {
	OnInit()                                                                // 加载活动数据 初始化数据结构
	OnStart()                                                               // 活动开启 只执行一次
	OnClose()                                                               // 活动结束 清除活动数据
	OnSave()                                                                // 保存活动数据
	OnEvent(key string, obj global.IPlayer, content map[string]interface{}) // 活动事件
	Update(time.Time, int64)                                                // 活动tick

	GetAward(obj global.IPlayer, index int32)
	Format(obj global.IPlayer) proto.Message

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

func (act *DefActivity) OnEvent(key string, obj global.IPlayer, content map[string]interface{}) {}

func (act *DefActivity) Update(now time.Time, elspNanoSecond int64) {}

func (act *DefActivity) GetAward(obj global.IPlayer, index int32) {}

func (act *DefActivity) Format(obj global.IPlayer) proto.Message { return nil }

func (act *DefActivity) SetActId(id int32) { act.ActId = id }

func (act *DefActivity) SetActType(t string) { act.ActType = t }

func (act *DefActivity) SetConfigId(id int64) { act.ConfigId = id }

func (act *DefActivity) SetRound(round int32) { act.Round = round }
