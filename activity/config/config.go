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
