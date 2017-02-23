package proxy

import (
	"fmt"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/mklimuk/goerr"
)

// additional error types
const (
	Conflict    goerr.ErrorType = 17
	InvalidType goerr.ErrorType = 18
)

//TargetsManager is responsible for registering proxy targets with the router
type TargetsManager interface {
	AddToPool(poolID, ID, targetURI string) error
	RemoveFromPool(poolID, ID string) error
	CreatePool(conf *TargetConfig) error
	Proxy(ctx *gin.Context)
}

//NewTargetsManager is the TargetsManager constructor
func NewTargetsManager(targets []*TargetConfig, keeper Gatekeeper) TargetsManager {
	t := &targetsManager{keeper: keeper}
	t.targets = make(map[string]Target)
	var tg Target
	var err error
	for _, conf := range targets {
		conf.keeper = t.keeper
		if tg, err = targetFromConfig(conf); err != nil {
			panic(err)
		}
		t.targets[conf.TID] = tg
	}
	return TargetsManager(t)
}

type targetsManager struct {
	targets map[string]Target
	keeper  Gatekeeper
}

func (t *targetsManager) AddToPool(poolID, ID, targetURI string) error {
	var p Pool
	var err error
	if p, err = t.getPool(poolID); err != nil {
		return err
	}
	var u *url.URL
	if u, err = url.Parse(targetURI); err != nil {
		return goerr.NewError("Invalid URL", goerr.BadRequest)
	}
	p.Add(ID, u)
	return nil
}

func (t *targetsManager) RemoveFromPool(poolID, ID string) error {
	var p Pool
	var err error
	if p, err = t.getPool(poolID); err != nil {
		return err
	}
	p.Remove(ID)
	return nil
}

func (t *targetsManager) getPool(poolID string) (Pool, error) {
	var p Target
	var ok bool
	if p, ok = t.targets[poolID]; !ok {
		return nil, goerr.NewError("Pool not found", goerr.NotFound)
	}
	if p.Type() != TypePool {
		return nil, goerr.NewError("Invalid target type", InvalidType)
	}
	return p.(Pool), nil
}

func (t *targetsManager) CreatePool(conf *TargetConfig) error {
	if _, ok := t.targets[conf.TID]; ok {
		return goerr.NewError("Pool already exists", Conflict)
	}
	conf.keeper = t.keeper
	p := NewPool(conf)
	t.targets[conf.TID] = p
	return nil
}

func (t *targetsManager) Proxy(ctx *gin.Context) {
	targetID := ctx.Param("id")
	var target Target
	var exists bool
	if target, exists = t.targets[targetID]; !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Target not found for id='%s'", targetID)})
		return
	}
	if log.GetLevel() >= log.DebugLevel {
		log.WithFields(log.Fields{"logger": "api-proxy.proxy", "target": target.ID(), "path": ctx.Param("path")}).
			Debug("Calling proxy target")
	}
	target.Handler()(ctx)
}

func targetFromConfig(conf *TargetConfig) (Target, error) {
	switch conf.TargetType {
	case TypeSingle:
		return NewSingle(conf)
	case TypePool:
		return NewPool(conf), nil
	}
	return nil, nil
}
