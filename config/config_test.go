package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func (suite *ConfigTestSuite) TestParseConfig() {
	a := assert.New(suite.T())
	Parse("test/config.yml")
	a.Len(Config.Targets, 1)
	a.Len(Config.Targets[0].Methods, 1)
	a.Equal("GET", Config.Targets[0].Methods[0])
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
