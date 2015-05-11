package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
)

type BanPost struct {
	Pattern string
	Vcl     string
}

func PostBan(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	service := ps.ByName("service")
	banpost := BanPost{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&banpost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if banpost.Pattern == "" && banpost.Vcl == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Pattern or VCL is required"))
		return
	} else if banpost.Pattern != "" && banpost.Vcl != "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Pattern or VCL is required, not both"))
		return
	}
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
				message.Msg = Banner(server, banpost, s.Secret, req)
				messages[server] = message
			}(server)
		}
		// Wait for all BANs to complete.
		wg.Wait()
		r.JSON(w, http.StatusOK, messages)
		return
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Service could not be found."))
		return
	}
}

func Banner(server string, banpost BanPost, secret string, req *http.Request) string {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Println(err)
		return err.Error()
	}
	defer conn.Close()
	varnishAuth(server, secret, conn)
	// sending the magic ban commmand to varnish.
	if banpost.Pattern != "" {
		conn.Write([]byte("ban req.url ~ " + banpost.Pattern + "$\n"))
	} else {
		conn.Write([]byte("ban " + banpost.Vcl + "\n"))
	}
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
		"vcl":     banpost.Vcl,
		"pattern": banpost.Pattern,
		"server":  server,
		"status":  status,
	})
	if reqID := req.Header.Get("X-Request-Id"); reqID != "" {
		entry = entry.WithField("request_id", reqID)
	}
	entry.Info("ban")
	return "ban status " + status
}
