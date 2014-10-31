package main

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/emicklei/go-restful"
)

func Pinger(server string, secret string) string {
	conn, err := net.Dial("tcp", server)
	defer conn.Close()
	if err != nil {
		return err.Error()
	}
	err = varnishAuth(server, secret, conn)
	if err != nil {
		log.Println(err)
	}
	conn.Write([]byte("ping\n"))
	pong := make([]byte, 32)
	_, err = conn.Read(pong)
	if err != nil {
		return err.Error()
	}
	status := string(pong)[13:32]
	status = strings.Trim(status, " ")
	return status
}

func GetPing(req *restful.Request, resp *restful.Response) {
	service := req.PathParameter("service")

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
				message.Msg = Pinger(server, s.Secret)
				messages[server] = message
			}(server)
		}
		// Wait for all PINGs to complete.
		wg.Wait()
		resp.WriteEntity(messages)
	} else {
		resp.WriteErrorString(http.StatusNotFound, "Service could not be found.")
		return
	}
}
