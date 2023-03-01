package activity

import "activity/tools/fsm"

const (
	StateWaitting = "waitting"
	StateRunning  = "running"
	StateStopped  = "stopped"
	StateClosed   = "closed"
)

const (
	EventStart   = "event_start"
	EventStop    = "event_stop"
	EventClose   = "event_close"
	EventRecover = "event_recover"
	EventRestart = "event_restart"

	ActionNone    = ""
	ActionStart   = "action_start"
	ActionClose   = "action_close"
	ActionStop    = "action_stop"
	ActionRecover = "action_recover"
	ActionRestart = "action_restart"
)

var (
	transitions = []fsm.Transition{
		{StateWaitting, EventStart, StateRunning, ActionStart},
		{StateWaitting, EventStop, StateStopped, ActionStop},
		{StateWaitting, EventClose, StateClosed, ActionClose},

		{StateRunning, EventStop, StateStopped, ActionStop},
		{StateRunning, EventClose, StateClosed, ActionClose},

		{StateStopped, EventRecover, StateRunning, ActionRecover},
		{StateStopped, EventClose, StateClosed, ActionStart},

		{StateClosed, EventRestart, StateWaitting, ActionRestart},
	}
)
