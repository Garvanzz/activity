package logic

import "activity/tools/fsm"

const (
	// state
	StateWaitting = "waitting"
	StateRunning  = "running"
	StateStopped  = "stopped"
	StateClosed   = "closed"

	// event
	EventNone    = ""
	EventStart   = "event_start"
	EventStop    = "event_stop"
	EventClose   = "event_close"
	EventRecover = "event_recover"
	EventRestart = "event_restart"

	// action
	ActionStart   = "action_start"
	ActionClose   = "action_close"
	ActionStop    = "action_stop"
	ActionRecover = "action_recover"
	ActionRestart = "action_restart"
)

var (
	transitions = []fsm.Transition{
		{StateWaitting, EventStart, StateRunning, ActionStart},
		{StateWaitting, EventClose, StateClosed, ActionClose}, // 等待开启中的活动 时间修改了 就不再处理了
		{StateRunning, EventStop, StateStopped, ActionStop},
		{StateRunning, EventClose, StateClosed, ActionClose},
		{StateStopped, EventRecover, StateRunning, ActionRecover},
		{StateClosed, EventRestart, StateWaitting, ActionRestart},
	}
)
