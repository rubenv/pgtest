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

	_, err = pg.DB.Exec("CREATE TABLE test (val text)")
	assert.NoError(err)

	err = pg.Stop()
	assert.NoError(err)
}
