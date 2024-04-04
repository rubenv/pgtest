package pgtest_test

import (
	"os"
	"testing"

	"github.com/rubenv/pgtest"
	"github.com/stretchr/testify/assert"
)

func TestPostgreSQL(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	pg, err := pgtest.Start()
	assert.NoError(err)
	assert.NotNil(pg)

	_, err = pg.DB.Exec("CREATE TABLE test (val text)")
	assert.NoError(err)

	err = pg.Stop()
	assert.NoError(err)
}

func TestPostgreSQLWithConfig(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	pg, err := pgtest.New().From("/usr/bin/").Start()
	assert.NoError(err)
	assert.NotNil(pg)

	_, err = pg.DB.Exec("CREATE TABLE test (val text)")
	assert.NoError(err)

	assert.NotEmpty(pg.Host)
	assert.NotEmpty(pg.Name)

	err = pg.Stop()
	assert.NoError(err)
}

func TestPersistent(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	dir, err := os.MkdirTemp("", "pgtest")
	assert.NoError(err)
	defer os.RemoveAll(dir)

	pg, err := pgtest.StartPersistent(dir)
	assert.NoError(err)
	assert.NotNil(pg)

	_, err = pg.DB.Exec("CREATE TABLE test (val text)")
	assert.NoError(err)

	_, err = pg.DB.Exec("INSERT INTO test VALUES ('foo')")
	assert.NoError(err)

	err = pg.Stop()
	assert.NoError(err)

	// Open it again
	pg, err = pgtest.StartPersistent(dir)
	assert.NoError(err)
	assert.NotNil(pg)

	var val string
	err = pg.DB.QueryRow("SELECT val FROM test").Scan(&val)
	assert.NoError(err)
	assert.Equal(val, "foo")

	err = pg.Stop()
	assert.NoError(err)
}

func TestAdditionalArgs(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	pg, err := pgtest.New().WithAdditionalArgs("-c", "wal_level=logical").Start()
	assert.NoError(err)
	assert.NotNil(pg)

	//Check if the wal_level is set to logical
	var walLevel string
	err = pg.DB.QueryRow("SHOW wal_level").Scan(&walLevel)
	assert.NoError(err)
	assert.Equal(walLevel, "logical")

	err = pg.Stop()
	assert.NoError(err)
}
