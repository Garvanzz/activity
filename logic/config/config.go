package config

//activity config
type ConfActivityConfig []ConfActivityElement
type ConfActivityElement struct {
	ID           int32  `json:"id"`
	Type         string `json:"type"`
	StartTime    string `json:"start_time"` //"Year-Month-Day Hour:Minute:Second"
	EndTime      string `json:"end_time"`
	ActTime      int    `json:"act_time"` // 0 一直开启\1 读取配置表时间\2 关闭活动
	DurationTime int64  `json:"duration_time"`
	RefreshType  int    `json:"refresh_type"`
}

// 战令活动
type ConfActivityWarOrder struct {
	FreeReward map[int32][]Reward `json:"award"`
	PayReward  map[int32][]Reward `json:"pay_award"`
	Pay        int32              `json:"pay"`
}

// 活跃活动
type ConfActivityTask struct {
	Tasks          map[int32]ConfActivityTaskInfo `json:"tasks"`
	Rewards        map[int32][]Reward             `json:"rewards"`
	ExtreCondition map[int32][]int32              `json:"extre_condition"`
	FinalId        int32                          `json:"final_id"`
}


type ConfActivityTaskInfo struct {
	Condition1 int64    `json:"condition1"`
	Condition2 int64    `json:"condition2"`
	TaskType   int64    `json:"taskType"`
	Reward     []Reward `json:"reward"`
}

type Reward struct {
	ItemID int32  `json:"itemId"`
	Num    uint32 `json:"num"`
}

// 累计消耗活动
type ConfActivityConsume struct {
	Reward map[int32][]Reward `json:"award"`
}