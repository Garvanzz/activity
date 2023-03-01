package activity

import (
	"activity/tools/log"
	"sync"
	"sync/atomic"
	"time"
	"activity/tools/fsm"
)

var (
	instance *Manager
)

type Manager struct {
	entitys  map[int32]*entity
	autoId   int32
	lastTick int64
	sm       *fsm.StateMachine
	lock     sync.RWMutex   // 除了增添、删除活动 其他地方有无必要增加锁
}

// 管理器生命周期函数
func (m *Manager) Create() {
	instance = new(Manager)
	instance.entitys = make(map[int32]*entity)
	instance.sm = fsm.NewStateMachine(&fsm.DefaultDelegate{P: instance}, transitions...)

	// load from redis,check new activity

	// 注册事件
}

func (m *Manager) Stop() {
	// 事件管理器 注销

	m.lock.Lock()
	for _, entity := range m.entitys {
		if entity.isActive() {
			entity.handler.OnSave()
		}
	}
	m.lock.Unlock()
}

func (m *Manager) Update(now time.Time, elspNanoSecond int64) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, entity := range m.entitys {
		if entity.State == StateClosed || entity.State == StateStopped {
			continue
		}

		action := checkActivityState(entity)
		if action != ActionNone {
			err := m.sm.Trigger(entity.State, action, entity)
			if err != nil {
				log.Error("sm trigger error:%v",err)
				continue
			}
		}

		if entity.isActive() {
			entity.handler.Update(now, elspNanoSecond)
		}
	}
}

// fms process
func (m *Manager) OnExit(fromState string, args []interface{}) {
	e := args[0].(*entity)
	if e.State != fromState {
		log.Error("")
		return
	}
}

//StateWaitting = "waitting"
//StateRunning  = "running"
//StateStopped  = "stopped"
//StateClosed   = "closed"
//
//EventStart   = "event_start"
//EventStop    = "event_stop"
//EventClose   = "event_close"
//EventRecover = "event_recover"
//EventRestart = "event_restart"

