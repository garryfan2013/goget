package controller

import (
	"errors"

	"github.com/garryfan2013/goget/sink"
	"github.com/garryfan2013/goget/source"
)

const (
	SchemePipeline = "pipeline"
)

var (
	ctors = make(map[string]Creator)
)

func Register(c Creator) {
	ctors[c.Scheme()] = c
}

/*
	The factory interface
*/
type Creator interface {
	Create() (ProgressController, error)
	Scheme() string
}

/*
	The main interface for a controller
*/
type Controller interface {
	/*
		Open a controller
		Allocate the resources needed for the controller
	*/
	Open(src source.StreamReader, sink sink.StreamWriter, sm WorkerStatsManager, nf Notifier) error

	/*
		Method to set parameters
	*/
	SetConfig(key string, value string)

	/*
		Start the controller
		A controller couldnt be started without Openning it first
	*/
	Start() error

	/*
		Stop a controller
		When a controller is stopped, all the workers are supposed to quit
		It's supposed to initialize new workers when calling Start() again
	*/
	Stop() error

	/*
		Close a controller, responsible for releasing the related resources
	*/
	Close()
}

type Stats struct {
	Rate int64 // The stream transport speed rate for the whole job
	Size int64 // The size of data to download for the whole job
	Done int64 // The number of bytes has been downloaded for the whole job
}

type WorkerStats struct {
	Offset int64 // The offset to read data for the specified worker
	Size   int64 // The size to read for the specified worker
	Done   int64 // The number of bytes has been downloaded for the specified worker
}

const (
	NotifyEventDone = iota
)

/*
	The StatsManager interface provide methods for controller to
	retrieve and update workers' status of current job
	This interface is likely implemented by a uppper level component
	which operates and manages the controller
*/
type WorkerStatsManager interface {
	/*
		Retrieve return the stats information of current job's workers

		The returned *Stats slice would be nil, if no stats information stored
		Which means that the job probably is newly added and not started

		For job that's done or stopped, the Retrieve returns the last status
		of the job workers' information
	*/
	Retrieve() ([]*WorkerStats, int64, int64)

	/*
		Since the controller does know the exact status of current job,
		the update method provide a way for controller component to update the
		status information of job's workers
	*/
	Update([]*WorkerStats) error
}

type Notifier interface {
	/*
		Notify the upper component that some event's occured
		the event is one of the NotifyEventXxxx consts
	*/
	Notify(event int) error
}

/*
	The progresser interface provide method for upper level component
	to retrieve the job information
	This interface should be implemented by specific controller
*/
type Progresser interface {
	Progress() (*Stats, error)
}

/*
	A specific controller needs to implement both interface
*/
type ProgressController interface {
	Controller
	Progresser
}

func GetProgressController(scheme string) (ProgressController, error) {
	ctor, exists := ctors[scheme]
	if !exists {
		return nil, errors.New("Unsupported controller scheme")
	}

	return ctor.Create()
}
