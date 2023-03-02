package logic

import (
	"activity/logic/config"
	"strings"
)

// get activity config by id
func GetConf(id int32) *config.ConfActivityElement {
	return &config.ConfActivityElement{}
}

//func GetConf(configId int64) *global.ConfActivityElement {
//	actConf, ok := cfg.ConfigMgr.GetCfg("ConfActivity").(map[int64]global.ConfActivityElement)[configId]
//	if !ok {
//		return nil
//	}
//
//	return &actConf
//}

func GetDataConf(configId int32) interface{} {
	//actDataConf, ok := cfg.ConfigMgr.GetCfg("ConfActivityData").(map[int64]interface{})[configId]
	//if !ok {
	//	return nil
	//}
	//
	//return actDataConf

	return nil
}

func Trim(s string) string {
	return strings.Trim(s, "\"")
}
