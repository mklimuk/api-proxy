package proxy

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TargetTestSuite struct {
	suite.Suite
}

func (suite *TargetTestSuite) SetupSuite() {
}

func (suite *TargetTestSuite) TearDownSuite() {
}

func (suite *TargetTestSuite) TestInterface() {
	a := assert.New(suite.T())
	c := &TargetConfig{TID: "tid", URL: "http://test.com", UpdatesToken: true, TargetType: TypeSingle, TargetProtocol: ProtocolHTTP}
	t, err := NewSingle(c)
	a.NoError(err)
	a.Equal("tid", t.ID())
	a.True(t.UpdateToken())
	a.Equal(TypeSingle, t.Type())
	a.Equal(ProtocolHTTP, t.Protocol())
	a.Equal("http://test.com", t.URI().String())
	c.URL = `à!!""(weirdOne:`
	c.uri = nil
	t, err = NewSingle(c)
	a.NoError(err)
}

func (suite *TargetTestSuite) TestPrivilegesForPath() {
	a := assert.New(suite.T())
	p := []*Path{
		&Path{Regex: `\/catalog\/templates\/[^\/\s]*$`, Privileges: 10, Method: http.MethodPost},
		&Path{Exact: `/catalog/templates`, Privileges: 5, Method: http.MethodGet},
	}
	c := &TargetConfig{Privileges: &Privileges{Default: 1, Paths: p}, TID: "testowy", URL: "http://test.com", TargetProtocol: ProtocolWebsocket, TargetType: TypeSingle}
	s, err := NewSingle(c)
	a.NoError(err)
	a.NotNil(s)
	a.Equal(1, s.PrivilegesForPath("/catalog/templates", "POST"))
	a.Equal(5, s.PrivilegesForPath("/catalog/templates", "GET"))
	a.Equal(10, s.PrivilegesForPath("/catalog/templates/test123", "POST"))
	a.Equal(1, s.PrivilegesForPath("/catalog/templates/test123", "PUT"))
	a.Equal(1, s.PrivilegesForPath("/different", "DELETE"))
	c.Privileges.Paths = append(c.Privileges.Paths, &Path{Regex: `\/audio\/file\/[^\*$`, Privileges: 7, Method: http.MethodGet})
	a.Equal(5, s.PrivilegesForPath("/catalog/templates", "GET"))
	a.Equal(100, s.PrivilegesForPath("/audio/file/23001", "GET"))
}

func (suite *TargetTestSuite) TestPathMatching() {
	a := assert.New(suite.T())
	p := &Path{Regex: `\/catalog\/templates\/[^\/\s]*$`}
	t := TargetConfig{}
	m, err := t.matchRegex(p, "/catalog/templates")
	a.NoError(err)
	a.False(m)
	m, err = t.matchRegex(p, "/catalog/templates/test")
	a.NoError(err)
	a.True(m)
	// invalid
	p.Regex = `\/catalog\/templates\/[^\/\s*$`
	p.parsedRegex = nil
	m, err = t.matchRegex(p, "/catalog/templates/test")
	a.Error(err)
	a.False(m)
	p.Regex = ``
	p.parsedRegex = nil
	m, err = t.matchRegex(p, "/catalog/templates/test")
	a.NoError(err)
	a.False(m)

}

func TestTargetTestSuite(t *testing.T) {
	suite.Run(t, new(TargetTestSuite))
}
