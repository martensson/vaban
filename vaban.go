/* Vaban - The Simple Varnish Ban REST Api. <benjamin@martensson.io> */
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
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
	ws := new(restful.WebService)
	ws.Path("/v1")
	ws.Filter(NCSACommonLogFormatLogger())
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/services").To(GetServices).
		// docs
		Doc("get all services").
		Operation("GetServices").
		Reads(Services{}))
	ws.Route(ws.GET("/service/{service}").To(GetService).
		// docs
		Doc("get a service").
		Operation("GetService").
		Param(ws.PathParameter("service", "identifier of the service").DataType("string")).
		Reads(Service{}))
	ws.Route(ws.GET("/service/{service}/ping").To(GetPing).
		// docs
		Doc("ping all hosts in service").
		Operation("GetPing").
		Param(ws.PathParameter("service", "identifier of the service").DataType("string")).
		Reads(Messages{}))
	ws.Route(ws.GET("/service/{service}/health").To(GetHealth).
		// docs
		Doc("get health status of all backends per host in service").
		Operation("GetHealth").
		Param(ws.PathParameter("service", "identifier of the service").DataType("string")).
		Reads(Messages{}))
	ws.Route(ws.POST("/service/{service}/ban").To(PostBan).
		// docs
		Doc("ban elements from all hosts in service").
		Operation("PostBan").
		Param(ws.PathParameter("service", "identifier of the service").DataType("string")).
		Reads(Messages{}))
	restful.DefaultResponseContentType(restful.MIME_JSON)
	restful.DefaultRequestContentType(restful.MIME_JSON)
	restful.Add(ws)

	sc := swagger.Config{
		WebServices: restful.RegisteredWebServices(), // you control what services are visible
		ApiPath:     "/api.json",
		// Optionally, specifiy where the UI is located
		SwaggerPath:     "/",
		SwaggerFilePath: "swagger"}
	swagger.InstallSwaggerService(sc)

	log.Println("Starting Vaban on :" + *port)
	http.ListenAndServe(":4000", nil)
}
