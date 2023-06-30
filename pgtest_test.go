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

	dir, err := ioutil.TempDir("", "pgtest")
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
