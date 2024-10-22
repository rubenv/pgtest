package pgtest_test

import (
	"testing"

	"github.com/rubenv/pgtest"
	"github.com/stretchr/testify/assert"
)

func TestPGConfig(t *testing.T) {
	assert := assert.New(t)

	config := pgtest.New().From("/usr/bin").DataDir("/tmp/data").Persistent().EnableFSync().SetDbName("mydb").SetPassword("mypassword").WithAdditionalArgs("-c", "log_statement=all")

	assert.True(config.IsPersistent)
	assert.EqualValues("/tmp/data", config.Dir)
	assert.EqualValues("/usr/bin", config.BinDir)
	assert.EqualValues([]string{"-c", "log_statement=all"}, config.AdditionalArgs)
	assert.EqualValues(true, config.FSync)
	assert.EqualValues("mydb", config.DbName)
	assert.EqualValues("mypassword", config.Password)
}
