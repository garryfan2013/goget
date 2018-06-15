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
	"github.com/garryfan2013/goget/util"

	_ "github.com/garryfan2013/goget/controller/pipeline"
	_ "github.com/garryfan2013/goget/sink/file"
	_ "github.com/garryfan2013/goget/source/ftp"
	_ "github.com/garryfan2013/goget/source/http"
)

const (
	DefaultTaskCount = 5
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
	src      source.StreamReader
	snk      sink.StreamWriter
	ctrl     controller.ProgressController
}

type JobManager struct {
	jobs map[string]*Job
	lock sync.RWMutex
}

func getInstance() *JobManager {
	once.Do(func() {
		instance = &JobManager{
			jobs: make(map[string]*Job),
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
	Interface proxy.ProxyManager implementation
*/
func (jm *JobManager) Get(id string) (*proxy.JobInfo, error) {
	jm.lock.RLock()
	defer jm.lock.RUnlock()

	if j, exists := jm.jobs[id]; exists {
		return &proxy.JobInfo{id, j.url, j.path}, nil
	}

	return nil, errors.New("Job not exists")
}

func (jm *JobManager) GetAll() ([]*proxy.JobInfo, error) {
	jm.lock.RLock()
	defer jm.lock.RUnlock()

	jobs := make([]*proxy.JobInfo, 0, len(jm.jobs))
	for id, job := range jm.jobs {
		jobs = append(jobs,
			&proxy.JobInfo{
				Id:   id,
				Url:  job.url,
				Path: job.path,
			})
	}

	return jobs, nil
}

func (jm *JobManager) Add(urlStr string, filePath string, uname string, passwd string, cnt int) (*proxy.JobInfo, error) {
	fmtUrl, err := url.Parse(urlStr)
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

	if err = ctrl.Open(src, snk); err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			ctrl.Close()
		}
	}()

	savePath := fmt.Sprintf("%s/%s", filePath, path.Base(urlStr))

	ctrl.SetConfig(config.KeyRemoteUrl, urlStr)
	ctrl.SetConfig(config.KeyLocalPath, savePath)

	if cnt <= 0 {
		cnt = DefaultTaskCount
	}
	ctrl.SetConfig(config.KeyTaskCount, fmt.Sprintf("%d", cnt))

	if uname != "" {
		ctrl.SetConfig(config.KeyUserName, uname)
	}
	if passwd != "" {
		ctrl.SetConfig(config.KeyPasswd, passwd)
	}

	id, err := util.UUIDStringGen()
	if err != nil {
		return nil, err
	}

	err = ctrl.Start()
	if err != nil {
		return nil, err
	}

	jm.lock.Lock()
	defer jm.lock.Unlock()

	if _, exists := jm.jobs[id]; exists {
		panic("It's not supposed to have job existed with same uuid as the lately generated")
	}

	jm.jobs[id] = &Job{
		url:      urlStr,
		path:     filePath,
		userName: uname,
		passwd:   passwd,
		src:      src,
		snk:      snk,
		ctrl:     ctrl,
	}

	return &proxy.JobInfo{
		Id:   id,
		Url:  urlStr,
		Path: filePath,
	}, nil
}

func (jm *JobManager) Stop(id string) error {
	jm.lock.RLock()
	defer jm.lock.RUnlock()

	job, ok := jm.jobs[id]
	if !ok {
		return ErrJobNotExist
	}

	return job.ctrl.Stop()
}

func (jm *JobManager) Delete(id string) error {
	jm.lock.Lock()
	defer jm.lock.Unlock()

	job, ok := jm.jobs[id]
	if !ok {
		return ErrJobNotExist
	}

	job.ctrl.Close()
	delete(jm.jobs, id)
	return nil
}

func (jm *JobManager) Progress(id string) (*proxy.Stats, error) {
	jm.lock.RLock()
	defer jm.lock.RUnlock()

	job, ok := jm.jobs[id]
	if !ok {
		return nil, ErrJobNotExist
	}

	stats, err := job.ctrl.Progress()
	if err != nil {
		return nil, err
	}

	return &proxy.Stats{
		Size: stats.Size,
		Done: stats.Done,
	}, nil
}
