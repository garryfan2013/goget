package local

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"sync"

	"github.com/garryfan2013/goget/config"
	"github.com/garryfan2013/goget/controller"
	"github.com/garryfan2013/goget/proxy"
	"github.com/garryfan2013/goget/sink"
	"github.com/garryfan2013/goget/source"
	"github.com/garryfan2013/goget/store"
	"github.com/garryfan2013/goget/util"

	_ "github.com/garryfan2013/goget/controller/pipeline"
	_ "github.com/garryfan2013/goget/sink/file"
	_ "github.com/garryfan2013/goget/source/ftp"
	_ "github.com/garryfan2013/goget/source/http"
	_ "github.com/garryfan2013/goget/store/engine"
)

const (
	DefaultTaskCount = 5
)

const (
	StatusStopped = iota
	StatusRunning
	StatusDone
)

var (
	instance *JobManager
	once     sync.Once

	ErrJobNotExist       = errors.New("Job with specified id not exists")
	ErrUnsupportedScheme = errors.New("Unsupported scheme")
)

/*
	Register the builder when this package's imported by others
*/
func init() {
	proxy.Register(&JobManagerBuilder{})
}

type Job struct {
	url      string
	path     string
	userName string
	passwd   string

	src  source.StreamReader
	snk  sink.StreamWriter
	ctrl controller.ProgressController
}

type JobDescriptor struct {
	Status int
	Ctrl   controller.ProgressController
}

type JobManager struct {
	jds   map[string]*JobDescriptor
	jstor *store.JobStore
	lock  sync.RWMutex
}

func getInstance() *JobManager {
	once.Do(func() {
		stor, err := store.NewJobStore(store.StoreBoltDB)
		if err != nil {
			panic("Create JobStore failed")
		}

		instance = &JobManager{
			jds:   make(map[string]*JobDescriptor),
			jstor: stor,
		}

		err = instance.load()
		if err != nil {
			panic("Load JobStore failed")
		}
	})

	return instance
}

type JobManagerBuilder struct{}

/*
	Interface proxy.Builder implementation
*/
func (jmb *JobManagerBuilder) Build() (proxy.ProxyManager, error) {
	return getInstance(), nil
}

func (jmb *JobManagerBuilder) Name() string {
	return proxy.ProxyLocal
}

