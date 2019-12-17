package pgtest

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	_ "github.com/lib/pq"
)

type PG struct {
	dir string
	cmd *exec.Cmd
	DB  *sql.DB
}

// Start a new PostgreSQL database, on temporary storage.
//
// This database has fsync disabled for performance, so it might run faster
// than your production database. This makes it less reliable in case of system
// crashes, but we don't care about that anyway during unit testing.
//
// Use the DB field to access the database connection
func Start() (*PG, error) {
	// Prepare data directory
	dir, err := ioutil.TempDir("", "pgtest")
	if err != nil {
		return nil, err
	}

	dataDir := path.Join(dir, "data")
	sockDir := path.Join(dir, "sock")

	err = os.MkdirAll(dataDir, 0700)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(sockDir, 0700)
	if err != nil {
		return nil, err
	}

	// Find executables root path
	binPath, err := findBinPath()
	if err != nil {
		return nil, err
	}

	// Initialize PostgreSQL data directory
	init := exec.Command(path.Join(binPath, "initdb"),
		"-D", dataDir,
		"--no-sync",
	)
	err = init.Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize DB: %w", err)
	}

	// Start PostgreSQL
	cmd := exec.Command(path.Join(binPath, "postgres"),
		"-D", dataDir, // Data directory
		"-k", sockDir, // Location for the UNIX socket
		"-h", "", // Disable TCP listening
		"-F", // No fsync, just go fast
	)

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("Failed to start PostgreSQL: %w", err)
	}

	// Connect to DB
	dsn := makeDSN(sockDir, "postgres")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Prepare test database
	r := retrier.New(retrier.ConstantBackoff(1000, 10*time.Millisecond), nil)
	err = r.Run(func() error {
		_, err := db.Exec("CREATE DATABASE test")
		return err
	})
	if err != nil {
		return nil, err
	}

	err = db.Close()
	if err != nil {
		return nil, err
	}

	// Connect to it properly
	dsn = makeDSN(sockDir, "test")
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	pg := &PG{
		cmd: cmd,
		dir: dir,

		DB: db,
	}

	return pg, nil
}

// Stop the database and remove storage files.
func (p *PG) Stop() error {
	defer func() {
		// Always try to remove it
		os.RemoveAll(p.dir)
	}()

	err := p.cmd.Process.Signal(os.Interrupt)
	if err != nil {
		return err
	}

	err = p.cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

// Needed because Ubuntu doesn't put initdb in $PATH
func findBinPath() (string, error) {
	// In $PATH (e.g. Fedora) great!
	p, err := exec.LookPath("initdb")
	if err == nil {
		return path.Dir(p), nil
	}

	// Look for a PostgreSQL in one of the folders Ubuntu uses
	folders := []string{
		"/usr/lib/postgresql/",
	}
	for _, folder := range folders {
		f, err := os.Stat(folder)
		if os.IsNotExist(err) {
			continue
		}
		if !f.IsDir() {
			continue
		}

		files, err := ioutil.ReadDir(folder)
		if err != nil {
			return "", err
		}
		for _, fi := range files {
			if !fi.IsDir() {
				continue
			}

			binPath := path.Join(folder, fi.Name(), "bin")
			_, err := os.Stat(path.Join(binPath, "initdb"))
			if err == nil {
				return binPath, nil
			}
		}
	}

	return "", fmt.Errorf("Did not find PostgreSQL executables installed")
}

func makeDSN(sockDir, dbname string) string {
	return fmt.Sprintf("host=%s dbname=%s", sockDir, dbname)
}
