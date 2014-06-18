# Vaban

*A quick and easy way to control groups of Varnish Cache hosts using a RESTful JSON API.*

[![Build Status](https://travis-ci.org/martensson/vaban.svg?branch=master)](https://travis-ci.org/martensson/vaban)

Vaban is built in Go for super high performance and concurrency. Every request and every ban spawns its own lightweight thread.
It supports Varnish 3 + 4, Authentication, Pattern-based and VCL-based banning.

## Getting Started

### Installing from source

#### Dependencies

* Git
* Go 1.1+

#### Clone and Build locally:

``` sh
git clone https://github.com/martensson/vaban.git
cd vaban
go build vaban.go
```

#### Create a config.yml file and add all your services:

``` yaml
---
service1:
  hosts:
    - "a.example.com:6082"
    - "b.example.com:6082"
    - "c.example.com:6082"
    - "d.example.com:6082"
    - "e.example.com:6082"
  secret: "1111-2222-3333-aaaa-bbbb-cccc"

service2:
  hosts:
    - "x.example.com:6082"
    - "y.example.com:6082"
    - "z.example.com:6082"
  secret: "1111-2222-3333-aaaa-bbbb-cccc"
```

#### Running Vaban
``` sh
./vaban -p 4000 -f /path/to/config.yml
```


**Make sure that the varnish admin interface is available, listening on 0.0.0.0:6082**


### REST API Reference

#### get status

    GET /
    Expected HTTP status code: 200

#### get all services
    
    GET /v1/services
    Expected HTTP status code: 200

#### get all hosts in service

    GET /v1/service/:service
    Expected HTTP status code: 200

#### tcp scan all hosts

    GET /v1/service/:service/ping
    Expected HTTP status code: 200

#### ban based on pattern

    POST /v1/service/:service/ban
    JSON Body: {"Pattern":"..."}
    Expected HTTP status code: 200

#### ban based on vcl

    POST /v1/service/:service/ban
    JSON Body: {"Vcl":"..."}
    Expected HTTP status code: 200

### CURL Examples

#### Get status of Vaban

``` sh
curl -i http://127.0.0.1:4000/
```

#### Get all groups

``` sh
curl -i http://127.0.0.1:4000/v1/services
```

#### Get all hosts in group

``` sh
curl -i http://127.0.0.1:4000/v1/service/group1
```

#### Scan hosts to see if tcp port is open

``` sh
curl -i http://127.0.0.1:4000/v1/service/group1/ping
```

#### Ban the root of your website.

``` sh
curl -i http://127.0.0.1:4000/v1/service/group1/ban -d '{"Pattern":"/"}'
```

#### Ban all css files

``` sh
curl -i http://127.0.0.1:4000/v1/service/group1/ban -d '{"Pattern":".*css"}'
```

#### Ban based on VCL, in this case all objects matching a host-header.

``` sh
curl -i http://127.0.0.1:4000/v1/service/group1/ban -d '{"Vcl":"req.http.Host == 'example.com'"}'
```
