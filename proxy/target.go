package proxy

import (
	"net/url"

	"github.com/gin-gonic/gin"
)

//TargetType enumerates supported target types
type TargetType string

//ProtocolType defines what kind of protocol is used by a given endpoint
type ProtocolType string

//Types of targets
const (
	TypeSingle TargetType = "single"
	TypePool   TargetType = "pool"
)

//Protocol types
const (
	ProtocolHTTP      ProtocolType = "HTTP"
	ProtocolWebsocket ProtocolType = "WS"
)

//Target defines an endpoint connection handler behavior
type Target interface {
	ID() string
	Type() TargetType
	Protocol() ProtocolType
	Handler() func(ctx *gin.Context)
	URI() *url.URL
	UpdateToken() bool
	PrivilegesForPath(path, method string) int
}

//Pool defines additional methods supported by a pool of endpoints
type Pool interface {
	Target
	Add(ID string, uri *url.URL)
	Remove(ID string)
}

// TargetConfig wraps proxy target configuration
type TargetConfig struct {
	TID            string       `yaml:"id"`
	TargetType     TargetType   `yaml:"type"`
	URL            string       `yaml:"url"`
	UpdatesToken   bool         `yaml:"updatesToken"`
	TargetProtocol ProtocolType `yaml:"protocol"`
	Privileges     *Privileges  `yaml:"privileges"`
	keeper         Gatekeeper
	uri            *url.URL
}

// Privileges regroups specific path privileges for a given endpoint
type Privileges struct {
	Default int     `yaml:"default"`
	Paths   []*Path `yaml:"paths"`
}

// Path defines path privileges
type Path struct {
	Path       string `yaml:"path"`
	Method     string `yaml:"method"`
	Privileges int    `yaml:"privileges"`
}

// ID returns proxy target's unique ID
func (t *TargetConfig) ID() string {
	return t.TID
}

// Type returns proxy target's type
func (t *TargetConfig) Type() TargetType {
	return t.TargetType
}

// UpdateToken defines wether a token needs to be updated during each request
func (t *TargetConfig) UpdateToken() bool {
	return t.UpdatesToken
}

// URI returns proxy target's URI
func (t *TargetConfig) URI() *url.URL {
	return t.uri
}

// Protocol returns proxy target's protocol
func (t *TargetConfig) Protocol() ProtocolType {
	return t.TargetProtocol
}

// PrivilegesForPath returns privileges for a given path. If there is no specific settings, default target privileges are returned.
func (t *TargetConfig) PrivilegesForPath(path, method string) int {
	for _, p := range (*t.Privileges).Paths {
		if p.Path == path && p.Method == method {
			return p.Privileges
		}
	}
	return t.Privileges.Default
}
