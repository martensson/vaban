# Vaban

*A quick and easy way to control clusters of Varnish Cache hosts using a RESTful JSON API.*

[![Build Status](https://travis-ci.org/martensson/vaban.svg?branch=master)](https://travis-ci.org/martensson/vaban)

Vaban is built in Go for high performance, concurrency and simplicity. Every request and every ban spawns its own lightweight thread.
It supports Varnish 3.0.3 + 4, Authentication, Pattern-based/VCL-based banning, health status, enable/disable backends, and more stuff to come. 

## Getting Started

### Installing from packages

The easiest way to install Vaban is from packages.

- Currently enabled for Ubuntu 14.04/12.04 and Debian 7.
- Current packages available from [packager.io](https://packager.io/gh/martensson/vaban/)

### Installing from source

#### Dependencies

* Git
* Go 1.3+

#### Clone and Build locally:

``` sh
git clone https://github.com/martensson/vaban.git
cd vaban
go build
```

### Create a config.yml file and add all your services:

Put the file inside your application root, if installing from package: /opt/vaban/config.yml

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

### Running Vaban

If compiling from source:
``` sh
./vaban -p 4000 -f /path/to/config.yml
```
If you installed from packages:
``` sh
vaban run web
```
or
``` sh
vaban scale web=1
service vaban start
vaban logs
```


**Make sure that the varnish admin interface is available on your hosts, listening on 0.0.0.0:6082**


### SWAGGER REST API Reference

Visit http://127.0.0.1:4000/

### CURL Examples

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

#### Check health status of all backends

``` sh
curl -i http://127.0.0.1:4000/v1/service/group1/health
```

#### Check health status of one backend

``` sh
curl -i http://127.0.0.1:4000/v1/service/group1/health/www01
```

#### force health status of one backend (can be healthy, sick or auto)

``` sh
curl -i http://127.0.0.1:4000/v1/service/group1/health/www01 -d '{"Set_health":"sick"}' -H 'Content-Type: application/json'
```

#### Ban the root of your website.

``` sh
curl -i http://127.0.0.1:4000/v1/service/group1/ban -d '{"Pattern":"/"}' -H 'Content-Type: application/json'
```

#### Ban all css files

``` sh
curl -i http://127.0.0.1:4000/v1/service/group1/ban -d '{"Pattern":".*css"}' -H 'Content-Type: application/json'
```

#### Ban based on VCL, in this case all objects matching a host-header.

``` sh
curl -i http://127.0.0.1:4000/v1/service/group1/ban -d '{"Vcl":"req.http.Host == 'example.com'"}' -H 'Content-Type: application/json'
```
