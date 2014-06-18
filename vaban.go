/* Vaban - The Simple Varnish Ban REST Api. <benjamin@martensson.io> */
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/ant0ine/go-json-rest/rest"
)

type Message struct {
	Msg string
}
type Messages map[string]Message

type BanPost struct {
	Pattern string
	Vcl     string
}

type Service struct {
	Hosts  []string `json:"Hosts"`
	Secret string   `json:"Secret"`
}
type Services map[string]Service

var services Services

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

func Pinger(server string) string {
	_, err := net.Dial("tcp", server)
	if err != nil {
		return err.Error()
	}
	return "tcp port open"
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
	log.Println(server, "ban status", strings.Trim(status, " "))
	return "ban status " + strings.Trim(status, " ")
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
				log.Println(server)
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

func main() {
	port := flag.String("p", "4000", "Listen on this port. (default 4000)")
	config := flag.String("f", "config.json", "Path to config. (default config.json)")
	flag.Parse()
	file, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(file, &services)
	if err != nil {
		log.Fatal("Problem parsing config: ", err)
	}
	handler := rest.ResourceHandler{
		EnableRelaxedContentType: true,
		EnableStatusService:      true,
		XPoweredBy:               "Vaban",
	}
	handler.SetRoutes(
		&rest.Route{"GET", "/",
			func(w rest.ResponseWriter, r *rest.Request) {
				w.WriteJson(handler.GetStatus())
			},
		},
		&rest.Route{"GET", "/v1/service/:service", GetService},
		&rest.Route{"GET", "/v1/services", GetServices},
		&rest.Route{"GET", "/v1/service/:service/ping", GetPing},
		&rest.Route{"POST", "/v1/service/:service/ban", PostBan},
	)
	log.Println("Starting Vaban on :" + *port)
	http.ListenAndServe(":"+*port, &handler)
}
