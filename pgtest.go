package pgtest

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	_ "github.com/lib/pq"
)

type PG struct {
	dir string
	cmd *exec.Cmd
	DB  *sql.DB
}

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

	// Initialize PostgreSQL data directory
	init := exec.Command("initdb",
		"-D", dataDir,
		"--no-sync",
	)
	err = init.Run()
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize DB: %w", err)
	}

	// Start PostgreSQL
	main, err := exec.LookPath("postgres")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(main,
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
	dsn := makeDSN(sockDir, "test")
	db, err := sql.Open("postgres", dsn)
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

func makeDSN(sockDir, dbname string) string {
	return fmt.Sprintf("host=%s dbname=%s", sockDir, dbname)
}