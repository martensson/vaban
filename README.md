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
* Go 1.4+

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

#### from source:
``` sh
./vaban -p 4000 -f /path/to/config.yml
```
#### from packages:
``` sh
vaban run web
```
or
``` sh
vaban scale web=1
service vaban start
vaban logs
```

#### from Docker

If you already have Docker install its really use to compile and run with two commands:

    docker build -t vaban .
    docker run -p 4000:4000 vaban


**Make sure that the varnish admin interface is available on your hosts, listening on 0.0.0.0:6082**

### API

A quick and easy way to control clusters of Varnish Cache hosts using a RESTful JSON API.

#### GET /v1/services

Get all groups:

+ Response 200 (application/json)

        [
            "group1",
            "group2",
            "group3",
        ]
        

#### GET /v1/service/group1

Get all hosts in group:

+ Response 200 (application/json)

        [
            "test01:6082"
        ]

#### GET /v1/service/group1/ping

Scan hosts to see if tcp port is open:

+ Response 200 (application/json)

        {
            "test01:6082": {
                "Msg": "PONG 1431078011 1.0"
            }
        }

#### GET /v1/service/group1/health

Check health status of all backends:

+ Response 200 (application/json)

        {
            "test01:6082": {
                "backend01(10.160.101.100,,80)": {
                    "Refs": "1",
                    "Admin": "probe",
                    "Probe": "Healthy 4/4"
                }
            }
        }



#### GET /v1/service/group1/health/www01

Check health status of one backend:

+ Response 200 (application/json)

        {
            "test01:6082": {
                "backend01(10.160.101.100,,80)": {
                    "Refs": "1",
                    "Admin": "probe",
                    "Probe": "Healthy 3/4"
                }
            }
        }

#### POST /v1/service/group1/health/www01

force health status of one backend (can be healthy, sick or auto):

+ Request (application/json)

        {"Set_health":"sick"}

+ Response 200 (application/json)

        {
            "test01:6082": {
                "Msg": "updated with status 200 0"
            }
        }

#### POST /v1/service/group1/ban

To ban elements in your cache.

+ Request Ban the root of your website (application/json)

        {"Pattern":"/"}
        
+ Request Ban all css files (application/json)

        {"Pattern":".*css"}
        
+ Request Ban based on VCL, in this case all objects matching a host-header. (application/json)

        {"Vcl":"req.http.Host == 'example.com'"}"}

+ Response 200 (application/json)

        {
            "test01:6082": {
                "Msg": "ban status 200 0"
            }
        }

