package pgtest_test

import (
	"testing"

	"github.com/rubenv/pgtest"
	"github.com/stretchr/testify/assert"
)

func TestPGConfig(t *testing.T) {
	assert := assert.New(t)

	config := pgtest.New().From("/usr/bin").DataDir("/tmp/data").Persistent()

	assert.True(config.IsPersistent)
	assert.EqualValues("/tmp/data", config.Dir)
	assert.EqualValues("/usr/bin", config.BinDir)
}
