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