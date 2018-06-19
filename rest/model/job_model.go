package model

const (
	ActionStop  = "stop"
	ActionStart = "start"
)

type Job struct {
	Id       string `json:"id,omitempty"`
	Url      string `json:"url"`
	Path     string `json:"path"`
	UserName string `json:"username,omitempty"`
	Passwd   string `json:"passwd,omitempty"`
	Count    int    `json:"count,omitempty"`
}

type Stats struct {
	Size uint64 `json:"size"`
	Done uint64 `json:"done"`
}

type JobAction struct {
	Action string `json:"action"`
}

type Jobs []Job
