package main

import (
	"net/http"

	"github.com/garryfan2013/goget/rest/route"
)

func main() {
	router := route.NewRouter()
	http.ListenAndServe("0.0.0.0:8000", router)
}
