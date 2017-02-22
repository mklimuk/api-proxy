package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UtilsTestSuite struct {
	suite.Suite
}

func (suite *UtilsTestSuite) SetupSuite() {
}

func (suite *UtilsTestSuite) TearDownSuite() {
}

func (suite *UtilsTestSuite) TestExtractToken() {
	a := assert.New(suite.T())
	a.Equal("abc", extractToken("Bearer abc"))
	a.Equal("abc", extractToken("abc"))
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}
