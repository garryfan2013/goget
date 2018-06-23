package store

import (
	"encoding/json"
)

var (
	jobBucketName = "JobBucket"

	ErrTypeConvert = "Cant convert type from interface{} to JobModel"
)

type JobModel struct {
	Id       string
	Url      string
	Path     string
	Username string
	Passwd   string
	Count    int
	Size     int64
	Done     int64
	Workers  []*WorkerModel
}

type WorkerModel struct {
	Offset int64
	Size   int64
	Done   int64
}

type JobStore struct {
	stor Storer
}

func NewJobStore(engine string) (*JobStore, error) {
	stor, err := GetStorer(engine)
	if err != nil {
		return nil, err
	}
	err = stor.CreateBucket(jobBucketName)
	if err != nil {
		return nil, err
	}

	return &JobStore{stor}, nil
}

func (s *JobStore) Put(j *JobModel) error {
	v, err := json.Marshal(j)
	if err != nil {
		return err
	}

	err = s.stor.Put(jobBucketName, j.Id, v)
	if err != nil {
		return err
	}

	return nil
}

func (s *JobStore) Get(id string) (*JobModel, error) {
	v, err := s.stor.Get(jobBucketName, id)
	if err != nil {
		return nil, err
	}

	var job JobModel
	err = json.Unmarshal(v, &job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (s *JobStore) GetSwap(id string, fn func(*JobModel) (*JobModel, error)) error {
	return s.stor.GetSwap(jobBucketName, id, func(val []byte) ([]byte, error) {
		var model JobModel
		err := json.Unmarshal(val, &model)
		if err != nil {
			return val, err
		}

		newModel, err := fn(&model)
		if err != nil {
			return val, err
		}

		data, err := json.Marshal(newModel)
		if err != nil {
			return val, err
		}

		return data, nil
	})
}

func (s *JobStore) Delete(id string) error {
	err := s.stor.Delete(jobBucketName, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *JobStore) ForEach(fn func(*JobModel) error) error {
	return s.stor.ForEach(jobBucketName, func(k string, v []byte) error {
		var job JobModel
		err := json.Unmarshal(v, &job)
		if err != nil {
			return err
		}

		return fn(&job)
	})
}
