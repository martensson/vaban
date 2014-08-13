package main

import "github.com/ant0ine/go-json-rest/rest"

func GetService(w rest.ResponseWriter, r *rest.Request) {
	service := r.PathParam("service")

	if s, ok := services[service]; ok {
		w.WriteJson(s.Hosts)
	} else {
		rest.NotFound(w, r)
		return
	}
}

func GetServices(w rest.ResponseWriter, r *rest.Request) {
	var keys []string

	for k, _ := range services {
		keys = append(keys, k)
	}
	w.WriteJson(keys)
}
