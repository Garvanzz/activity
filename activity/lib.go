package activity

import (
	"activity/activity/config"
	"strings"
)

type CEvent struct {
	Obj     IPlayer
	Type    int
	Content interface{}
}

// get activity config by id
func getConf(id int32) *config.ConfActivityElement {
	return &config.ConfActivityElement{}
}

//func getConf(configId int64) *global.ConfActivityElement {
//	actConf, ok := cfg.ConfigMgr.GetCfg("ConfActivity").(map[int64]global.ConfActivityElement)[configId]
//	if !ok {
//		return nil
//	}
//
//	return &actConf
//}
//
//func getDataConf(configId int64) interface{} {
//	actDataConf, ok := cfg.ConfigMgr.GetCfg("ConfActivityData").(map[int64]interface{})[configId]
//	if !ok {
//		return nil
//	}
//
//	return actDataConf
//}

func Trim(s string) string {
	return strings.Trim(s, "\"")
}
