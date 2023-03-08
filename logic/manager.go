package logic

import (
	"activity/global"
	"activity/logic/data"
	"activity/tools/fsm"
	"activity/tools/log"
	"activity/tools/redis"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"
)

var (
	instance *Manager
)

var _ global.ActivityManager = &Manager{}

type Manager struct {
	AutoId   int32
	LastTick int64
	entitys  sync.Map
	sm       *fsm.StateMachine
}

func GetInstance() global.ActivityManager {
	return instance
}

func init() {
	instance = new(Manager)
	instance.sm = fsm.NewStateMachine(&fsm.DefaultDelegate{P: instance}, transitions...)
}

func (m *Manager) Create() bool {
	reply, err := redis.RedisExec("GET", "ActivityMgr")
	if err != nil {
		log.Error("load activity manager from redis error:%v", err)
		return false
	}

	if reply != nil {
		err := json.Unmarshal(reply.([]byte), m)
		if err != nil {
			log.Error("unmarshal activity manager data error:", err)
			return false
		}
	}

	reply, err = redis.RedisExec("GET", "ActivityData")
	if err != nil {
		log.Error("load activity entity from redis error:%v", err)
		return false
	}

	entitys := make(map[int32]*entity)
	if reply != nil {
		err := json.Unmarshal(reply.([]byte), &entitys)
		if err != nil {
			log.Error("unmarshal activity entity data error:", err)
			return false
		}
	}

	// stop 重新被加载以后检查配置吗

	existIds := make(map[int32]int)
	for id, entity := range entitys {
		existIds[entity.CfgId] = 1
		m.entitys.Store(id, entity)

		// 只有运行中的活动需要加载数据
		if entity.State == StateRunning {
			handler, ok := getActivityHandler(entity)
			if !ok {
				continue
			}

			entity.handler = handler
			entity.load()
			entity.handler.OnInit()
		}

		event := entity.checkConfig()
		if event != EventNone {
			err := m.sm.Trigger(entity.State, event, entity)
			if err != nil {
				log.Error("%v", err)
				continue
			}
		}
	}

	// 根据配置加载新活动
	confs := global.AllJsons["ConfActivity"].(map[int32]global.ConfActivityElement)
	for _, conf := range confs {
		if _, ok := existIds[conf.ID]; !ok {
			m.register(conf.ID)
		}
	}

	return true
}

func (m *Manager) Stop() bool {
	entitys := make(map[int32]*entity, 0)
	m.entitys.Range(func(key, value interface{}) bool {
		entity := value.(*entity)
		entity.save()
		entitys[key.(int32)] = entity
		return true
	})

	b, err := json.Marshal(entitys)
	if err != nil {
		log.Error("activity manager stop marshal error:%v", err)
		return false
	}

	redis.RedisExec("SET", "ActivityData", string(b))

	b, err = json.Marshal(m)
	if err != nil {
		log.Error("activity manager stop marshal error1:%v", err)
		return false
	}

	redis.RedisExec("SET", "ActivityMgr", string(b))

	return true
}

func (m *Manager) Update(now time.Time, elspNanoSecond int64) {
	m.entitys.Range(func(key, value interface{}) bool {
		entity := value.(*entity)

		if entity.State == StateStopped {
			return true
		}

		event := entity.checkState()
		if event != EventNone {
			err := m.sm.Trigger(entity.State, event, entity)
			if err != nil {
				log.Error("sm trigger error:%v", err)
				return true
			}
		}

		if entity.isActive() {
			entity.handler.Update(now, elspNanoSecond)
		}

		return true
	})
}

// fsm process
func (m *Manager) OnExit(fromState string, args []interface{}) {
	e := args[0].(*entity)
	if e.State != fromState {
		log.Error("OnExit state error:%v,currentState:%v", fromState, e.State)
		return
	}
}

func (m *Manager) Action(action string, fromState string, toState string, args []interface{}) error {
	e := args[0].(*entity)

	switch action {
	case ActionStart: // waitting -> running
		log.Debug("actionStart:%v,%v", e.Id, e.CfgId)
		if handler, ok := getActivityHandler(e); ok {
			e.handler = handler
			e.handler.OnInit()
			e.handler.OnStart()
		} else {
			log.Error("activity start error")
		}
	case ActionClose:
		log.Debug("actionClose:%v,%v", e.Id, e.CfgId)
		if e.handler != nil {
			e.handler.OnClose()
			e.handler = nil
			data.DelData(e.Id) // 活动关闭清空数据
		}
	case ActionStop: // running -> stop
		log.Debug("actionStop:%v,%v", e.Id, e.CfgId)
		if fromState == StateRunning {
			e.save()
			e.handler = nil
		}
	case ActionRecover: // stop -> running
		log.Debug("actionRecover:%v,%v", e.Id, e.CfgId)
		if handler, ok := getActivityHandler(e); ok {
			e.handler = handler
			e.load()
			e.handler.OnInit()
		} else {
			log.Error("activity recover error")
		}
	case ActionRestart: // closed -> waitting
		log.Debug("actionRestart:%v,%v", e.Id, e.CfgId)
		// 分配新的id
		e.Id = m.Id()
	default:
		log.Error("unprocessed action:%v", action)
	}

	return nil
}

