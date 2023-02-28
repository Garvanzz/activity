package activity

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"qiniupkg.com/x/log.v7"
)

var (
	instance *Manager
)

type Manager struct {
}

func (a *Manager) OnStart() {
}

func (a *Manager) AfterStart() {
}

func (a *Manager) OnStop() {
}

func (a *Manager) AfterStop() {
}

type ActStatus uint8

const (
	ActStatus_NotOpen ActStatus = iota + 1 // 未到活动开启时间
	ActStatus_Running
	ActStatus_Stopped
	ActStatus_Closed
)

// 配置表时间类型
const (
	ActTime_Close      = iota // 关闭活动
	ActTime_AlwaysOpen        // 常驻活动
	ActTime_CheckTime         // 配置表时间
)

var _ global.ActMgr = &ActManager{}

var ActMgr *ActManager

func init() {
	ActMgr = new(ActManager)
	ActMgr.actEntity = make(map[int32]*ActEntity)
}

type ActManager struct {
	actEntity    map[int32]*ActEntity
	actId        int32
	lastSaveTime int64
	m            sync.RWMutex
}

func (mgr *ActManager) Create() bool {
	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_PlayerOnline, mgr)  //上线事件
	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_PlayerOffline, mgr) //下线事件
	global.ServerG.GetEventRouter().AddEventListener(global.Event_Type_ActivityEvent, mgr) //活动事件

	conn := global.ServerG.GetDBEngine().Redis.Get()
	defer conn.Close()
	data, err := conn.Do("get", ActMgrData)
	if err != nil {
		log.Error("load actEntity from redis error:", err)
		return false
	}

	if data != nil {
		err = json.Unmarshal(data.([]byte), &mgr.actEntity)
		if err != nil {
			log.Error("load actEntity json unmarshal error:", err)
			return false
		}
	}

	reply, err := conn.Do("get", ActMgrId)
	if err != nil {
		log.Error("load actEntity id from redis error:", err)
		return false
	}
	if reply != nil {
		mgr.actId = BytesToInt(reply.([]byte))
		log.Debug("mgr actId:%v", mgr.actId)
	}

	// 加载已有的活动
	for _, entity := range mgr.actEntity {
		handler := getActivityHandler(entity.ActType)
		if handler == nil {
			log.Error("no activity handler:%v", entity.ActType)
			return false
		}
		handler.SetActType(entity.ActType)
		handler.SetActId(entity.ActId)
		handler.SetConfigId(entity.CfgId)
		handler.SetRound(entity.Round)

		entity.m = handler
		entity.operating(ActOperationInit)
		entity.checkCfg() // 检查配置表
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

	mgr.Save()
	return true
}

func (mgr *ActManager) Stop() bool {
	global.ServerG.GetEventRouter().DelEventListener(global.Event_Type_PlayerOnline, mgr)  //上线事件
	global.ServerG.GetEventRouter().DelEventListener(global.Event_Type_PlayerOffline, mgr) //下线事件
	global.ServerG.GetEventRouter().DelEventListener(global.Event_Type_ActivityEvent, mgr) //活动事件

	return mgr.Save()
}

func (mgr *ActManager) Save() bool {
	conn := global.ServerG.GetDBEngine().Redis.Get()
	defer conn.Close()
	data, err := json.Marshal(mgr.actEntity)
	if err != nil {
		log.Error("actManager stop marshal active data error:", err)
		return false
	}

	_, err = conn.Do("set", ActMgrData, data)
	if err != nil {
		log.Error("actManager stop set active data error:", err)
		return false
	}

	_, err = conn.Do("set", ActMgrId, IntToBytes(mgr.actId))
	if err != nil {
		log.Error("actManager stop set active data error:", err)
		return false
	}

	mgr.m.Lock()
	for _, entity := range mgr.actEntity {
		if entity.isActive() {
			entity.m.OnSave()
		}
	}
	mgr.m.Unlock()

	return true
}

