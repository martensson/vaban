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

	"github.com/ant0ine/go-json-rest/rest"
)

type BanPost struct {
	Pattern string
	Vcl     string
}

func PostBan(w rest.ResponseWriter, r *rest.Request) {
	service := r.PathParam("service")
	banpost := BanPost{}
	err := r.DecodeJsonPayload(&banpost)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if banpost.Pattern == "" && banpost.Vcl == "" {
		rest.Error(w, "Pattern or VCL is required", 400)
		return
	} else if banpost.Pattern != "" && banpost.Vcl != "" {
		rest.Error(w, "Pattern or VCL is required, not both.", 400)
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
		w.WriteJson(messages)
	} else {
		rest.NotFound(w, r)
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
	conn.Read(byte_status)
	conn.Close()
	// cast byte to string and only keep the status code (always max 13 char), the rest we dont care.
	status := string(byte_status)[0:12]
	log.Println(server, "banned with status", strings.Trim(status, " "))
	return "ban status " + strings.Trim(status, " ")
}
