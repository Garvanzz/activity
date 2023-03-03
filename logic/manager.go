package logic

import (
	"activity/global"
	"activity/logic/config"
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
	Entitys  map[int32]*Entity
	AutoId   int32
	LastTick int64
	sm       *fsm.StateMachine
	lock     sync.RWMutex // TODO:除了增添、删除活动 其他地方有无必要增加锁
}

func GetInstance() global.ActivityManager {
	return instance
}

func (m *Manager) Create() {
	instance = new(Manager)
	instance.Entitys = make(map[int32]*Entity)
	instance.sm = fsm.NewStateMachine(&fsm.DefaultDelegate{P: instance}, transitions...)

	reply, err := redis.RedisExec("GET", "activityMgr")
	if err != nil {
		log.Error("load activity manager from redis error:%v", err)
		return
	}

	if reply != nil {
		err := json.Unmarshal(reply.([]byte), m)
		if err != nil {
			log.Error("unmarshal activity manager data error:", err)
			return
		}
	}

	for _, entity := range m.Entitys {
		if handler, ok := getActivityHandler(entity); ok {
			entity.handler = handler

			// 只有运行中的活动需要加载数据
			if entity.State == StateRunning {
				entity.load()
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
	}

	// 根据配置加载新活动
	confs := make(map[int64]config.ConfActivityElement, 0)

	existIds := make(map[int32]int)
	for _, entity := range m.Entitys {
		existIds[entity.CfgId] = 1
	}

	for _, conf := range confs {
		if _, ok := existIds[conf.ID]; !ok {
			m.register(conf.ID)
		}
	}

	return
}

func (m *Manager) Stop() {
	//m.lock.Lock()
	for _, entity := range m.Entitys {
		d, err := entity.handler.Marshal()
		if err != nil {
			log.Error("")
			continue
		}

		if d != "" {
			data.SaveData(entity.Id, d)
		}
	}
	//m.lock.Unlock()

	b, err := json.Marshal(m)
	if err != nil {
		log.Error("activity manager stop marshal error:%v", err)
		return
	}

	redis.RedisExec("SET", "ActivityMgr", string(b))
}

func (m *Manager) Update(now time.Time, elspNanoSecond int64) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, entity := range m.Entitys {
		if entity.State == StateStopped {
			continue
		}

		event := entity.checkState()
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

// fsm process
func (m *Manager) OnExit(fromState string, args []interface{}) {
	e := args[0].(*Entity)
	if e.State != fromState {
		log.Error("OnExit state error:%v,currentState:%v", fromState, e.State)
		return
	}
}

func (m *Manager) Action(action string, fromState string, toState string, args []interface{}) error {
	e := args[0].(*Entity)

	switch action {
	case ActionStart: // waitting -> running
		e.handler.OnInit()
		e.handler.OnStart()
	case ActionClose:
		e.handler.OnClose()

		// clear data
		data.DelData(e.Id)
	case ActionStop:
		if fromState == StateRunning {
			// save data
		}
	case ActionRecover: // stop -> running
		d := data.LoadData(e.Id)
		if err := e.handler.UnMarshal(d); err != nil {
			log.Error("")
		}

		e.handler.OnInit()

	case ActionRestart: // closed -> waitting
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
	e := args[0].(*Entity)
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
	key, ok := content["key"]
	if !ok {
		return
	}

	eventKey, ok := key.(string)
	if ok && eventKey != "" {
		for _, entity := range m.Entitys {
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

	e := new(Entity)
	e.Id = id
	e.Type = conf.Type
	e.CfgId = conf.ID
	e.StartTime = startTime
	e.EndTime = endTime
	e.TimeType = conf.ActTime

	if handler, ok := getActivityHandler(e); ok {
		e.handler = handler
		m.Entitys[id] = e
		return
	}
}
