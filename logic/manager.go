package logic

import (
	"activity/global"
	"activity/logic/data"
	"activity/tools/fsm"
	"activity/tools/log"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	instance *Manager
)

type Manager struct {
	Entitys  map[int32]*Entity
	AutoId   int32
	LastTick int64
	sm       *fsm.StateMachine
	lock     sync.RWMutex // 除了增添、删除活动 其他地方有无必要增加锁
}

// 管理器生命周期函数
func (m *Manager) Create() {
	instance = new(Manager)
	instance.Entitys = make(map[int32]*Entity)
	instance.sm = fsm.NewStateMachine(&fsm.DefaultDelegate{P: instance}, transitions...)

	// TODO:注册事件
	var reply []byte
	err := json.Unmarshal(reply, m)
	if err != nil {
		log.Error("unmarshal activity manager data error:", err)
		return
	}

	// TODO:load activity
	for _, entity := range m.Entitys {
		if handler, ok := getActivityHandler(entity); !ok {
			return
		} else {
			entity.handler = handler

			// load data
			entity.load()

			event := entity.checkNewCfg() // 检查配置表
			if event != EventNone {
				err := m.sm.Trigger(entity.State, event, entity)
				if err != nil {
					log.Error("")
					continue
				}
			}

		}
	}

	// 加载配置中的新活动
	actCfgs := cfg.ConfigMgr.GetCfg("ConfActivity").(map[int64]global.ConfActivityElement)
	m := make(map[int64]int)
	for _, entity := range mgr.actEntity {
		m[entity.CfgId] = 1
	}
	for _, actCfg := range actCfgs {
		_, ok := m[actCfg.ID]
		if !ok {
			err = mgr.Register(actCfg)
			if err != nil {
				log.Error("actMgr load new activity error:%v", err)
			}
		}
	}

	return
}

func (m *Manager) Stop() {
	// 事件管理器 注销

	m.lock.Lock()
	for _, entity := range m.entitys {
		// TODO:统一处理活动数据
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

		event := checkState(entity)
		if event != EventNone {
			err := m.sm.Trigger(entity.State, event, entity)
			if err != nil {
				log.Error("sm trigger error:%v", err)
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

func (m *Manager) Action(action string, fromState string, toState string, args []interface{}) error {
	e := args[0].(*entity)

	switch action {
	case ActionStart: // waitting -> running
		e.handler.OnInit()
		e.handler.OnStart()
	case ActionClose:
		e.handler.OnClose()

		// clear data
		data.DelData(e.Id)
	case ActionStop: // do nothing
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

// 事件回调
func (m *Manager) OnEvent(event *global.CEvent) {
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
func (m *Manager) notify(obj global.IPlayer, content map[string]interface{}) {
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

	conf := GetConf(cfgId)

	var startTime, endTime int64

	if conf.StartTime != "" {
		parseTime, err := time.ParseInLocation("2006-01-02 15:04:05", Trim(conf.StartTime), time.Local)
		if err != nil {
			log.Error("")
			return
		}
		startTime = parseTime.Unix()
	}

	if conf.EndTime != "" {
		parseTime, err := time.ParseInLocation("2006-01-02 15:04:05", Trim(conf.EndTime), time.Local)
		if err != nil {
			log.Error("")
			return
		}
		endTime = parseTime.Unix()
	}

	e := new(entity)
	handler := getActivityHandler(conf.Type)

	if handler == nil {
		log.Error("")
		return
	}

	e.Id = id
	e.Type = conf.Type
	e.CfgId = conf.ID
	e.handler = handler
	e.StartTime = startTime
	e.EndTime = endTime
	e.TimeType = conf.ActTime

	//m.lock.Lock()
	//m.entitys[id] = e
	//m.lock.Unlock()
}

// 检查配置表
func (e *Entity) checkNewCfg() (event string) {
	event = EventNone

	conf := GetConf(e.CfgId)
	if conf == nil {
		log.Error("activity config error:%v", e.CfgId)
		return
	}

	// set time type
	e.TimeType = conf.ActTime

	// 暂停状态不再检查配置
	if e.State == StateStopped {
		return
	}

	switch e.TimeType {
	case global.ActTime_AlwaysOpen: // 常驻活动
		if e.State == StateWaitting || e.State == StateClosed { // TODO: close 不能再被打开了 当成新活动处理
			event = EventStart
		}
	case global.ActTime_CheckTime: // 检查活动配置表
		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", Trim(conf.StartTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse startTime err:%v", err)
			return
		}

		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", Trim(conf.EndTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse endTime err:%v", err)
			return
		}

		if startTime.Unix() >= endTime.Unix() {
			log.Error("checkCfg startTime>=endTime err")
			return
		}

		now := time.Now().Unix()

		e.StartTime = startTime.Unix()
		e.EndTime = endTime.Unix()

		if e.State == StateRunning {
			if startTime.Unix() > now { // 如果新开始时间还没到就关闭活动 等待重新开启
				event = EventClose
			}
		}
		// TODO:closed 状态后续的处理
		// closed 和 not_open走这里
		//	act.StartTime = startTime
		//	act.EndTime = endTime
		//	act.Status = ActStatus_NotOpen
		//}
	case global.ActTime_Close: // 关闭活动
		if e.State == StateRunning {
			event = EventClose
		}
	default:
		log.Error("checkCfg ActTime error:%v", conf.ActTime)
	}

	return
}

func checkState(e *Entity) (event string) {
	event = EventNone

	if e.State == StateStopped {
		return
	}

	now := time.Now().Unix()

	// check time
	switch e.State {
	case StateWaitting:
		if now >= e.StartTime && now < e.EndTime {
			event = EventStart
		} else if now >= e.EndTime { // 如果活动刚开启就已经结束了 不操作
			event = EventClose
		}
	case StateRunning:
		if now > e.EndTime { // 活动正常结束
			event = EventClose
		}
	}

	return
}