func (m *Manager) Action(action string, fromState string, toState string, args []interface{}) error {
	e := args[0].(*entity)

	switch action {
	case ActionStart: // waitting -> running
	case ActionClose:
	case ActionStop:
	case ActionRecover: // stop -> running
	case ActionRestart: // stop -> waitting
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
	return atomic.AddInt32(&m.autoId, 1)
}

// func (mgr *ActManager) Create() bool {
// 	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_PlayerOnline, mgr)  //上线事件
// 	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_PlayerOffline, mgr) //下线事件
// 	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_ActivityEvent, mgr) //活动事件

// 	conn := global.ServerG.GetDBEngine().Redis.Get()
// 	defer conn.Close()
// 	data, err := conn.Do("get", ActMgrData)
// 	if err != nil {
// 		log.Error("load actEntity from redis error:", err)
// 		return false
// 	}

// 	if data != nil {
// 		err = json.Unmarshal(data.([]byte), &mgr.actEntity)
// 		if err != nil {
// 			log.Error("load actEntity json unmarshal error:", err)
// 			return false
// 		}
// 	}

// 	reply, err := conn.Do("get", ActMgrId)
// 	if err != nil {
// 		log.Error("load actEntity id from redis error:", err)
// 		return false
// 	}
// 	if reply != nil {
// 		mgr.actId = BytesToInt(reply.([]byte))
// 		log.Debug("mgr actId:%v", mgr.actId)
// 	}

// 	// 加载已有的活动
// 	for _, entity := range mgr.actEntity {
// 		handler := getActivityHandler(entity.ActType)
// 		if handler == nil {
// 			log.Error("no activity handler:%v", entity.ActType)
// 			return false
// 		}
// 		handler.SetActType(entity.ActType)
// 		handler.SetActId(entity.ActId)
// 		handler.SetConfigId(entity.CfgId)
// 		handler.SetRound(entity.Round)

// 		entity.m = handler
// 		entity.operating(ActOperationInit)
// 		entity.checkCfg() // 检查配置表
// 	}

// 	// 加载配置中的新活动
// 	actCfgs := cfg.ConfigMgr.GetCfg("ConfActivity").(map[int64]global.ConfActivityElement)
// 	m := make(map[int64]int)
// 	for _, entity := range mgr.actEntity {
// 		m[entity.CfgId] = 1
// 	}
// 	for _, actCfg := range actCfgs {
// 		_, ok := m[actCfg.ID]
// 		if !ok {
// 			err = mgr.Register(actCfg)
// 			if err != nil {
// 				log.Error("actMgr load new activity error:%v", err)
// 			}
// 		}
// 	}

// 	mgr.Save()
// 	return true
// }

//事件回调
func (m *Manager) OnEvent(event *CEvent) {
	if event == nil {
		return
	}

	if event.Obj == nil {
		return
	}

	switch event.Type {
	//case global.Event_Type_PlayerOnline:
	//	mgr.notifyEvent(obj, map[string]interface{}{"key": "player_online"})
	//case global.Event_Type_PlayerOffline:
	case 1:
		content, ok := event.Content.(map[string]interface{})
		if !ok {
			return
		}

		m.notify(event.Obj, content)
	}
}

// 事件分发
func (m *Manager) notify(obj IPlayer, content map[string]interface{}) {
	key, ok := content["key"]
	if !ok {
		return
	}

	eventKey, ok := key.(string)
	if ok && eventKey != "" {
		m.lock.RLock()
		defer m.lock.RUnlock()
		for _, entity := range m.entitys {
			if entity.isActive() {
				entity.handler.OnEvent(eventKey, obj, content)
			}
		}
	}
}

// register new activity
func (m *Manager) register(cfgId int32) {
	id := m.Id()

	e := new(entity)

	cfg := getConf(cfgId)

	var startTime, endTime int64

	switch cfg.ActTime {
	case ActTime_AlwaysOpen: // 常驻活动
		startTime = time.Now().Unix()
	case ActTime_CheckTime: // 检查活动配置表
		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", Trim(cfg.StartTime), time.Local)
		if err != nil {
			log.Error("")
			return
		}

		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", Trim(cfg.EndTime), time.Local)
		if err != nil {
			log.Error("")
			return
		}

		if startTime.Unix() >= endTime.Unix() {
			log.Error("")
			return
		}
	case ActTime_Close: // 关闭活动
	default:
		log.Error("")
		return
	}

	handler := getActivityHandler(cfg.Type)
	if handler == nil {
		log.Error("")
		return
	}

	e.Id = id
	e.Type = cfg.Type
	e.CfgId = cfg.ID
	e.handler = handler
	e.Time = ActivityTime{StartTime: startTime, EndTime: endTime, TimeType: cfg.ActTime}

	entity.refreshStatus()
	if entity.ActTime == ActTime_AlwaysOpen {
		entity.operating(ActOperationInit)
		entity.operating(ActOperationStart)
	}

	m.lock.Lock()
	m.entitys[id] = e
	m.lock.Unlock()

}

// 检查配置表
func (e *entity) checkCfg() {
	actCfg := getConf(e.handler.GetCfgId())
	if actCfg == nil {
		log.Error("actEntity check Cfg find no cfg:%v", act.CfgId)
		return
	}

	// 暂停状态不再检查配置
	if act.Status == ActStatus_Stopped {
		return
	}

	switch actCfg.ActTime {
	case ActTime_AlwaysOpen: // 常驻活动
		if act.Status == ActStatus_NotOpen || act.Status == ActStatus_Closed {
			act.operating(ActOperationInit)
			act.operating(ActOperationStart)
		}

		act.Status = ActStatus_Running
		act.ActTime = ActTime_AlwaysOpen
	case ActTime_CheckTime: // 检查活动配置表
		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", Trim(actCfg.StartTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse startTime err:%v", err)
			return
		}

		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", Trim(actCfg.EndTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse endTime err:%v", err)
			return
		}

		if startTime.Unix() >= endTime.Unix() {
			log.Error("checkCfg startTime>=endTime err")
			return
		}

		now := time.Now().Unix()
		if act.Status == ActStatus_Running {
			act.EndTime = endTime
			if startTime.Unix() > now { // 如果新开始时间还没到就关闭活动 等待重新开启
				act.StartTime = startTime
				act.operating(ActOperationClose)
				act.Status = ActStatus_NotOpen
			}
		} else { // closed 和 not_open走这里
			act.StartTime = startTime
			act.EndTime = endTime
			act.Status = ActStatus_NotOpen
		}

		act.ActTime = ActTime_CheckTime
	case ActTime_Close: // 关闭活动
		if act.Status == ActStatus_Running {
			act.Status = ActStatus_Closed
			act.operating(ActOperationClose)
		}

		act.Status = ActStatus_Closed
		act.ActTime = ActTime_Close
	default:
		log.Error("checkCfg ActTime error:%v", actCfg.ActTime)
	}
}

func checkActivityState(e *entity) (action string) {
	action = ActionNone

	if e.State == StateStopped {
		return
	}

	if e.ActTime == ActTime_AlwaysOpen { // 常驻活动
		e.Status = ActStatus_Running
		return
	} else if e.ActTime == ActTime_Close { // 活动关闭
		e.Status = ActStatus_Closed
		return
	}

	now := time.Now().Unix()

	startTime := e.StartTime.Unix()
	endTime := e.EndTime.Unix()

	// 检查活动时间
	if e.Status == ActStatus_NotOpen || act.Status == 0 {
		if now < startTime {
			e.Status = ActStatus_NotOpen
			return
		} else if now >= startTime && now < endTime {
			e.Status = ActStatus_Running
			e.operating(ActOperationInit)
			e.operating(ActOperationStart)
			return
		} else { // 如果活动刚开启就已经结束了 不操作
			e.Status = ActStatus_Closed
			return
		}
	} else if e.Status == ActStatus_Running {
		if now > endTime { // 活动结束
			e.Status = ActStatus_Closed
			e.operating(ActOperationClose)
		}
	}
}