func (m *Manager) OnActionFailure(action string, fromState string, toState string, args []interface{}, err error) {
}

func (m *Manager) OnEnter(toState string, args []interface{}) {
	e := args[0].(*entity)
	e.State = toState
}

func (m *Manager) Id() int32 {
	return atomic.AddInt32(&m.AutoId, 1)
}

// 事件回调
func (m *Manager) OnEvent(event *global.CEvent) {
	if event == nil {
		return
	}

	if event.Obj == nil {
		return
	}

	switch event.Type {
	case global.Event_Type_PlayerOnline:
		m.notify(event.Obj, map[string]interface{}{"key": "player_online"})
	case global.Event_Type_PlayerOffline:
	case global.Event_Type_ActivityEvent:
		content, ok := event.Content.(map[string]interface{})
		if !ok {
			return
		}

		m.notify(event.Obj, content)
	}
}

// 事件分发
func (m *Manager) notify(obj global.IPlayer, content map[string]interface{}) {
	if key, ok := content["key"]; !ok {
		return
	} else {
		eventKey, ok := key.(string)
		if ok && eventKey != "" {
			m.entitys.Range(func(k, v interface{}) bool {
				entity := v.(*entity)
				if entity.isActive() {
					entity.handler.OnEvent(eventKey, obj, content)
				}
				return true
			})
		}
	}
}

// register new activity
func (m *Manager) register(cfgId int32) {
	id := m.Id()

	conf := global.GetConf(cfgId)

	var startTime, endTime int64

	if conf.ActTime == global.ActTime_CheckTime {
		if conf.StartTime == "" || conf.EndTime == "" {
			log.Error("register timer error")
			return
		}
		parseTime, err := time.ParseInLocation("2006-01-02 15:04:05", Trim(conf.StartTime), time.Local)
		if err != nil {
			log.Error("parse start time error")
			return
		}
		startTime = parseTime.Unix()

		parseTime, err = time.ParseInLocation("2006-01-02 15:04:05", Trim(conf.EndTime), time.Local)
		if err != nil {
			log.Error("parse end time error")
			return
		}
		endTime = parseTime.Unix()

		if startTime >= endTime {
			log.Error("register timer error1")
			return
		}
	}

	e := new(entity)
	e.Id = id
	e.Type = conf.Type
	e.CfgId = conf.ID
	e.StartTime = startTime
	e.EndTime = endTime
	e.TimeType = conf.ActTime

	switch e.TimeType {
	case global.ActTime_Close:
		e.State = StateClosed
	case global.ActTime_CheckTime, global.ActTime_AlwaysOpen:
		e.State = StateWaitting
	}

	m.entitys.Store(id, e)
	return
}

// TODO: API
func (m *Manager) GetAward(obj global.IPlayer, id int32, index int32) {
	v, ok := m.entitys.Load(id)
	if !ok {
		//obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityNotFound))
		//obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
		return
	}

	entity := v.(*entity)

	if !entity.isActive() {
		//obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityNotOpen))
		//obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
	}

	entity.handler.GetAward(obj, index)
}

func (m *Manager) GetActivityData(id int32, obj global.IPlayer) {
	v, ok := m.entitys.Load(id)
	if !ok {
		//obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityNotFound))
		return
	}

	entity := v.(*entity)
	if !entity.isActive() {
		//obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityNotOpen))
		return
	}

	//response := &proto_activity.ResponseActivityData{}
	//response.Id = entity.Id
	//response.ConfigId = entity.CfgId

	//data := entity.handler.Format(obj)
	//setProtoByType(entity.ActType, response, data)
	//log.Debug("GetActivityData :%v", response)
	//obj.GetConnection().Send(response)
}