/*
	Load jobModel from JobStore, and initialize each job
*/
func (jm *JobManager) load() error {
	err := jm.jstor.ForEach(func(model *store.JobModel) error {
		jm.jds[model.Id] = &JobDescriptor{
			Status: StatusStopped,
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

/*
	Interface proxy.ProxyManager implementation
*/
func (jm *JobManager) Get(id string) (*proxy.JobInfo, error) {
	jm.lock.RLock()
	defer jm.lock.RUnlock()

	_, exists := jm.jds[id]
	if !exists {
		return nil, proxy.ErrJobNotExists
	}

	model, err := jm.jstor.Get(id)
	if err != nil {
		return nil, err
	}

	return &proxy.JobInfo{
		Id:   id,
		Url:  model.Url,
		Path: model.Path,
	}, nil
}

func (jm *JobManager) GetAll() ([]*proxy.JobInfo, error) {
	jm.lock.RLock()
	defer jm.lock.RUnlock()

	jobs := make([]*proxy.JobInfo, 0, len(jm.jds))
	err := jm.jstor.ForEach(func(model *store.JobModel) error {
		jobs = append(jobs,
			&proxy.JobInfo{
				Id:   model.Id,
				Url:  model.Url,
				Path: model.Path,
			})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func (jm *JobManager) start(model *store.JobModel) (controller.ProgressController, error) {
	fmtUrl, err := url.Parse(model.Url)
	if err != nil {
		return nil, err
	}

	src, err := source.GetStreamReader(fmtUrl.Scheme)
	if err != nil {
		return nil, err
	}

	snk, err := sink.GetStreamWriter(sink.SchemeLocalFile)
	if err != nil {
		return nil, err
	}

	ctrl, err := controller.GetProgressController(controller.SchemePipeline)
	if err != nil {
		return nil, err
	}

	if err = ctrl.Open(src, snk, &JobStatsManager{
		id: model.Id,
		jm: jm,
	}); err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			ctrl.Close()
		}
	}()

	savePath := fmt.Sprintf("%s/%s", model.Path, path.Base(model.Url))

	ctrl.SetConfig(config.KeyRemoteUrl, model.Url)
	ctrl.SetConfig(config.KeyLocalPath, savePath)
	ctrl.SetConfig(config.KeyTaskCount, fmt.Sprintf("%d", model.Count))

	if model.Username != "" {
		ctrl.SetConfig(config.KeyUserName, model.Username)
	}
	if model.Passwd != "" {
		ctrl.SetConfig(config.KeyPasswd, model.Passwd)
	}

	err = ctrl.Start()
	if err != nil {
		return nil, err
	}

	return ctrl, nil
}

func (jm *JobManager) Start(id string) error {
	jm.lock.Lock()
	defer jm.lock.Unlock()

	jd, exists := jm.jds[id]
	if !exists {
		return proxy.ErrJobNotExists
	}

	if jd.Status == StatusRunning {
		return nil
	}

	model, err := jm.jstor.Get(id)
	if err != nil {
		return err
	}

	ctrl, err := jm.start(model)
	if err != nil {
		return err
	}

	jd.Status = StatusRunning
	jd.Ctrl = ctrl
	return nil
}

func (jm *JobManager) Add(urlStr string, filePath string, uname string, passwd string, cnt int) (*proxy.JobInfo, error) {
	jm.lock.Lock()
	defer jm.lock.Unlock()

	id, err := util.UUIDStringGen()
	if err != nil {
		return nil, err
	}

	if cnt <= 0 {
		cnt = DefaultTaskCount
	}

	err = jm.jstor.Put(&store.JobModel{
		Id:       id,
		Url:      urlStr,
		Path:     filePath,
		Username: uname,
		Passwd:   passwd,
		Count:    cnt,
	})
	if err != nil {
		return nil, err
	}

	jm.jds[id] = &JobDescriptor{
		Status: StatusStopped,
	}

	return &proxy.JobInfo{
		Id:   id,
		Url:  urlStr,
		Path: filePath,
	}, nil
}

func (jm *JobManager) Stop(id string) error {
	jm.lock.Lock()
	defer jm.lock.Unlock()

	jd, exists := jm.jds[id]
	if !exists {
		return ErrJobNotExist
	}

	if jd.Status == StatusRunning {
		err := jd.Ctrl.Stop()
		if err != nil {
			return err
		}

		jd.Status = StatusStopped
	}

	return nil
}

func (jm *JobManager) Delete(id string) error {
	jm.lock.Lock()
	defer jm.lock.Unlock()

	jd, exists := jm.jds[id]
	if !exists {
		return ErrJobNotExist
	}

	jd.Ctrl.Close()
	err := jm.jstor.Delete(id)
	if err != nil {
		return err
	}

	delete(jm.jds, id)
	return nil
}

func (jm *JobManager) done(id string) error {
	jm.lock.Lock()
	defer jm.lock.Unlock()

	jd, exists := jm.jds[id]
	if !exists {
		return ErrJobNotExist
	}

	jd.Status = StatusDone
	return nil
}

func (jm *JobManager) Progress(id string) (*proxy.Stats, error) {
	jm.lock.RLock()
	defer jm.lock.RUnlock()

	jd, exists := jm.jds[id]
	if !exists {
		return nil, ErrJobNotExist
	}

	if jd.Status != StatusRunning {
		model, err := jm.jstor.Get(id)
		if err != nil {
			return nil, err
		}

		/*
			Probably the job has just been added but not started yet
		*/
		if model.Workers == nil {
			return &proxy.Stats{
				Size: 0,
				Done: 0,
			}, nil
		}

		/*
			In case the job's done or stopped for now
		*/
		var total, done int64
		for _, v := range model.Workers {
			total += v.Size
			done += v.Done
		}
		return &proxy.Stats{
			Size: total,
			Done: done,
		}, nil
	}

	/*
		For job now is running
	*/
	stats, err := jd.Ctrl.Progress()
	if err != nil {
		return nil, err
	}

	return &proxy.Stats{
		Size: stats.Size,
		Done: stats.Done,
	}, nil
}

type JobStatsManager struct {
	id string
	jm *JobManager
}

/*
	Interface controller.StatsManager implementation
*/
func (jsm *JobStatsManager) Retrieve() ([]*controller.Stats, int64, int64) {
	model, err := jsm.jm.jstor.Get(jsm.id)
	if err != nil {
		return nil, 0, 0
	}

	if len(model.Workers) == 0 {
		return nil, 0, 0
	}

	stats := make([]*controller.Stats, len(model.Workers))
	for i, v := range model.Workers {
		stats[i] = &controller.Stats{
			Offset: v.Offset,
			Size:   v.Size,
			Done:   v.Done,
		}
	}

	return stats, int64(model.Size), int64(model.Done)
}

func (jsm *JobStatsManager) Update(stats []*controller.Stats) error {
	workers := make([]*store.WorkerModel, len(stats))
	for i, v := range stats {
		workers[i] = &store.WorkerModel{
			Offset: v.Offset,
			Size:   v.Size,
			Done:   v.Done,
		}
	}

	return jsm.jm.jstor.GetSwap(jsm.id, func(m *store.JobModel) (*store.JobModel, error) {
		m.Workers = workers
		return m, nil
	})
}

func (jsm *JobStatsManager) Notify(event int) error {
	switch event {
	case controller.NotifyEventDone:
		return jsm.jm.done(jsm.id)
	default:
		fmt.Printf("Unkown notify event %d\n", event)
		return errors.New("Unkown notify event")
	}
}
