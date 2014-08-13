package main

import (
	"net"
	"sync"

	"github.com/ant0ine/go-json-rest/rest"
)

func Pinger(server string) string {
	_, err := net.Dial("tcp", server)
	if err != nil {
		return err.Error()
	}
	return "tcp port open"
}

func GetPing(w rest.ResponseWriter, r *rest.Request) {
	service := r.PathParam("service")

	if s, ok := services[service]; ok {
		// We need the WaitGroup for some awesome Go concurrency of our BANs
		var wg sync.WaitGroup
		messages := Messages{}
		for _, server := range s.Hosts {
			// Increment the WaitGroup counter.
			wg.Add(1)
			go func(server string) {
				// Decrement the counter when the goroutine completes.
				defer wg.Done()
				message := Message{}
				message.Msg = Pinger(server)
				messages[server] = message
			}(server)
		}
		// Wait for all PINGs to complete.
		wg.Wait()
		w.WriteJson(messages)
	} else {
		rest.NotFound(w, r)
		return
	}
}
