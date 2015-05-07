/* This test is hardcoded to use my own varnish test server. Adapt for your own use. */
package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"gopkg.in/yaml.v1"
)

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
func config() {
	file, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(file, &services)
	if err != nil {
		log.Fatal("Problem parsing config: ", err)
	}
}

func TestServices(t *testing.T) {
	res := httptest.NewRecorder()
	config()
	handler := initialize()
	req, err := http.NewRequest("GET", "http://localhost:4000/v1/services", nil)
	if err != nil {
		t.Error(err)
	}
	handler.ServeHTTP(res, req)
	expect(t, res.Code, http.StatusOK)
}

func TestService(t *testing.T) {
	res := httptest.NewRecorder()
	config()
	handler := initialize()
	req, err := http.NewRequest("GET", "http://localhost:4000/v1/service/yr-stage", nil)
	if err != nil {
		t.Error(err)
	}
	handler.ServeHTTP(res, req)
	expect(t, res.Code, http.StatusOK)
}

func TestPing(t *testing.T) {
	res := httptest.NewRecorder()
	config()
	handler := initialize()
	req, err := http.NewRequest("GET", "http://localhost:4000/v1/service/yr-stage/ping", nil)
	if err != nil {
		t.Error(err)
	}
	handler.ServeHTTP(res, req)
	expect(t, res.Code, http.StatusOK)
}

func TestServiceHealth(t *testing.T) {
	res := httptest.NewRecorder()
	config()
	handler := initialize()
	req, err := http.NewRequest("GET", "http://localhost:4000/v1/service/yr-stage/health", nil)
	if err != nil {
		t.Error(err)
	}
	handler.ServeHTTP(res, req)
	expect(t, res.Code, http.StatusOK)
}

func TestBackendHealth(t *testing.T) {
	res := httptest.NewRecorder()
	config()
	handler := initialize()
	req, err := http.NewRequest("GET", "http://localhost:4000/v1/service/yr-stage/health/yr", nil)
	if err != nil {
		t.Error(err)
	}
	handler.ServeHTTP(res, req)
	expect(t, res.Code, http.StatusOK)
}

func TestBanVCL(t *testing.T) {
	res := httptest.NewRecorder()
	config()
	handler := initialize()
	var content = []byte(`{"Vcl": "req.http.Host == 'example.com'"}`)
	req, err := http.NewRequest("POST", "http://localhost:4000/v1/service/yr-stage/ban", bytes.NewBuffer(content))
	if err != nil {
		t.Error(err)
	}
	handler.ServeHTTP(res, req)
	expect(t, res.Code, http.StatusOK)
}

func TestBanPattern(t *testing.T) {
	res := httptest.NewRecorder()
	config()
	handler := initialize()
	var content = []byte(`{"Pattern": "/example"}`)
	req, err := http.NewRequest("POST", "http://localhost:4000/v1/service/yr-stage/ban", bytes.NewBuffer(content))
	if err != nil {
		t.Error(err)
	}
	handler.ServeHTTP(res, req)
	expect(t, res.Code, http.StatusOK)
}
