package logic

import (
	"activity/logic/config"
	"strings"
)

func GetConf(configId int32) *config.ConfActivityElement {
	actConf, ok := config.AllJsons["ConfActivity"].(map[int32]config.ConfActivityElement)[configId]
	if !ok {
		return nil
	}

	return &actConf
}

func GetDataConf(configId int32) interface{} {
	actDataConf, ok := config.AllJsons["ConfActivity"].(map[int32]interface{})[configId]
	if !ok {
		return nil
	}

	return actDataConf
}

func Trim(s string) string {
	return strings.Trim(s, "\"")
}
