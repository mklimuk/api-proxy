package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/mklimuk/api-proxy/api"
	"github.com/mklimuk/api-proxy/config"
	"github.com/mklimuk/api-proxy/proxy"
	"github.com/mklimuk/husar/util"

	"github.com/gin-gonic/gin"

	log "github.com/Sirupsen/logrus"
)

const (
	defaultLogLevel = "warn"
	defaultConfig   = "/etc/husar/config.yml"
)

func main() {

	clog := log.WithFields(log.Fields{"logger": "auth.main"})

	level := util.GetEnv("LOG", defaultLogLevel)
	conf := util.GetEnv("CONFIG", defaultConfig)

	var err error
	var l log.Level
	if l, err = log.ParseLevel(level); err != nil {
		clog.WithField("level", level).Panicln("Invalid log level")
	}
	log.SetLevel(l)

	rawConf := config.Parse(conf)
	fmt.Printf("Loaded configuration:\n %s\n", rawConf)

	clog.Info("Initializing services")
	router := gin.New()

	var authURL *url.URL

	if authURL, err = url.Parse("http://auth"); err != nil {
		panic(err)
	}
	keeper := proxy.NewGatekeeper(authURL)
	rp := proxy.NewTargetsManager(config.Config.Targets, keeper)

	clog.Info("Initializing REST router...")
	p := api.NewProxyAPI(rp)
	c := api.NewControlAPI()
	p.AddRoutes(router)
	c.AddRoutes(router)
	clog.Fatal(http.ListenAndServe(":8080", router))
}
