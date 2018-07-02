package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/garryfan2013/goget/proxy"
	"github.com/garryfan2013/goget/rest/model"
)

const (
	BadInterfaceValue = "Interface cant convert to proxy.ProxyManager"
)

func GetJobHandler(w http.ResponseWriter, r *http.Request, d interface{}) {
	vars := mux.Vars(r)
	id := vars["id"]

	pm, ok := d.(proxy.ProxyManager)
	if !ok {
		panic(BadInterfaceValue)
	}

	job, err := pm.Get(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data, err := json.Marshal(&model.Job{
		Id:   job.Id,
		Url:  job.Url,
		Path: job.Path,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func GetJobsHandler(w http.ResponseWriter, r *http.Request, d interface{}) {
	pm, ok := d.(proxy.ProxyManager)
	if !ok {
		panic(BadInterfaceValue)
	}

	nativeJobs, err := pm.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jobs := make([]model.Job, len(nativeJobs))
	for i := 0; i < len(nativeJobs); i++ {
		jobs[i].Id = nativeJobs[i].Id
		jobs[i].Url = nativeJobs[i].Url
		jobs[i].Path = nativeJobs[i].Path
	}

	err = json.NewEncoder(w).Encode(jobs)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func PostJobHandler(w http.ResponseWriter, r *http.Request, d interface{}) {
	var job model.Job
	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pm, ok := d.(proxy.ProxyManager)
	if !ok {
		panic(BadInterfaceValue)
	}

	nativeJob, err := pm.Add(job.Url, job.Path, job.UserName, job.Passwd, job.Count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(&model.Job{
		Id:   nativeJob.Id,
		Url:  nativeJob.Url,
		Path: nativeJob.Path,
	})
	if err != nil {
		fmt.Println(err)
	}
}

func PostJobActionHandler(w http.ResponseWriter, r *http.Request, d interface{}) {
	vars := mux.Vars(r)
	id := vars["id"]

	var action model.JobAction
	err := json.NewDecoder(r.Body).Decode(&action)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pm, ok := d.(proxy.ProxyManager)
	if !ok {
		panic(BadInterfaceValue)
	}

	if action.Action == model.ActionStop {
		err = pm.Stop(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if action.Action == model.ActionStart {
		err = pm.Start(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func GetJobProgressHandler(w http.ResponseWriter, r *http.Request, d interface{}) {
	vars := mux.Vars(r)
	id := vars["id"]

	pm, ok := d.(proxy.ProxyManager)
	if !ok {
		panic(BadInterfaceValue)
	}

	nativeStat, err := pm.Progress(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(&model.Stats{
		Rate:   uint64(nativeStat.Rate),
		Status: int64(nativeStat.Status),
		Size:   uint64(nativeStat.Size),
		Done:   uint64(nativeStat.Done),
	})
	if err != nil {
		fmt.Println(err)
	}
}

func DeleteJobHandler(w http.ResponseWriter, r *http.Request, d interface{}) {
	vars := mux.Vars(r)
	id := vars["id"]

	pm, ok := d.(proxy.ProxyManager)
	if !ok {
		panic(BadInterfaceValue)
	}

	err := pm.Delete(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
