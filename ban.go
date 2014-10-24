package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/emicklei/go-restful"
)

type BanPost struct {
	Pattern string
	Vcl     string
}

func PostBan(req *restful.Request, resp *restful.Response) {
	service := req.PathParameter("service")
	banpost := BanPost{}

	err := req.ReadEntity(&banpost)
	if err != nil {
		resp.AddHeader("Content-Type", "text/plain")
		resp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	if banpost.Pattern == "" && banpost.Vcl == "" {
		resp.WriteErrorString(http.StatusBadRequest, "Pattern or VCL is required")
		return
	} else if banpost.Pattern != "" && banpost.Vcl != "" {
		resp.WriteErrorString(http.StatusBadRequest, "Pattern or VCL is required, not both.")
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
				message.Msg = Banner(server, banpost, s.Secret)
				messages[server] = message
			}(server)
		}
		// Wait for all BANs to complete.
		wg.Wait()
		resp.WriteEntity(messages)
	} else {
		resp.WriteErrorString(http.StatusNotFound, "Service could not be found.")
		return
	}
}

func Banner(server string, banpost BanPost, secret string) string {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		log.Println(err)
		return err.Error()
	}
	// I want to allocate 512 bytes, enough to read the varnish help output.
	reply := make([]byte, 512)
	conn.Read(reply)
	rp := regexp.MustCompile("[a-z]{32}") //find challenge string
	challenge := rp.FindString(string(reply))
	if challenge != "" {
		// time to authenticate
		hash := sha256.New()
		hash.Write([]byte(challenge + "\n" + secret + "\n" + challenge + "\n"))
		md := hash.Sum(nil)
		mdStr := hex.EncodeToString(md)
		conn.Write([]byte("auth " + mdStr + "\n"))
		auth_reply := make([]byte, 512)
		conn.Read(auth_reply)
		log.Println(server, "auth status", strings.Trim(string(auth_reply)[0:12], " "))
	}
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
	conn.Close()
	// cast byte to string and only keep the status code (always max 13 char), the rest we dont care.
	status := string(byte_status)[0:12]
	status = strings.Trim(status, " ")
	log.Println(server, "banned with status", status)
	return "ban status " + status
}
