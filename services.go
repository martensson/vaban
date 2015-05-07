package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func GetService(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	service := ps.ByName("service")

	if s, ok := services[service]; ok {
		r.JSON(w, http.StatusOK, s.Hosts)
		return
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Service could not be found."))
		return
	}
}

func GetServices(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var keys []string
	for k, _ := range services {
		keys = append(keys, k)
	}
	r.JSON(w, http.StatusOK, keys)
}
