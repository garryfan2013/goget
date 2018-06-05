package manager

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"sync"

	"github.com/garryfan2013/goget/config"
	"github.com/garryfan2013/goget/sink"
	"github.com/garryfan2013/goget/source"

	"github.com/satori/go.uuid"
)

const (
	DefaultTaskCount = 5

	SchemeHttp  = "http"
	SchemeHttps = "https"
	SchemeFtp   = "ftp"
)

var (
	instance *JobManager
	once     sync.Once

	InvalidJobId    = JobId{0}
	ErrInvalidJobId = errors.New("Invalid job id")
)

type JobId uuid.UUID

func (jid JobId) String() string {
	id := uuid.UUID(jid)
	return id.String()
}

func FromString(str string) (JobId, error) {
	id, err := uuid.FromString(str)
	if err != nil {
		return InvalidJobId, err
	}

	return JobId(id), nil
}

type Job struct {
	url      string
	path     string
	userName string
	passwd   string
	src      source.StreamReader
	snk      sink.StreamWriter
	ctrl     ProgressController
}

type JobManager struct {
	jobs map[JobId]*Job
}

func GetInstance() *JobManager {
	once.Do(func() {
		instance = &JobManager{
			jobs: make(map[JobId]*Job),
		}
	})

	return instance
}

var ErrUnsupportedScheme = errors.New("Unsupported scheme")

func getStreamReader(urlStr string) (source.StreamReader, error) {
	fmtUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	ctor := source.Get(fmtUrl.Scheme)
	if ctor == nil {
		return nil, ErrUnsupportedScheme
	}

	return ctor.Create()
}

func getStreamWriter(scheme string) (sink.StreamWriter, error) {
	ctor := sink.Get(scheme)
	if ctor == nil {
		return nil, ErrUnsupportedScheme
	}

	return ctor.Create()
}

func getProgressController(scheme string) (ProgressController, error) {
	ctor := Get(scheme)
	if ctor == nil {
		return nil, ErrUnsupportedScheme
	}

	return ctor.Create()
}

func (jm *JobManager) Add(url string, filePath string, uname string, passwd string, cnt int) (JobId, error) {
	src, err := getStreamReader(url)
	if err != nil {
		return InvalidJobId, err
	}

	snk, err := getStreamWriter(sink.SchemeLocalFile)
	if err != nil {
		return InvalidJobId, err
	}

	ctrl, err := getProgressController(SchemePipeline)
	if err != nil {
		return InvalidJobId, err
	}

	if err = ctrl.Open(src, snk); err != nil {
		return InvalidJobId, err
	}

	savePath := fmt.Sprintf("%s/%s", filePath, path.Base(url))

	ctrl.SetConfig(config.KeyRemoteUrl, url)
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

	id, err := uuid.NewV4()
	if err != nil {
		return InvalidJobId, err
	}
	jobId := JobId(id)

	fmt.Println("New job Started... please wait")
	if err = ctrl.Start(); err != nil {
		return InvalidJobId, err
	}

	jm.jobs[jobId] = &Job{
		url:      url,
		path:     filePath,
		userName: uname,
		passwd:   passwd,
		src:      src,
		snk:      snk,
		ctrl:     ctrl,
	}

	return jobId, nil
}

func (jm *JobManager) Stop(id JobId) error {
	job, ok := jm.jobs[id]
	if !ok {
		return ErrInvalidJobId
	}

	return job.ctrl.Stop()
}

func (jm *JobManager) Delete(id JobId) error {
	return nil
}

func (jm *JobManager) Progress(id JobId) (*Stats, error) {
	job, ok := jm.jobs[id]
	if !ok {
		return nil, ErrInvalidJobId
	}

	return job.ctrl.Progress()
}
