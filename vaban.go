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
	restful.DefaultResponseContentType(restful.MIME_JSON)
	restful.DefaultRequestContentType(restful.MIME_JSON)
	restful.PrettyPrintResponses = true
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/services").To(GetServices).
		// docs
		Doc("get all services").
		Returns(200, "msg received.", nil).
		Operation("GetServices"))
	ws.Route(ws.GET("/service/{service}").To(GetService).
		// docs
		Doc("get all hosts in service").
		Operation("GetService").
		Returns(200, "msg received.", nil).
		Param(ws.PathParameter("service", "identifier of the service").DataType("string")))
	ws.Route(ws.GET("/service/{service}/ping").To(GetPing).
		// docs
		Doc("ping all hosts in service").
		Operation("GetPing").
		Returns(200, "msg received.", Messages{}).
		Param(ws.PathParameter("service", "identifier of the service").DataType("string")))
	ws.Route(ws.GET("/service/{service}/health").To(GetHealth).
		// docs
		Doc("get health status of all backends").
		Operation("GetHealth").
		Returns(200, "msg received.", Servers{}).
		Param(ws.PathParameter("service", "identifier of the service").DataType("string")))
	ws.Route(ws.GET("/service/{service}/health/{backend}").To(GetHealth).
		// docs
		Doc("get health status of specific backend").
		Operation("GetHealth").
		Returns(200, "msg received.", Servers{}).
		Param(ws.PathParameter("service", "identifier of the service").DataType("string")).
		Param(ws.PathParameter("backend", "identifier of the backend").DataType("string")))
	ws.Route(ws.POST("/service/{service}/health/{backend}").To(PostHealth).
		// docs
		Doc("set health status on specific backend (auto/sick/healthy)").
		Operation("PostHealth").
		Param(ws.PathParameter("service", "identifier of the service").DataType("string")).
		Param(ws.PathParameter("backend", "identifier of the backend").DataType("string")).
		Notes("Valid health status is 'auto', 'sick' or 'healthy'. Default is 'auto'.").
		Returns(200, "msg received.", Messages{}).
		Reads(HealthPost{}))
	ws.Route(ws.POST("/service/{service}/ban").To(PostBan).
		// docs
		Doc("ban elements from all hosts in service").
		Operation("PostBan").
		Param(ws.PathParameter("service", "identifier of the service").DataType("string")).
		Returns(200, "msg received.", Messages{}).
		Reads(BanPost{}))
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
