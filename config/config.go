package config

import (
	"time"

	"github.com/mklimuk/api-proxy/proxy"
)

/*
Configuration is a struct containing different configuration options
*/
type Configuration struct {
	Targets []*proxy.TargetConfig `yaml:"targets"`
}

//Timezone is a reference timezone for the system
var Timezone, _ = time.LoadLocation("Europe/Warsaw")