func (mgr *ActManager) Update(now time.Time, elspNanoSecond int64) {
	// 每隔1分钟保存一次数据库
	if mgr.lastSaveTime == 0 {
		mgr.lastSaveTime = now.Unix()
	} else {
		if now.Unix()-mgr.lastSaveTime > 300 {
			mgr.lastSaveTime = now.Unix()
			mgr.Save()
		}
	}

	mgr.m.RLock()
	defer mgr.m.RUnlock()
	for _, entity := range mgr.actEntity {
		if entity.Status == ActStatus_Closed || entity.Status == ActStatus_Stopped {
			continue
		}

		entity.refreshStatus()
		if entity.isActive() {
			entity.m.Update(now, elspNanoSecond)
		}
	}
}

func (mgr *ActManager) CreateId() int32 {
	atomic.AddInt32(&mgr.actId, 1)
	return mgr.actId
}

//事件回调
func (mgr *ActManager) OnEvent(event *event.CEvent) {
	if event == nil {
		return
	}

	if event.Obj == nil {
		return
	}

	obj, ok := event.Obj.(global.IPlayer)
	if !ok {
		return
	}

	switch event.Type {
	case global.Event_Type_PlayerOnline:
		mgr.notifyEvent(obj, map[string]interface{}{"key": "player_online"})
	case global.Event_Type_PlayerOffline:
	case global.Event_Type_ActivityEvent:
		var content map[string]interface{}
		m, ok := event.Content.(map[string]interface{})
		if ok {
			content = m
		} else {
			content = make(map[string]interface{})
		}

		mgr.notifyEvent(obj, content)
	}
}

func (mgr *ActManager) notifyEvent(obj global.IPlayer, content map[string]interface{}) {
	key, ok := content["key"]
	if ok {
		eventKey, ok := key.(string)
		if ok && eventKey != "" {
			mgr.m.RLock()
			defer mgr.m.RUnlock()
			for _, entity := range mgr.actEntity {
				if entity.isActive() {
					entity.m.OnEvent(eventKey, obj, content)
				}
			}
		}
	}
}

// redis 回调
func (mgr *ActManager) OnRet(ret *dbengine.CDBRet) {}

func (mgr *ActManager) Register(actCfg global.ConfActivityElement) error {
	actId := mgr.CreateId()
	entity, err := CreateEntity(actCfg, actId) // 关闭活动类型
	if err != nil {
		return fmt.Errorf("register create entity error:%v", err)
	}

	entity.refreshStatus()
	if entity.ActTime == ActTime_AlwaysOpen {
		entity.operating(ActOperationInit)
		entity.operating(ActOperationStart)
	}

	mgr.m.Lock()
	mgr.actEntity[actId] = entity
	mgr.m.Unlock()

	return nil
}

func CreateEntity(cfg global.ConfActivityElement, actId int32) (*ActEntity, error) {
	entity := new(ActEntity)

	var (
		startTime time.Time
		endTime   time.Time
		err       error
	)

	switch cfg.ActTime {
	case ActTime_AlwaysOpen: // 常驻活动
		startTime = time.Now()
	case ActTime_CheckTime: // 检查活动配置表
		startTime, err = time.ParseInLocation("2006-01-02 15:04:05", Trim(cfg.StartTime), time.Local)
		if err != nil {
			return nil, fmt.Errorf("CreateAct parse startTime err:%v", err)
		}

		endTime, err = time.ParseInLocation("2006-01-02 15:04:05", Trim(cfg.EndTime), time.Local)
		if err != nil {
			return nil, fmt.Errorf("CreateAct parse endTime err:%v", err)
		}

		if startTime.Unix() >= endTime.Unix() {
			return nil, fmt.Errorf("CreateAct startTime>=endTime err")
		}
	case ActTime_Close: // 关闭活动
	default:
		return nil, fmt.Errorf("CreateAct ActTime error:%v", cfg.ActTime)
	}

	entity.ActId = actId
	entity.ActType = cfg.Type
	entity.CfgId = cfg.ID
	entity.StartTime = startTime
	entity.EndTime = endTime
	entity.ActTime = cfg.ActTime

	handler := getActivityHandler(cfg.Type)
	if handler == nil {
		return nil, fmt.Errorf("CreateAct get active handler nil,id:%v", cfg.Type)
	}
	handler.SetActType(entity.ActType)
	handler.SetActId(entity.ActId)
	handler.SetConfigId(entity.CfgId)
	entity.m = handler

	return entity, nil
}

