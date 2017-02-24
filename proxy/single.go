package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/koding/websocketproxy"
)

//NewSingle is a single HTTP proxy target constructor
func NewSingle(t *TargetConfig) (Target, error) {
	s := &single{
		TargetConfig: *t,
	}
	var err error
	if s.uri, err = url.Parse(t.URL); err != nil || s.uri == nil {
		return nil, err
	}
	if t.TargetProtocol == ProtocolHTTP {
		s.rp = httputil.NewSingleHostReverseProxy(s.URI())
	} else {
		proxy := websocketproxy.NewProxy(s.URI())
		proxy.Upgrader = upgrader
		s.rp = proxy
	}
	return Target(s), nil
}

type single struct {
	TargetConfig
	rp http.Handler
}

func (t *single) Handler() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		path := ctx.Param("path")
		checkAuthAndServe(t, path, t.rp, ctx)
	}
}
