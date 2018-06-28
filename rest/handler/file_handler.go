package handler

import (
	"net/http"
)

var (
	fileHandler = http.FileServer(http.Dir("dist"))
)

func StaticFileHandler(w http.ResponseWriter, r *http.Request, d interface{}) {
	fileHandler.ServeHTTP(w, r)
}
