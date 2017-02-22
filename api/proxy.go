package api

import (
	"net/http"

	"github.com/mklimuk/api-proxy/proxy"
	"github.com/mklimuk/goerr"
	"github.com/mklimuk/husar/rest"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

type proxyAPI struct {
	manager proxy.TargetsManager
}

type poolInfo struct {
	PoolID          string             `json:"id"`
	PrivilegesLevel int                `json:"privilegesLevel,opitempty"`
	Protocol        proxy.ProtocolType `json:"protocol"`
	TargetID        string             `json:"targetId,opitempty"`
	TargetURI       string             `json:"targetUri,omitempty"`
}

//NewProxyAPI is the proxy proxy API constructor
func NewProxyAPI(manager proxy.TargetsManager) rest.API {
	p := proxyAPI{manager}
	return rest.API(&p)
}

//AddRoutes initializes proxy API routes
func (p *proxyAPI) AddRoutes(router *gin.Engine) {
	router.POST("/pool", p.createPool)
	router.DELETE("/pool/:poolId", p.deletePool)
	router.POST("/pool/:poolId", p.addToPool)
	router.DELETE("/pool/:poolId/:endpointId", p.deleteFromPool)
	router.Any("/api/:id/*path", p.proxy)
	router.GET("/ws/:id/*path", p.proxy)
}

func (p *proxyAPI) createPool(ctx *gin.Context) {
	defer rest.ErrorHandler(ctx)
	var err error
	l := new(poolInfo)
	if err = ctx.BindJSON(l); err != nil {
		log.WithFields(log.Fields{"logger": "proxy.api", "method": "createPool", "error": err}).
			Warn("Could not parse request")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse input", "details": err.Error()})
		return
	}

	pool := &proxy.TargetConfig{
		TID:               l.PoolID,
		TargetProtocol:    l.Protocol,
		DefaultPrivileges: l.PrivilegesLevel,
	}

	if err = p.manager.CreatePool(pool); err != nil {
		switch goerr.GetType(err) {
		case proxy.Conflict:
			ctx.JSON(http.StatusConflict, gin.H{"error": "Pool with this id already exists"})
			return
		default:
			log.WithFields(log.Fields{"logger": "proxy.api", "method": "createPool", "error": err}).
				WithError(err).Error("Error processing request")
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error occured", "details": err.Error()})
			return
		}
	}
	ctx.JSON(http.StatusOK, l)
}

func (p *proxyAPI) addToPool(ctx *gin.Context) {
	defer rest.ErrorHandler(ctx)
	var err error
	u := new(poolInfo)
	if err = ctx.BindJSON(u); err != nil {
		log.WithFields(log.Fields{"logger": "proxy.api", "method": "createUser", "error": err}).
			Warn("Could not parse request")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse input", "details": err.Error()})
		return
	}
	if err = p.manager.AddToPool(u.PoolID, u.TargetID, u.TargetURI); err != nil {
		switch goerr.GetType(err) {
		case goerr.NotFound:
			ctx.JSON(http.StatusNotFound, goerr.GetCtx(err))
			return
		case proxy.Conflict:
			ctx.JSON(http.StatusConflict, goerr.GetCtx(err))
			return
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error occured", "details": err.Error()})
			return
		}
	}
	ctx.AbortWithStatus(http.StatusOK)
}

func (p *proxyAPI) deletePool(ctx *gin.Context) {
	defer rest.ErrorHandler(ctx)
	//var err error

	ctx.JSON(http.StatusOK, gin.H{"token": "abc"})
}

func (p *proxyAPI) deleteFromPool(ctx *gin.Context) {
	defer rest.ErrorHandler(ctx)
	//var err error

	ctx.JSON(http.StatusOK, gin.H{"token": "abc"})
}

func (p *proxyAPI) proxy(ctx *gin.Context) {
	p.manager.Proxy(ctx)
}
