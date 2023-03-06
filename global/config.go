package global

import (
	"activity/tools/log"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

//activity config
type ConfActivity []ConfActivityElement
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

// 消耗类活动配置条目(用于解析配置)
type ActivityConsumeObj struct {
	Activity int32    `json:"activity"`
	Id       int32    `json:"id"`
	Num      int32    `json:"num"`
	Reward   []Reward `json:"reward"`
}

// 活跃活动配置条目(用于解析配置)
type ActivityTaskObj struct {
	Activity   int32    `json:"activity"`
	Id         int32    `json:"id"`
	Condition1 int64    `json:"condition1"`
	Condition2 int64    `json:"condition2"`
	Type       int64    `json:"type"`
	Reward     []Reward `json:"reward"`
	Row        int32    `json:"row"`
	Column     int32    `json:"column"`
}

var (
	AllJsons map[string]interface{}
)

func Init() {
	AllJsons = make(map[string]interface{})

	jsonName := "ConfActivity"
	obj := &ConfActivity{}

	jsonPath, err := os.Getwd()
	if err != nil {
		log.Error("os wd:%v", err)
		return
	}

	data, err := ioutil.ReadFile(fmt.Sprintf("%s\\json\\%s.json", jsonPath, jsonName))
	if err != nil {
		log.Fatal("%v", err)
	}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		log.Fatal("%v", err)
	}
	idconfMap := make(map[int32]ConfActivityElement)
	for _, v := range *obj {
		idconfMap[v.ID] = v
	}
	AllJsons[jsonName] = idconfMap

	m := make(map[int32]interface{})
	for _, activity := range idconfMap {
		switch activity.Type {
		case ActivityType_Cousume:
			data, _ := ioutil.ReadFile(fmt.Sprintf("%s\\json\\%s.json", jsonPath, activity.Type))
			list := make([]*ActivityConsumeObj, 0)
			err := json.Unmarshal(data, &list)
			if err != nil {
				log.Fatal("3:%v", err)
				continue
			}

			actIds := make(map[int32]int)
			for _, info := range list {
				actIds[info.Activity] = 1
			}
			for activityId := range actIds {
				obj := ConfActivityConsume{
					Reward: make(map[int32][]Reward),
				}

				for _, info := range list {
					if info.Activity == activityId {
						obj.Reward[info.Num] = info.Reward
					}
				}

				m[activityId] = obj
			}
		case ActivityType_Task:
			data, _ := ioutil.ReadFile(fmt.Sprintf("%s\\json\\%s.json", jsonPath, activity.Type))
			list := make([]*ActivityTaskObj, 0)
			err := json.Unmarshal(data, &list)
			if err != nil {
				log.Fatal("%v", err)
				continue
			}

			actIds := make(map[int32]int)
			for _, info := range list {
				actIds[info.Activity] = 1
			}
			for activityId := range actIds {
				obj := ConfActivityTask{
					Tasks:          make(map[int32]ConfActivityTaskInfo),
					Rewards:        make(map[int32][]Reward),
					ExtreCondition: make(map[int32][]int32),
				}

				matrix := [4][4]int32{}
				for _, info := range list {
					if info.Activity == activityId {
						if info.Type != 0 {
							obj.Tasks[info.Id] = ConfActivityTaskInfo{
								Condition1: info.Condition1,
								Condition2: info.Condition2,
								TaskType:   info.Type,
								Reward:     info.Reward,
							}

							matrix[info.Row-1][info.Column-1] = info.Id
							obj.Rewards[info.Id] = info.Reward
						}
					}
				}

				for _, info := range list {
					if info.Activity == activityId {
						if info.Type == 0 {
							// 最终奖励
							if info.Row == info.Column {
								obj.Rewards[info.Id] = info.Reward
								obj.FinalId = info.Id
								continue
							}

							obj.Rewards[info.Id] = info.Reward

							cs := make([]int32, 0)
							if info.Row > info.Column {
								for i := info.Row - 1; i > 0; i-- {
									cs = append(cs, matrix[i-1][info.Column-1])
								}
							} else {
								for i := info.Column - 1; i > 0; i-- {
									cs = append(cs, matrix[info.Row-1][i-1])
								}
							}

							obj.ExtreCondition[info.Id] = cs
						}
					}
				}

				m[activityId] = obj
			}
		default:
		}
	}

	AllJsons["ConfActivityData"] = m
}

func GetConf(configId int32) *ConfActivityElement {
	actConf, ok := AllJsons["ConfActivity"].(map[int32]ConfActivityElement)[configId]
	if !ok {
		return nil
	}

	return &actConf
}

func GetDataConf(configId int32) interface{} {
	actDataConf, ok := AllJsons["ConfActivityData"].(map[int32]interface{})[configId]
	if !ok {
		return nil
	}

	return actDataConf
}
