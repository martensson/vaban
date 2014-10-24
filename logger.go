package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
)

var accessfile, err = os.OpenFile("/var/log/vaban_access_log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
var logger *log.Logger = log.New(accessfile, "", 0)

func NCSACommonLogFormatLogger() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		var username = "-"
		if req.Request.URL.User != nil {
			if name := req.Request.URL.User.Username(); name != "" {
				username = name
			}
		}
		chain.ProcessFilter(req, resp)
		logger.Printf("%s - %s [%s] \"%s %s %s\" %d %d",
			strings.Split(req.Request.RemoteAddr, ":")[0],
			username,
			time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			req.Request.Method,
			req.Request.URL.RequestURI(),
			req.Request.Proto,
			resp.StatusCode(),
			resp.ContentLength(),
		)
	}
}
