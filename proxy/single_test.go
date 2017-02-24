package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/mklimuk/goerr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SingleTestSuite struct {
	suite.Suite
	router *gin.Engine
	s      Target
	keeper GatekeeperMock
	serv   *httptest.Server
}

func (suite *SingleTestSuite) SetupSuite() {
	log.SetLevel(log.DebugLevel)
	suite.router = gin.New()
	suite.router.GET("/pong/*rest", testHandler)
	suite.router.POST("/pong/*rest", testHandler)
	suite.serv = httptest.NewServer(suite.router)
	p := []*Path{
		&Path{Regex: `\/catalog\/templates\/[^\/\s]*$`, Privileges: 10, Method: http.MethodPost},
		&Path{Exact: `/catalog/templates`, Privileges: 5, Method: http.MethodGet},
	}
	c := &TargetConfig{Privileges: &Privileges{Default: 0, Paths: p}, TID: "test", URL: fmt.Sprintf("%s/pong", suite.serv.URL), TargetProtocol: ProtocolHTTP, TargetType: TypeSingle}
	suite.keeper = GatekeeperMock{}
	c.keeper = &suite.keeper
	suite.s, _ = NewSingle(c)
	suite.router.GET("/api/:id/*path", suite.s.Handler())
	suite.router.POST("/api/:id/*path", suite.s.Handler())
}

func (suite *SingleTestSuite) TearDownSuite() {
	suite.serv.Close()
}

func (suite *SingleTestSuite) TestDefaultPath() {
	a := assert.New(suite.T())
	suite.keeper.On("CheckAccess", "", 0, false).Return("", nil).Once()
	res, err := http.Get(fmt.Sprintf("%s%s", suite.serv.URL, "/api/test/catalog"))
	a.NoError(err)
	a.Equal(http.StatusOK, res.StatusCode)
}

func (suite *SingleTestSuite) TestNoHeader() {
	a := assert.New(suite.T())
	suite.keeper.On("CheckAccess", "", 5, false).Return("", goerr.NewError("unauthorized", goerr.Unauthorized)).Once()
	res, err := http.Get(fmt.Sprintf("%s%s", suite.serv.URL, "/api/test/catalog/templates"))
	a.NoError(err)
	a.Equal(http.StatusUnauthorized, res.StatusCode)
}

func (suite *SingleTestSuite) TestNoPrivileges() {
	a := assert.New(suite.T())
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", suite.serv.URL, "/api/test/catalog/templates"), nil)
	req.Header.Set("Authorization", "testToken")
	suite.keeper.On("CheckAccess", "testToken", 5, false).Return("testTokenRes", goerr.NewError("unauthorized", goerr.Unauthorized)).Once()
	res, err := client.Do(req)
	a.NoError(err)
	a.Equal(http.StatusUnauthorized, res.StatusCode)
}

func (suite *SingleTestSuite) TestProxy() {
	a := assert.New(suite.T())
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", suite.serv.URL, "/api/test/catalog/templates"), nil)
	req.Header.Set("Authorization", "testToken")
	suite.keeper.On("CheckAccess", "testToken", 5, false).Return("testTokenRes", nil).Once()
	res, err := client.Do(req)
	a.NoError(err)
	a.Equal(http.StatusOK, res.StatusCode)
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", suite.serv.URL, "/api/test/catalog/templates/template1"), nil)
	req.Header.Set("Authorization", "testToken")
	suite.keeper.On("CheckAccess", "testToken", 10, false).Return("testTokenRes", nil).Once()
	res, err = client.Do(req)
	a.NoError(err)
	a.Equal(http.StatusOK, res.StatusCode)
}

func TestSingleTestSuite(t *testing.T) {
	suite.Run(t, new(SingleTestSuite))
}

func testHandler(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusOK)
}
