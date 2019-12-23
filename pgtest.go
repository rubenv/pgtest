// Spawns a PostgreSQL server with a single database configured. Ideal for unit
// tests where you want a clean instance each time. Then clean up afterwards.
//
// Requires PostgreSQL to be installed on your system (but it doesn't have to be running).
package pgtest

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type PG struct {
	dir string
	cmd *exec.Cmd
	DB  *sql.DB

	stderr io.ReadCloser
	stdout io.ReadCloser
}

// Start a new PostgreSQL database, on temporary storage.
//
// This database has fsync disabled for performance, so it might run faster
// than your production database. This makes it less reliable in case of system
// crashes, but we don't care about that anyway during unit testing.
//
// Use the DB field to access the database connection
func Start() (*PG, error) {
	// Handle dropping permissions when running as root
	me, err := user.Current()
	if err != nil {
		return nil, err
	}
	isRoot := me.Username == "root"

	pgUID := int(0)
	pgGID := int(0)
	if isRoot {
		pgUser, err := user.Lookup("postgres")
		if err != nil {
			return nil, fmt.Errorf("Could not find postgres user, which is required when running as root: %s", err)
		}

		uid, err := strconv.ParseInt(pgUser.Uid, 10, 64)
		if err != nil {
			return nil, err
		}
		pgUID = int(uid)

		gid, err := strconv.ParseInt(pgUser.Gid, 10, 64)
		if err != nil {
			return nil, err
		}
		pgGID = int(gid)
	}

	// Prepare data directory
	dir, err := ioutil.TempDir("", "pgtest")
	if err != nil {
		return nil, err
	}

	dataDir := path.Join(dir, "data")
	sockDir := path.Join(dir, "sock")

	err = os.MkdirAll(dataDir, 0711)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(sockDir, 0711)
	if err != nil {
		return nil, err
	}

	if isRoot {
		err = os.Chmod(dir, 0711)
		if err != nil {
			return nil, err
		}

		err = os.Chown(dataDir, pgUID, pgGID)
		if err != nil {
			return nil, err
		}

		err = os.Chown(sockDir, pgUID, pgGID)
		if err != nil {
			return nil, err
		}
	}

	// Find executables root path
	binPath, err := findBinPath()
	if err != nil {
		return nil, err
	}

	// Initialize PostgreSQL data directory
	init := prepareCommand(isRoot, path.Join(binPath, "initdb"),
		"-D", dataDir,
		"--no-sync",
	)
	out, err := init.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize DB: %w -> %s", err, string(out))
	}

	// Start PostgreSQL
	cmd := prepareCommand(isRoot, path.Join(binPath, "postgres"),
		"-D", dataDir, // Data directory
		"-k", sockDir, // Location for the UNIX socket
		"-h", "", // Disable TCP listening
		"-F", // No fsync, just go fast
	)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stderr.Close()
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		stderr.Close()
		stdout.Close()
		return nil, fmt.Errorf("Failed to start PostgreSQL: %w", err)
	}

	// Connect to DB
	dsn := makeDSN(sockDir, "postgres")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		serr, _ := ioutil.ReadAll(stderr)
		sout, _ := ioutil.ReadAll(stdout)
		stderr.Close()
		stdout.Close()
		return nil, fmt.Errorf("Failed to connect to DB: %s\nOUT: %s\nERR: %s", err, string(sout), string(serr))
	}

	// Prepare test database
	err = retry(func() error {
		_, err := db.Exec("CREATE DATABASE test")
		return err
	}, 1000, 10*time.Millisecond)
	if err != nil {
		serr, _ := ioutil.ReadAll(stderr)
		sout, _ := ioutil.ReadAll(stdout)
		stderr.Close()
		stdout.Close()
		return nil, fmt.Errorf("Failed to initialize DB: %s\nOUT: %s\nERR: %s", err, string(sout), string(serr))
	}

	err = db.Close()
	if err != nil {
		stderr.Close()
		stdout.Close()
		return nil, err
	}

	// Connect to it properly
	dsn = makeDSN(sockDir, "test")
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		stderr.Close()
		stdout.Close()
		return nil, err
	}

	pg := &PG{
		cmd: cmd,
		dir: dir,

		DB: db,

		stderr: stderr,
		stdout: stdout,
	}

	return pg, nil
}

// Stop the database and remove storage files.
func (p *PG) Stop() error {
	if p == nil {
		return nil
	}

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

	if p.stderr != nil {
		p.stderr.Close()
	}

	if p.stdout != nil {
		p.stdout.Close()
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

func retry(fn func() error, attempts int, interval time.Duration) error {
	for {
		err := fn()
		if err == nil {
			return nil
		}

		attempts -= 1
		if attempts <= 0 {
			return err
		}

		time.Sleep(interval)
	}
}

func prepareCommand(isRoot bool, command string, args ...string) *exec.Cmd {
	if !isRoot {
		return exec.Command(command, args...)
	}

	return exec.Command("su",
		"-",
		"postgres",
		"-c",
		strings.Join(append([]string{command}, args...), " "),
	)
}
