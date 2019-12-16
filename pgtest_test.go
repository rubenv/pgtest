package pgtest_test

import (
	"testing"

	"github.com/rubenv/pgtest"
	"github.com/stretchr/testify/assert"
)

func TestPostgreSQL(t *testing.T) {
	assert := assert.New(t)

	pg, err := pgtest.Start()
	assert.NoError(err)
	assert.NotNil(pg)

	err = pg.Stop()
	assert.NoError(err)
}
