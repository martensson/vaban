/* Vaban - The Simple Varnish Ban REST Api. <benjamin@martensson.io> */
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"

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
	Secret  string   `json:"Secret"`
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

func Banner(host string, pattern string, version int, secret string) string {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		log.Println(err)
		return "tcp port closed"
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
		log.Println(host, "auth status", strings.Trim(string(auth_reply)[0:12], " "))
	}
	// sending the magic ban commmand to varnish.
	if version >= 4 {
		conn.Write([]byte("ban req.url ~ " + pattern + "$\n"))
	} else {
		conn.Write([]byte("ban.url " + pattern + "$\n"))
	}
	// again, 64 bytes is enough for this.
	byte_status := make([]byte, 64)
	conn.Read(byte_status)
	conn.Close()
	// cast byte to string and only keep the status code (always max 13 char), the rest we dont care.
	status := string(byte_status)[0:12]
	log.Println(host, "ban status", strings.Trim(status, " "))
	return "ban status " + strings.Trim(status, " ")
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
			message.Msg = Banner(server, patternpost.Pattern, s.Version, s.Secret)
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
	log.Println("Starting Vaban on :4000")
	http.ListenAndServe(":4000", &handler)
}
