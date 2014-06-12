/* Vaban - The Simple Varnish Ban REST Api. <benjamin@martensson.io> */
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
)

type Message struct {
	Msg string
}
type Messages map[string]Message

type PatternPost struct {
	Pattern string
}

type Service struct {
	Hosts   []string `json:"Hosts"`
	Version int      `json:"Version"`
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

func Pinger(host string) string {
	_, err := net.Dial("tcp", host)
	if err != nil {
		return "tcp port closed"
	}
	return "tcp port open"
}

func Banner(host string, pattern string) string {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		log.Println(err)
		return "tcp port closed"
	}
	// I want to allocate 512 bytes, enough to read the varnish help output.
	reply := make([]byte, 512)
	conn.Read(reply)
	// sending the magic ban commmand to varnish.
	conn.Write([]byte("ban.url " + pattern + "$\n"))
	// again, 64 bytes is enough for this.
	byte_status := make([]byte, 64)
	conn.Read(byte_status)
	conn.Close()
	// cast byte to string and only keep the status code, the rest we dont care.
	status := string(byte_status)[0:5]
	log.Println(host, "ban status", status)
	return "ban status " + status
}

func GetPing(w rest.ResponseWriter, r *rest.Request) {
	service := r.PathParam("service")

	if s, ok := services[service]; ok {
		messages := Messages{}
		for _, server := range s.Hosts {
			message := Message{}
			message.Msg = Pinger(server)
			messages[server] = message
		}
		w.WriteJson(messages)
	} else {
		rest.NotFound(w, r)
		return
	}
}

func PostBan(w rest.ResponseWriter, r *rest.Request) {
	service := r.PathParam("service")
	patternpost := PatternPost{}
	err := r.DecodeJsonPayload(&patternpost)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if patternpost.Pattern == "" {
		rest.Error(w, "Pattern is required", 400)
		return
	}

	if s, ok := services[service]; ok {
		messages := Messages{}
		for _, server := range s.Hosts {
			message := Message{}
			message.Msg = Banner(server, patternpost.Pattern)
			messages[server] = message
		}
		w.WriteJson(messages)
	} else {
		rest.NotFound(w, r)
		return
	}
	var message Message
	message.Msg = "ban sent to service " + service
}

func main() {
	file, err := ioutil.ReadFile("config.json")
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
		&rest.Route{"GET", "/v1/service/:service/ping", GetPing},
		&rest.Route{"POST", "/v1/service/:service/ban", PostBan},
	)
	log.Println("Starting Vaban on :3000")
	http.ListenAndServe(":3000", &handler)
}
