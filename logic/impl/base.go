package impl

import (
	"activity/global"
	"github.com/golang/protobuf/proto"
	"time"
)

var _ global.IActivity = &BaseActivity{}

type BaseInfo interface {
	GetCfgId() int32
	GetType() string
	GetId() int32
}

type BaseActivity struct {
	BaseInfo
}

func (base *BaseActivity) OnInit() {}

func (base *BaseActivity) OnStart() {}

func (base *BaseActivity) OnClose() {}

func (base *BaseActivity) Marshal() (string, error) { return "", nil }

func (base *BaseActivity) UnMarshal(data string) error { return nil }

func (base *BaseActivity) OnEvent(key string, obj global.IPlayer, content map[string]interface{}) {}

func (base *BaseActivity) Update(now time.Time, elspNanoSecond int64) {}

func (base *BaseActivity) Format(obj global.IPlayer) proto.Message { return nil }

func (base *BaseActivity) OnDayReset() {}

func (base *BaseActivity) GetAward(obj global.IPlayer, index int32) {}
