/* Vaban - The Simple Varnish Ban REST Api. <benjamin@martensson.io> */
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
	"github.com/pilu/xrequestid"
	"github.com/thoas/stats"
	"github.com/unrolled/render"
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

var r = render.New(render.Options{
	IndentJSON: true,
})

func initialize() *negroni.Negroni {
	vabanstats := stats.New()
	n := negroni.New(
		negroni.NewRecovery(),
		NewLogger(),
		xrequestid.New(8),
	)
	router := httprouter.New()
	router.GET("/", func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		stats := vabanstats.Data()
		r.JSON(w, http.StatusOK, stats)
	})
	router.GET("/v1/services", GetServices)
	router.GET("/v1/service/:service", GetService)
	router.GET("/v1/service/:service/ping", GetPing)
	router.GET("/v1/service/:service/health", GetHealth)
	router.GET("/v1/service/:service/health/:backend", GetHealth)
	router.POST("/v1/service/:service/health/:backend", PostHealth)
	router.POST("/v1/service/:service/ban", PostBan)
	// add router and clear mux.context values at the end of request life-times
	n.UseHandler(router)
	return n
}

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
	n := initialize()
	log.Println("Starting vaban on :" + *port)
	n.Run(":" + *port)
}
