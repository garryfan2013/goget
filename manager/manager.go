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
	Version          = "1.0.0"
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

type Job struct {
	url      string
	path     string
	userName string
	passwd   string
	src      source.StreamReader
	snk      sink.StreamWriter
	ctrl     Controller
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

func GetProtocol(urlStr string) (int, error) {
	fmtUrl, err := url.Parse(urlStr)
	if err != nil {
		return -1, err
	}

	switch fmtUrl.Scheme {
	case SchemeHttp:
	case SchemeHttps:
		return source.HttpProtocol, nil
	case SchemeFtp:
		return source.FtpProtocol, nil
	default:
	}

	return -1, fmt.Errorf("Unsupported url scheme: %s", fmtUrl)
}

func (jm *JobManager) Add(url string, filePath string, uname string, passwd string, cnt int) (JobId, error) {

	protocol, err := GetProtocol(url)
	if err != nil {
		return InvalidJobId, err
	}

	src, err := source.NewStreamReader(protocol)
	if err != nil {
		return InvalidJobId, err
	}

	snk, err := sink.NewStreamWriter(sink.LocalFileType)
	if err != nil {
		return InvalidJobId, err
	}

	ctrl, err := NewController(MultiTaskType)
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
