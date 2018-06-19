package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/garryfan2013/goget/proxy"
	_ "github.com/garryfan2013/goget/proxy/local"
	_ "github.com/garryfan2013/goget/proxy/rpc_proxy"
	"github.com/garryfan2013/goget/rest/model"
)

var (
	pm proxy.ProxyManager
)

func init() {
	pm, _ = proxy.GetProxyManager(proxy.ProxyRPC)
	if pm == nil {
		panic("Cant get proxy manager for RPC")
	}
}

func GetJobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

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

func GetJobsHandler(w http.ResponseWriter, r *http.Request) {
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

func PostJobHandler(w http.ResponseWriter, r *http.Request) {
	var job model.Job
	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
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

func PostJobActionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var action model.JobAction
	err := json.NewDecoder(r.Body).Decode(&action)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if action.Action == model.ActionStop {
		err = pm.Stop(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func GetJobProgressHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	nativeStat, err := pm.Progress(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(&model.Stats{
		Size: uint64(nativeStat.Size),
		Done: uint64(nativeStat.Done),
	})
	if err != nil {
		fmt.Println(err)
	}
}

func DeleteJobHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := pm.Delete(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
