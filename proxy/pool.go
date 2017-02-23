package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/koding/websocketproxy"
)

//NewPool is a pooled proxy target constructor
func NewPool(t *TargetConfig) Pool {
	p := &pool{
		TargetConfig: *t,
		rp:           make(map[string]http.Handler),
	}
	return Pool(p)
}

type pool struct {
	TargetConfig
	rp map[string]http.Handler
}

func (t *pool) Handler() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		path := ctx.Param("path")
		var id string
		id, path = extractID(path)

		var rp http.Handler
		var ok bool
		if rp, ok = t.rp[id]; !ok {
			ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("target %s not found", id)})
			return
		}
		checkAuthAndServe(t, path, rp, ctx)
	}
}

func (t *pool) Add(ID string, uri *url.URL) {
	if t.Protocol() == ProtocolHTTP {
		t.rp[ID] = httputil.NewSingleHostReverseProxy(uri)
	} else {
		t.rp[ID] = websocketproxy.NewProxy(uri)
	}
}

func (t *pool) Remove(ID string) {
	delete(t.rp, ID)
}

var extract = regexp.MustCompile(`\/`)

func extractID(path string) (string, string) {
	parts := extract.Split(path, 3)
	if len(parts) < 3 {
		return parts[1], ""
	}
	return parts[1], fmt.Sprintf("/%s", parts[2])
}
