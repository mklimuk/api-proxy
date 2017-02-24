package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type GatekeeperTestSuite struct {
	suite.Suite
	router    *gin.Engine
	s         Target
	keeper    GatekeeperMock
	serv      *httptest.Server
	url       *url.URL
	authorize bool
	perm      int
}

func (suite *GatekeeperTestSuite) SetupSuite() {
	log.SetLevel(log.DebugLevel)
	suite.router = gin.New()
	suite.router.POST("/token/check", suite.fakeAuth)
	suite.serv = httptest.NewServer(suite.router)
	suite.url, _ = url.Parse(suite.serv.URL)
}

func (suite *GatekeeperTestSuite) TearDownSuite() {
	suite.serv.Close()
}

func (suite *GatekeeperTestSuite) TestEmptyToken() {
	a := assert.New(suite.T())
	k := NewGatekeeper(suite.url)
	t, err := k.CheckAccess("", 0, true)
	a.NoError(err)
	a.Equal("", t)
	t, err = k.CheckAccess("", 3, true)
	a.Error(err)
	a.Equal("", t)
}

func (suite *GatekeeperTestSuite) TestHappyPath() {
	a := assert.New(suite.T())
	k := NewGatekeeper(suite.url)
	suite.perm = 7
	suite.authorize = true
	t, err := k.CheckAccess("test", 5, true)
	a.NoError(err)
	a.Equal("updated", t)
}

func (suite *GatekeeperTestSuite) TestTooLowPrivileges() {
	a := assert.New(suite.T())
	k := NewGatekeeper(suite.url)
	suite.authorize = true
	suite.perm = 3
	t, err := k.CheckAccess("test", 5, true)
	a.Error(err)
	a.Equal("updated", t)
}

func (suite *GatekeeperTestSuite) TestUnauthorized() {
	a := assert.New(suite.T())
	k := NewGatekeeper(suite.url)
	suite.authorize = false
	suite.perm = 3
	t, err := k.CheckAccess("test", 5, true)
	a.Error(err)
	a.Equal("test", t)
}

func TestGatekeeperTestSuite(t *testing.T) {
	suite.Run(t, new(GatekeeperTestSuite))
}

func (suite *GatekeeperTestSuite) fakeAuth(ctx *gin.Context) {
	a := new(checkToken)
	defer ctx.Request.Body.Close()
	json.NewDecoder(ctx.Request.Body).Decode(a)
	a.Claims = claims{Permissions: suite.perm}
	if a.Update {
		a.Token = "updated"
	}
	if !suite.authorize {
		ctx.JSON(http.StatusUnauthorized, a)
		return
	}
	ctx.JSON(http.StatusOK, a)
}
