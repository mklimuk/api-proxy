package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/mklimuk/api-proxy/proxy"
	"github.com/mklimuk/auth/config"
	"github.com/mklimuk/goerr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	suite.Suite
	router *gin.Engine
	p      proxy.TargetsManagerMock
	serv   *httptest.Server
}

func (suite *APITestSuite) SetupSuite() {
	log.SetLevel(log.DebugLevel)
	suite.p = proxy.TargetsManagerMock{}
	p := NewProxyAPI(&suite.p)
	c := NewControlAPI()
	suite.router = gin.New()
	p.AddRoutes(suite.router)
	c.AddRoutes(suite.router)
	suite.serv = httptest.NewServer(suite.router)
}

func (suite *APITestSuite) TearDownSuite() {
	suite.serv.Close()
}

func (suite *APITestSuite) TestHealth() {
	a := assert.New(suite.T())
	res, err := http.Get(fmt.Sprintf("%s%s", suite.serv.URL, "/health"))
	a.NoError(err)
	a.Equal(http.StatusOK, res.StatusCode)
}

func (suite *APITestSuite) TestVersion() {
	a := assert.New(suite.T())
	config.Ver = config.Version{Version: "0.1.0"}
	res, err := http.Get(fmt.Sprintf("%s%s", suite.serv.URL, "/version"))
	a.NoError(err)
	a.Equal(http.StatusOK, res.StatusCode)
}

func (suite *APITestSuite) TestCreatePool() {
	a := assert.New(suite.T())
	// test no body (parse error)
	res, err := http.Post(fmt.Sprintf("%s%s", suite.serv.URL, "/pool"), "application/x.pool.req+json", nil)
	a.NoError(err)
	a.Equal(http.StatusBadRequest, res.StatusCode)
	req := &proxy.TargetConfig{TID: "test", TargetProtocol: proxy.ProtocolHTTP}
	var b []byte
	b, err = json.Marshal(&req)
	a.NoError(err)
	// test unauthorized
	suite.p.On("CreatePool", mock.AnythingOfType("*proxy.TargetConfig")).Return(goerr.NewError("conflict", proxy.Conflict)).Once()
	res, err = http.Post(fmt.Sprintf("%s%s", suite.serv.URL, "/pool"), "application/x.pool.req+json", bytes.NewReader(b))
	a.NoError(err)
	a.Equal(http.StatusConflict, res.StatusCode)
	// test internal error
	suite.p.On("CreatePool", mock.AnythingOfType("*proxy.TargetConfig")).Return(errors.New("dummy")).Once()
	res, err = http.Post(fmt.Sprintf("%s%s", suite.serv.URL, "/pool"), "application/x.pool.req+json", bytes.NewReader(b))
	a.NoError(err)
	a.Equal(http.StatusInternalServerError, res.StatusCode)
	// happy path
	suite.p.On("CreatePool", mock.AnythingOfType("*proxy.TargetConfig")).Return(nil).Once()
	res, err = http.Post(fmt.Sprintf("%s%s", suite.serv.URL, "/pool"), "application/x.pool.req+json", bytes.NewReader(b))
	a.NoError(err)
	a.Equal(http.StatusOK, res.StatusCode)
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
