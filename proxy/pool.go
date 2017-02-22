package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
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
		id := ctx.Param("targetID")
		if id == "" {
			// broadcast mode
			// send request to all and aggregate responses somehow
		}
		var rp http.Handler
		var ok bool
		if rp, ok = t.rp[id]; !ok {
			ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("target %s not found", id)})
			return
		}
		rp.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func (t *pool) handleWebsocket(ctx *gin.Context) {

}

func (t *pool) Add(ID string, uri *url.URL) {
	t.rp[ID] = httputil.NewSingleHostReverseProxy(uri)
}

func (t *pool) Remove(ID string) {
	delete(t.rp, ID)
}