// 获取活动状态列表
func (mgr *ActManager) GetActivityStatus(obj global.IPlayer) {
	response := new(proto_activity.ResponseActivityStatus)
	response.Info = make([]*proto_activity.ActivityInfo, 0)

	mgr.m.RLock()
	defer mgr.m.RUnlock()
	for actId, entity := range mgr.actEntity {
		if entity.isActive() { // 只用给正在开启的活动

			//这里成长礼包特殊处理
			if entity.ActType == global.ActivityType_GrowGift {
				pd := entity.m.(*ActivityGrowGift).getPlayerData(obj)
				if pd.EndTime == 0 { // 初始化
					pd.init(entity.CfgId, obj)
				}
				pd.refresh(entity.CfgId)
				if !pd.Closed {
					response.Info = append(response.Info, &proto_activity.ActivityInfo{
						Id:        actId,
						ConfigId:  entity.CfgId,
						StartTime: entity.StartTime.Unix(),
						EndTime:   pd.EndTime,
					})
				}
				continue
			}

			// 新手礼包特殊处理 购买过就不再显示活动
			if entity.ActType == global.ActivityType_Newcomer {
				if obj.GetBeginnerPurchase() {
					continue
				}
			}

			var endTime int64
			if entity.ActTime == ActTime_AlwaysOpen {
				endTime = 0
			} else {
				endTime = entity.EndTime.Unix()
			}

			response.Info = append(response.Info, &proto_activity.ActivityInfo{
				Id:        actId,
				ConfigId:  entity.CfgId,
				StartTime: entity.StartTime.Unix(),
				EndTime:   endTime,
			})
		}
	}

	//log.Debug("GetActivityStatus :%v", response)
	obj.GetConnection().Send(response)
}

// 获取单个活动数据
func (mgr *ActManager) GetActivityData(activeId int32, obj global.IPlayer) {
	mgr.m.RLock()
	entity, ok := mgr.actEntity[activeId]
	mgr.m.RUnlock()

	if !ok {
		obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityNotFound))
		return
	}

	if !entity.isActive() {
		obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityNotOpen))
		return
	}

	response := &proto_activity.ResponseActivityData{}
	response.Id = entity.ActId
	response.ConfigId = entity.CfgId

	data := entity.m.Format(obj)
	setProtoByType(entity.ActType, response, data)
	//log.Debug("GetActivityData :%v", response)
	obj.GetConnection().Send(response)
}

// 获取活动数据列表
func (mgr *ActManager) GetActivityDataList(obj global.IPlayer) {
	response := &proto_activity.ResponseActivityDataList{
		List: make([]*proto_activity.ResponseActivityData, 0),
	}

	mgr.m.RLock()
	defer mgr.m.RUnlock()

	for _, entity := range mgr.actEntity {
		if !entity.isActive() {
			return
		}

		actData := &proto_activity.ResponseActivityData{}
		actData.Id = entity.ActId
		actData.ConfigId = entity.CfgId

		data := entity.m.Format(obj)
		setProtoByType(entity.ActType, actData, data)
		response.List = append(response.List, actData)
	}
	//log.Debug("GetActivityDataList :%v", response)
	obj.GetConnection().Send(response)
}

// 领取活动奖励
func (mgr *ActManager) GetAward(obj global.IPlayer, actId, index int32) {
	mgr.m.RLock()
	entity, ok := mgr.actEntity[actId]
	mgr.m.RUnlock()
	if !ok {
		obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityNotFound))
		obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
		return
	}

	if entity.isActive() {
		entity.m.GetAward(obj, index)
	} else {
		obj.GetConnection().SendError(int(proto_base.ErrorCode_ActivityNotOpen))
		obj.GetConnection().Send(&proto_activity.ResponseActivityAward{Success: false})
	}
}

// 解锁战令
func (mgr *ActManager) UnlockWarOrder(obj global.IPlayer, id int32) {
	mgr.m.RLock()
	for _, entity := range mgr.actEntity {
		if entity.ActType == global.ActivityType_PvpWarOrder ||
			entity.ActType == global.ActivityType_WarOrder ||
			entity.ActType == global.ActivityType_PveWarOrder {

			if entity.isActive() {
				activity := entity.m.(*ActivityWarOrder)
				activity.unlock(obj, id)
			}
		}
	}
	mgr.m.RUnlock()
}

