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

func GetHealth(req *restful.Request, resp *restful.Response) {
	service := req.PathParameter("service")
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
				backends = Health(server, s.Secret)
				servers[server] = backends
			}(server)
		}
		// Wait for all BANs to complete.
		wg.Wait()
		resp.WriteEntity(servers)
	} else {
		resp.WriteErrorString(http.StatusNotFound, "Service could not be found.")
		return
	}
}

func Health(server string, secret string) Backends {
	backends := Backends{}
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Println(err)
		return backends
	}
	varnishAuth(server, secret, conn)
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
