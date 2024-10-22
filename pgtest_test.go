package pgtest_test

import (
	"database/sql"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
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

func TestWrongDbNameAndPassword(t *testing.T) {
	testDbWithUserNameAndPassword(t, "wrongdbName", "wrongPassword", false)
}

func TestWrongDbName(t *testing.T) {
	testDbWithUserNameAndPassword(t, "wrongdbName", "correctpassword", false)
}

func TestWrongDbPassword(t *testing.T) {
	testDbWithUserNameAndPassword(t, "correctdbname", "wrongpassword", false)
}

func TestCorrectCredentials(t *testing.T) {
	testDbWithUserNameAndPassword(t, "correctdbname", "correctpassword", true)
}

// util functions for the dbname/password tests
func testDbWithUserNameAndPassword(t *testing.T, databaseName, password string, assertErrorNil bool) {
	t.Parallel()

	assert := assert.New(t)

	pg, err := pgtest.New().SetDbName("correctdbname").SetPassword("correctpassword").Start()
	assert.NoError(err)
	assert.NotNil(pg)

	// connect using username and password via a different connection
	// using the sockDir.
	dsn := makeDsn(getSockDir(pg, t), databaseName, password)
	// not testing the error returned by Open, because
	// sometimes it returns without connecting.
	// so we use .Ping to get the actual error.
	connection, _ := sql.Open("postgres", dsn)
	err = connection.Ping()
	if assertErrorNil {
		assert.NoError(err)
	} else {
		assert.Error(err)
	}

	err = connection.Close()
	assert.NoError(err)

	err = pg.Stop()
	assert.NoError(err)
}

func pgUser() string {
	currentUser, err := user.Current()
	isRoot := currentUser.Username == "root"
	if isRoot {
		return "postgres"
	}
	if err != nil {
		return "postgres" // fallback to postgres if we can't get the current user
	}
	return currentUser.Username
}

func makeDsn(sockDir, dbname, password string) string {
	dsnUser := ""
	dsnPassword := ""
	user := pgUser()
	// add user if defined
	if user != "" {
		dsnUser = fmt.Sprintf("user=%s", user)
	}
	// add password if defined
	if password != "" {
		dsnPassword = fmt.Sprintf("password=%s", password)
	}
	return fmt.Sprintf("host=%s dbname=%s %s %s", sockDir, dbname, dsnUser, dsnPassword)
}

func getSockDir(pg *pgtest.PG, t *testing.T) string {
	// Use reflection to access the private 'dir' field
	pgValue := reflect.ValueOf(pg).Elem()
	dirField := pgValue.FieldByName("dir")
	if !dirField.IsValid() {
		t.Fatal("Unable to find 'dir' field in PostgreSQL struct")
	}

	dbRoot := dirField.String()

	return filepath.Join(dbRoot, "sock")
}
