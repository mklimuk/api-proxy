package proxy

import (
	"net/http"
	"net/url"
	"regexp"

	log "github.com/Sirupsen/logrus"
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

const maxPrivileges = 100

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
	Keeper() Gatekeeper
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
	Exact       string `yaml:"exact"`
	Regex       string `yaml:"regex"`
	Method      string `yaml:"method"`
	Privileges  int    `yaml:"privileges"`
	parsedRegex *regexp.Regexp
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

// Keeper returns proxy target's gatekeeper
func (t *TargetConfig) Keeper() Gatekeeper {
	return t.keeper
}

// PrivilegesForPath returns privileges for a given path. If there is no specific settings, default target privileges are returned.
func (t *TargetConfig) PrivilegesForPath(path, method string) int {
	for _, p := range (*t.Privileges).Paths {
		if p.Method == method {
			if p.Exact == path {
				return p.Privileges
			}
			var match bool
			var err error
			if match, err = t.matchRegex(p, path); err != nil {
				log.WithFields(log.Fields{"target": t.ID(), "path": p.Regex}).
					WithError(err).Error("Error parsing regex for path")
				return maxPrivileges
			}
			if match {
				return p.Privileges
			}
		}
	}
	return t.Privileges.Default
}

func (t *TargetConfig) matchRegex(path *Path, toCheck string) (bool, error) {
	if path.Regex == "" {
		return false, nil
	}
	if path.parsedRegex == nil {
		var err error
		if path.parsedRegex, err = regexp.Compile(path.Regex); err != nil {
			return false, err
		}
	}
	return path.parsedRegex.MatchString(toCheck), nil
}

func checkAuthAndServe(t Target, path string, rp http.Handler, ctx *gin.Context) {
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
		if token, err = t.Keeper().CheckAccess(extractToken(h), condition, t.UpdateToken()); err != nil {
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

	rp.ServeHTTP(ctx.Writer, ctx.Request)
}
