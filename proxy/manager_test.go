package proxy

import (
	"net/http"
	"testing"

	log "github.com/Sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ManagerTestSuite struct {
	suite.Suite
}

func (suite *ManagerTestSuite) SetupSuite() {
	log.SetLevel(log.DebugLevel)
}

func (suite *ManagerTestSuite) TearDownSuite() {

}

func (suite *ManagerTestSuite) TestConstructor() {
	k := &GatekeeperMock{}
	targets := []*TargetConfig{
		&TargetConfig{TargetType: TypeSingle, TargetProtocol: ProtocolHTTP, TID: "t1", URL: "http://t1.com"},
		&TargetConfig{TargetType: TypePool, TargetProtocol: ProtocolWebsocket, TID: "t2", URL: "http://t2.com"},
		&TargetConfig{TargetType: TypeSingle, TargetProtocol: ProtocolHTTP, TID: "t3", URL: "http://t3.com"},
	}
	m := NewTargetsManager(targets, k)
	a := assert.New(suite.T())
	a.Len(m.(*targetsManager).targets, 3)
}

func (suite *ManagerTestSuite) TestCreatePool() {
	k := &GatekeeperMock{}
	targets := []*TargetConfig{}
	m := NewTargetsManager(targets, k)
	c := &TargetConfig{TargetType: TypePool, TargetProtocol: ProtocolWebsocket, TID: "t2", URL: "http://t2.com"}
	m.CreatePool(c)
	a := assert.New(suite.T())
	p, err := m.(*targetsManager).getPool("t2")
	a.NoError(err)
	a.NotNil(p)
}

func (suite *ManagerTestSuite) TestAddToPool() {
	k := &GatekeeperMock{}
	targets := []*TargetConfig{
		&TargetConfig{TargetType: TypePool, TargetProtocol: ProtocolHTTP, TID: "t2"},
		&TargetConfig{TargetType: TypePool, TargetProtocol: ProtocolWebsocket, TID: "t1"},
	}
	m := NewTargetsManager(targets, k)
	m.AddToPool("t2", "p1", "http://p1.com")
	m.AddToPool("t2", "p2", "http://p2.com")
	m.AddToPool("t1", "ws1", "http://ws1.com")
	a := assert.New(suite.T())
	a.Len(m.(*targetsManager).targets["t2"].(*pool).rp, 2)
}

func (suite *ManagerTestSuite) TestDeleteFromPool() {
	p := NewPool(&TargetConfig{TargetType: TypePool, TargetProtocol: ProtocolHTTP, TID: "t2"})
	f := &fakeHandler{}
	p.(*pool).rp = map[string]http.Handler{
		"r1": f,
		"r2": f,
	}
	m := &targetsManager{targets: map[string]Target{
		"t2": p,
	}}
	a := assert.New(suite.T())
	a.Len(m.targets["t2"].(*pool).rp, 2)
	m.RemoveFromPool("t2", "r1")
	a.Len(m.targets["t2"].(*pool).rp, 1)
	a.NotNil(m.targets["t2"].(*pool).rp["r2"])
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}

type fakeHandler struct{}

func (f *fakeHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {}
