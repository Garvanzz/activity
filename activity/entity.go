package activity

import "activity/activity/impl"

type entity struct {
	Id      int32          `json:"id"`
	CfgId   int32          `json:"cfg_id"`
	Type    string         `json:"type"`
	State   string         `json:"state"`
	handler impl.IActivity `json:"-"`
	Time    ActivityTime   `json:"time"`
}

type ActivityTime struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
	TimeType  int   `json:"time_type"`
}

func (e *entity) isActive() bool {
	return e.State == StateRunning
}
