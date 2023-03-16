# goconfig

[![Build Status](https://github.com/demdxx/goconfig/workflows/run%20tests/badge.svg)](https://github.com/demdxx/goconfig/actions?workflow=run%20tests)
[![Go Report Card](https://goreportcard.com/badge/github.com/demdxx/goconfig)](https://goreportcard.com/report/github.com/demdxx/goconfig)
[![GoDoc](https://godoc.org/github.com/demdxx/goconfig?status.svg)](https://godoc.org/github.com/demdxx/goconfig)
[![Coverage Status](https://coveralls.io/repos/github/demdxx/goconfig/badge.svg)](https://coveralls.io/github/demdxx/goconfig)

Golang config initialization module which provides simple functionality for the loading of configs based on **struct**ures description.

## Example

### config.go
```go
package config

type serverConfig struct {
	HTTP struct {
		Listen       string        `default:":8080"  json:"listen" yaml:"listen" cli:"http-listen" env:"SERVER_HTTP_LISTEN"`
		ReadTimeout  time.Duration `default:"120s"   json:"read_timeout" yaml:"read_timeout" env:"SERVER_HTTP_READ_TIMEOUT"`
		WriteTimeout time.Duration `default:"120s"   json:"write_timeout" yaml:"write_timeout" env:"SERVER_HTTP_WRITE_TIMEOUT"`
	}
	GRPC struct {
		Listen  string        `default:"tcp://:8081" json:"listen" yaml:"listen" cli:"grpc-listen" env:"SERVER_GRPC_LISTEN"`
		Timeout time.Duration `default:"120s"        json:"timeout" yaml:"timeout" env:"SERVER_GRPC_TIMEOUT"`
	}
	Profile struct {
		Mode   string `json:"mode" yaml:"mode" default:"" env:"SERVER_PROFILE_MODE"`
		Listen string `json:"listen" yaml:"listen" default:"" env:"SERVER_PROFILE_LISTEN"`
	}
}

type ConfigType struct {
	ServiceName    string `json:"service_name" yaml:"service_name" env:"SERVICE_NAME" default:"disk"`
	DatacenterName string `json:"datacenter_name" yaml:"datacenter_name" env:"DC_NAME" default:"??"`
	Hostname       string `json:"hostname" yaml:"hostname" env:"HOSTNAME" default:""`
	Hostcode       string `json:"hostcode" yaml:"hostcode" env:"HOSTCODE" default:""`

	LogAddr  string `default:"" env:"LOG_ADDR"`
	LogLevel string `default:"debug" env:"LOG_LEVEL"`

	Server serverConfig `json:"server" yaml:"server"`
}

var Config ConfigType
```

### main.go

```go
package main

import (
  configLoader "github/demdxx/goconfig"
  "config"
)

func init() {
  if err := configLoader.Load(&config.Config); err != nil {
    panic(err)
  }
}

func main() {
  // ...
}
```

## Dependencies

* [github.com/caarlos0/env](github.com/caarlos0/env)
* [github.com/gravitational/configure](github.com/gravitational/configure)
* [github.com/hashicorp/hcl](github.com/hashicorp/hcl)
* [github.com/mcuadros/go-defaults](github.com/mcuadros/go-defaults)

## TODO

* [ ] Add support of environment prefixes
