package main

import (
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/emicklei/go-restful"
)

type HealthStatus struct {
	Refs  string
	Admin string
	Probe string
}
type Backends map[string]HealthStatus
type Servers map[string]Backends

type HealthPost struct {
	Set_health string
}

func GetHealth(req *restful.Request, resp *restful.Response) {
	service := req.PathParameter("service")
	backend := req.PathParameter("backend")
	healthpost := HealthPost{}
	req.ReadEntity(&healthpost)
	if s, ok := services[service]; ok {
		// We need the WaitGroup for some awesome Go concurrency
		var wg sync.WaitGroup
		servers := Servers{}
		for _, server := range s.Hosts {
			// Increment the WaitGroup counter.
			wg.Add(1)
			go func(server string) {
				// Decrement the counter when the goroutine completes.
				defer wg.Done()
				backends := Backends{}
				backends = StatusHealth(server, s.Secret, backend)
				servers[server] = backends
			}(server)
		}
		wg.Wait()
		resp.WriteEntity(servers)
	} else {
		resp.WriteErrorString(http.StatusNotFound, "Service could not be found.")
		return
	}
}

func PostHealth(req *restful.Request, resp *restful.Response) {
	service := req.PathParameter("service")
	backend := req.PathParameter("backend")
	healthpost := HealthPost{}
	err := req.ReadEntity(&healthpost)
	if err != nil {
		resp.AddHeader("Content-Type", "text/plain")
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	if healthpost.Set_health == "" {
		resp.WriteErrorString(http.StatusBadRequest, "Set_health is required")
		return
	}
	if s, ok := services[service]; ok {
		// We need the WaitGroup for some awesome Go concurrency
		var wg sync.WaitGroup
		messages := Messages{}
		for _, server := range s.Hosts {
			// Increment the WaitGroup counter.
			wg.Add(1)
			go func(server string) {
				// Decrement the counter when the goroutine completes.
				defer wg.Done()
				message := Message{}
				message.Msg = UpdateHealth(server, s.Secret, backend, healthpost)
				messages[server] = message
			}(server)
		}
		wg.Wait()
		resp.WriteEntity(messages)
	} else {
		resp.WriteErrorString(http.StatusNotFound, "Service could not be found.")
		return
	}
}

func UpdateHealth(server string, secret string, backend string, healthpost HealthPost) string {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Println(err)
		return err.Error()
	}
	defer conn.Close()
	err = varnishAuth(server, secret, conn)
	if err != nil {
		log.Println(err)
	}
	conn.Write([]byte("backend.set_health " + backend + " " + healthpost.Set_health + "\n"))
	// again, 64 bytes is enough for this.
	byte_status := make([]byte, 64)
	_, err = conn.Read(byte_status)
	if err != nil {
		log.Printf("Could not read packet : %s", err.Error())
		return err.Error()
	}
	// cast byte to string and only keep the status code (always max 13 char), the rest we dont care.
	status := string(byte_status)[0:12]
	status = strings.Trim(status, " ")
	log.Println(server, "set_health", backend, healthpost.Set_health, "status", status)
	return "updated with status " + status
}

func StatusHealth(server string, secret string, backend string) Backends {
	backends := Backends{}
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Println(err)
		return backends
	}
	defer conn.Close()
	err = varnishAuth(server, secret, conn)
	if err != nil {
		log.Println(err)
	}
	if backend == "" {
		conn.Write([]byte("backend.list\n"))
	} else {
		conn.Write([]byte("backend.list " + backend + "\n"))
	}
	conn.Write([]byte("backend.list\n"))
	byte_health := make([]byte, 512)
	n, err := conn.Read(byte_health)
	if err != nil {
		log.Printf("Could not read packet : %s", err.Error())
		return backends
	}
	status := string(byte_health[:n])
	for _, line := range strings.Split(status, "\n") {
		rp := regexp.MustCompile("^(\\S+\\))[\\s]+(\\S+)[\\s]+(\\S+)[\\s]+(.+)")
		list := rp.FindStringSubmatch(line)
		if len(list) > 0 {
			hs := HealthStatus{}
			hs.Refs = list[2]
			hs.Admin = list[3]
			hs.Probe = list[4]
			backends[list[1]] = hs
		}
	}
	return backends
}
