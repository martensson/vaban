package main

import (
	"net/http"

	"github.com/emicklei/go-restful"
)

func GetService(req *restful.Request, resp *restful.Response) {
	service := req.PathParameter("service")

	if s, ok := services[service]; ok {
		resp.WriteEntity(s.Hosts)
	} else {
		resp.WriteErrorString(http.StatusNotFound, "Service could not be found.")
		return
	}
}

func GetServices(req *restful.Request, resp *restful.Response) {
	var keys []string
	for k, _ := range services {
		keys = append(keys, k)
	}
	resp.WriteEntity(keys)
}
