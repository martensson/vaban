/* Vaban - The Simple Varnish Ban REST Api. <benjamin@martensson.io> */
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/ant0ine/go-json-rest/rest"
	"gopkg.in/yaml.v1"
)

type Message struct {
	Msg string
}
type Messages map[string]Message

type Service struct {
	Hosts  []string
	Secret string
}
type Services map[string]Service

var services Services

func main() {
	accessfile, err := os.OpenFile("/tmp/vaban_access_log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	accesslogger := log.New(accessfile, "", 0)
	port := flag.String("p", "4000", "Listen on this port. (default 4000)")
	config := flag.String("f", "config.yml", "Path to config. (default config.yml)")
	flag.Parse()
	file, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(file, &services)
	if err != nil {
		log.Fatal("Problem parsing config: ", err)
	}
	handler := rest.ResourceHandler{
		EnableRelaxedContentType: true,
		EnableStatusService:      true,
		XPoweredBy:               "Vaban",
		EnableLogAsJson:          true,
		Logger:                   accesslogger,
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