// 暂停活动
func (mgr *ActManager) StopActivity(actId int32) {
	mgr.m.RLock()
	entity, ok := mgr.actEntity[actId]
	mgr.m.RUnlock()
	if !ok {
		log.Debug("stop activity err:find no activity %v", actId)
		return
	}

	if entity.isActive() {
		entity.Status = ActStatus_Stopped
		entity.m.OnSave()
	}
}

// 重启活动
func (mgr *ActManager) RestartActivity(actId int32) {
	mgr.m.RLock()
	defer mgr.m.RUnlock()
	entity, ok := mgr.actEntity[actId]
	if !ok {
		log.Debug("restart activity err:find no activity %v", actId)
		return
	}

	if entity.Status == ActStatus_Stopped {
		now := time.Now().Unix()
		if entity.ActTime == ActTime_CheckTime && entity.EndTime.Unix() <= now {
			entity.Status = ActStatus_Closed
			entity.m.OnClose()
		} else {
			entity.Status = ActStatus_Running
		}
	} else {
		log.Debug("restart activity err:activity status is not stop %v", actId)
	}
}

// 删除活动
func (mgr *ActManager) DelActivity(actId int32) {
	mgr.m.RLock()
	entity, ok := mgr.actEntity[actId]
	mgr.m.RUnlock()
	if !ok {
		log.Debug("del activity err:find no activity %v", actId)
		return
	}

	if entity.Status == ActStatus_Running {
		entity.operating(ActOperationClose)
	}

	mgr.m.Lock()
	delete(mgr.actEntity, actId)
	mgr.m.Unlock()
}

//======================================================================================
type ActOperation int

const (
	ActOperationInit ActOperation = iota
	ActOperationStart
	ActOperationClose
)

type ActEntity struct {
	m IActivity

	ActType string `json:"act_type"`
	ActId   int32  `json:"act_id"`
	Round   int32  `json:"round"`
	CfgId   int64  `json:"cfg_id"`

	Status    ActStatus `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	ActTime   int       `json:"act_time"` // 活动时间类型
}

func (act *ActEntity) isActive() bool {
	return act.Status == ActStatus_Running
}

// 刷新活动状态
func (act *ActEntity) refreshStatus() {
	// stop 不需要刷新
	if act.Status == ActStatus_Stopped {
		return
	}

	if act.ActTime == ActTime_AlwaysOpen { // 常驻活动
		act.Status = ActStatus_Running
		return
	} else if act.ActTime == ActTime_Close { // 活动关闭
		act.Status = ActStatus_Closed
		return
	}

	now := time.Now().Unix()

	startTime := act.StartTime.Unix()
	endTime := act.EndTime.Unix()

	// 检查活动时间
	if act.Status == ActStatus_NotOpen || act.Status == 0 {
		if now < startTime {
			act.Status = ActStatus_NotOpen
			return
		} else if now >= startTime && now < endTime {
			act.Status = ActStatus_Running
			act.operating(ActOperationInit)
			act.operating(ActOperationStart)
			return
		} else { // 如果活动刚开启就已经结束了 不操作
			act.Status = ActStatus_Closed
			return
		}
	} else if act.Status == ActStatus_Running {
		if now > endTime { // 活动结束
			act.Status = ActStatus_Closed
			act.operating(ActOperationClose)
		}
	}
}

// 活动操作
func (act *ActEntity) operating(operation ActOperation) {
	switch operation {
	case ActOperationInit:
		log.Debug("活动Init：%v,%v,%v", act.ActId, act.ActType, act.CfgId)
		act.m.OnInit()
	case ActOperationStart:
		log.Debug("活动Start：%v,%v,%v", act.ActId, act.ActType, act.CfgId)
		act.m.OnStart()
		act.Round++
		act.m.SetRound(act.Round)
	case ActOperationClose:
		log.Debug("活动Close：%v,%v,%v", act.ActId, act.ActType, act.CfgId)
		act.m.OnClose()
		delActivityData(act.ActId)
	default:
	}
}

// 检查配置表
func (act *ActEntity) checkCfg() {
	actCfg := getConf(act.CfgId)
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