func (m *Manager) GetActivityStatus(obj global.IPlayer) {
	//response := new(proto_activity.ResponseActivityStatus)
	//response.Info = make([]*proto_activity.ActivityInfo, 0)

	m.entitys.Range(func(key, value interface{}) bool {
		entity := value.(*entity)

		log.Debug("status Id:%v,CfgId:%v,State:%v,TimeType:%v,StartTime:%v,EndTime:%v,handler:%v",
			entity.Id, entity.CfgId, entity.State, entity.TimeType, entity.StartTime, entity.EndTime, entity.handler != nil)

		//if entity.isActive() {
		//这里成长礼包特殊处理
		//if entity.Type == global.ActivityType_GrowGift {
		//	pd := entity.handler.(*ActivityGrowGift).getPlayerData(obj)
		//	if pd.EndTime == 0 { // 初始化
		//		pd.init(entity.CfgId, obj)
		//	}
		//	pd.refresh(entity.CfgId)
		//	if !pd.Closed {
		//		response.Info = append(response.Info, &proto_activity.ActivityInfo{
		//			Id:        actId,
		//			ConfigId:  entity.CfgId,
		//			StartTime: entity.StartTime.Unix(),
		//			EndTime:   pd.EndTime,
		//		})
		//	}
		//	continue
		//}

		// 新手礼包特殊处理 购买过就不再显示活动
		//if entity.Type == global.ActivityType_Newcomer {
		//if obj.GetBeginnerPurchase() {
		//return true
		//}
		//}

		//var endTime int64
		//if entity.TimeType == global.ActTime_AlwaysOpen {
		//	endTime = 0
		//} else {
		//	endTime = entity.EndTime
		//}

		//response.Info = append(response.Info, &proto_activity.ActivityInfo{
		//	Id:        entity.Id,
		//	ConfigId:  entity.CfgId,
		//	StartTime: entity.StartTime.Unix(),
		//	EndTime:   endTime,
		//})
		//}
		return true
	})

	//log.Debug("GetActivityStatus :%v", response)
	//obj.GetConnection().Send(response)
}

func (m *Manager) GetActivityDataList(obj global.IPlayer) {
	//response := &proto_activity.ResponseActivityDataList{
	//	List: make([]*proto_activity.ResponseActivityData, 0),
	//}

	m.entitys.Range(func(key, value interface{}) bool {
		entity := value.(*entity)

		if !entity.isActive() {
			return true
		}

		//ret := &proto_activity.ResponseActivityData{}
		//ret.Id = entity.Id
		//ret.ConfigId = entity.CfgId
		//
		//d := entity.handler.Format(obj)
		//setProtoByType(entity.Type, ret, data)
		//response.List = append(response.List, actData)
		return true
	})

	//log.Debug("GetActivityDataList :%v", response)
	//obj.GetConnection().Send(response)
}

// 解锁战令
func (m *Manager) UnlockWarOrder(obj global.IPlayer, id int32) {
	m.entitys.Range(func(key, value interface{}) bool {
		entity := value.(*entity)
		if entity.Type == global.ActivityType_PvpWarOrder ||
			entity.Type == global.ActivityType_WarOrder ||
			entity.Type == global.ActivityType_PveWarOrder {

			if entity.isActive() {
				//TODO:unlock 战令
				//activity := entity.handler.(*ActivityWarOrder)
				//activity.unlock(obj, id)
			}
		}
		return true
	})
}

func (m *Manager) StopActivity(id int32) bool {
	v, ok := m.entitys.Load(id)
	if !ok {
		log.Debug("stop activity id error:%v", id)
		return false
	}

	entity := v.(*entity)

	if !entity.isActive() {
		return false
	}

	if err := m.sm.Trigger(entity.State, EventStop, entity); err != nil {
		log.Error("%v", err)
		return false
	}

	return true
}

func (m *Manager) RecoverActivity(id int32) bool {
	v, ok := m.entitys.Load(id)
	if !ok {
		log.Debug("recover activity id error:%v", id)
		return false
	}

	entity := v.(*entity)

	if entity.State != StateStopped {
		log.Error("recover activity state error:%v", entity.State)
		return false
	}

	err := m.sm.Trigger(entity.State, EventRecover, entity)
	if err != nil {
		log.Error("%v", err)
		return false
	}

	return true
}

func (m *Manager) DelActivity(id int32) bool {
	v, ok := m.entitys.Load(id)
	if !ok {
		log.Debug("del activity id error:%v", id)
		return false
	}

	entity := v.(*entity)

	if entity.State == StateRunning {
		err := m.sm.Trigger(entity.State, EventStop, entity)
		if err != nil {
			log.Error("%v", err)
			return false
		}
	}

	data.DelData(id)
	m.entitys.Delete(id)
	return true
}
