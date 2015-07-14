package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
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

func GetHealth(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	service := ps.ByName("service")
	backend := ps.ByName("backend")
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
		r.JSON(w, http.StatusOK, servers)
	} else {
		w.Write([]byte("Service could not be found."))
		return
	}
}

func PostHealth(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	service := ps.ByName("service")
	backend := ps.ByName("backend")
	healthpost := HealthPost{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&healthpost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if healthpost.Set_health == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Set_health is required"))
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
				message.Msg = UpdateHealth(server, s.Secret, backend, healthpost, req)
				messages[server] = message
			}(server)
		}
		wg.Wait()
		r.JSON(w, http.StatusOK, messages)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Service could not be found."))
		return
	}
}

func UpdateHealth(server string, secret string, backend string, healthpost HealthPost, req *http.Request) string {
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
	entry := logrus.WithFields(logrus.Fields{
		"set_health": healthpost.Set_health,
		"backend":    backend,
		"server":     server,
		"status":     status,
	})
	if reqID := req.Header.Get("X-Request-Id"); reqID != "" {
		entry = entry.WithField("request_id", reqID)
	}
	entry.Info("health")
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
	byte_health := make([]byte, 2048)
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
