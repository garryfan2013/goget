package model

const (
	ActionStop  = "stop"
	ActionStart = "start"
)

const (
	StatusStopped = iota
	StatusRunning
	StatusDone
)

type Job struct {
	Id       string `json:"id,omitempty"`
	Url      string `json:"url"`
	Path     string `json:"path"`
	UserName string `json:"username,omitempty"`
	Passwd   string `json:"passwd,omitempty"`
	Count    int    `json:"count,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Done     int64  `json:"done,omitempty"`
	Status   int    `json:"status,omitempty"`
}

type Stats struct {
	Rate   uint64 `json:"rate"`
	Status int64  `json:"status"`
	Size   uint64 `json:"size"`
	Done   uint64 `json:"done"`
}

type JobAction struct {
	Action string `json:"action"`
}

type Jobs []Job
