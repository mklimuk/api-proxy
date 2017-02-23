package proxy

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

//TargetsManagerMock is a mock of the TargetsManager interface
type TargetsManagerMock struct {
	mock.Mock
}

//AddToPool is a mocked method
func (m *TargetsManagerMock) AddToPool(poolID, ID, targetURI string) error {
	args := m.Called(poolID, ID, targetURI)
	return args.Error(0)
}

//RemoveFromPool is a mocked method
func (m *TargetsManagerMock) RemoveFromPool(poolID, ID string) error {
	args := m.Called(poolID, ID)
	return args.Error(0)
}

//CreatePool is a mocked method
func (m *TargetsManagerMock) CreatePool(conf *TargetConfig) error {
	args := m.Called(conf)
	return args.Error(0)
}

//Proxy is a mocked method
func (m *TargetsManagerMock) Proxy(ctx *gin.Context) {
	m.Called(ctx)
}

//GatekeeperMock is a mock of the Gatekeeper interface
type GatekeeperMock struct {
	mock.Mock
}

//CheckAccess is a mocked method
func (m *GatekeeperMock) CheckAccess(token string, accessPrivileges int, updateToken bool) (string, error) {
	args := m.Called(token, accessPrivileges, updateToken)
	return args.String(0), args.Error(1)
}
