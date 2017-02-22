package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/koding/websocketproxy"
)

//NewSingle is a single HTTP proxy target constructor
func NewSingle(t *TargetConfig) Target {
	s := &single{
		TargetConfig: *t,
	}
	if t.TargetProtocol == ProtocolHTTP {
		s.rp = httputil.NewSingleHostReverseProxy(s.URI())
	} else {
		s.rp = websocketproxy.NewProxy(s.URI())
	}
	return Target(s)
}

type single struct {
	TargetConfig
	rp http.Handler
}

func (t *single) Handler() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		path := ctx.Param("path")
		condition := t.PrivilegesForPath(path, ctx.Request.Method)
		// if the API is protected we should perform necessary checks
		if condition > 0 {
			h := ctx.Request.Header.Get("Authorization")
			if h == "" {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Requested API is not public but no Authorization header is present"})
				return
			}

			var token string
			var err error
			if token, err = t.keeper.CheckAccess(extractToken(h), condition, t.UpdateToken()); err != nil {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token", "details": err.Error()})
				return
			}
			ctx.Header("token", token)
		}

		// rewrite request URL/URL
		ctx.Request.RequestURI = path
		var err error
		if ctx.Request.URL, err = url.Parse(ctx.Request.RequestURI); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse target path", "description": err.Error()})
			return
		}
		t.rp.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func (t *single) handleWebsocket(ctx *gin.Context) {

}
