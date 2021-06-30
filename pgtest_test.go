package pgtest_test

import (
	"io/ioutil"
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

func TestPostgreSQLFromPath(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	config := pgtest.NewConfig().From("/usr/bin/")
	pg, err := pgtest.Start(config)
	assert.NoError(err)
	assert.NotNil(pg)

	_, err = pg.DB.Exec("CREATE TABLE test (val text)")
	assert.NoError(err)

	err = pg.Stop()
	assert.NoError(err)
}

func TestPersistent(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	dir, err := ioutil.TempDir("", "pgtest")
	assert.NoError(err)
	defer os.RemoveAll(dir)
	config := pgtest.NewConfig().Persistent().DataDir(dir)
	pg, err := pgtest.Start(config)
	assert.NoError(err)
	assert.NotNil(pg)

	_, err = pg.DB.Exec("CREATE TABLE test (val text)")
	assert.NoError(err)

	_, err = pg.DB.Exec("INSERT INTO test VALUES ('foo')")
	assert.NoError(err)

	err = pg.Stop()
	assert.NoError(err)

	// Open it again
	pg, err = pgtest.Start(config)
	assert.NoError(err)
	assert.NotNil(pg)

	var val string
	err = pg.DB.QueryRow("SELECT val FROM test").Scan(&val)
	assert.NoError(err)
	assert.Equal(val, "foo")

	err = pg.Stop()
	assert.NoError(err)
}
