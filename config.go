package pgtest

type PGConfig struct {
	BinDir         string   // Directory to look for postgresql binaries including initdb, postgres
	Dir            string   // Directory for storing database files, removed for non-persistent configs
	IsPersistent   bool     // Whether to make the current configuraton persistent or not
	AdditionalArgs []string // Additional arguments to pass to the postgres command
	FSync          bool     // Sets -F flag when false
	DbName         string
	Password       string
}

func New() *PGConfig {
	return &PGConfig{
		BinDir:       "",
		Dir:          "",
		IsPersistent: false,
		FSync:        false, // disable fsync by default - just go fast.
		DbName:       "test",
		Password:     "",
	}
}

func (c *PGConfig) Persistent() *PGConfig {
	c.IsPersistent = true
	return c
}

func (c *PGConfig) From(dir string) *PGConfig {
	c.BinDir = dir
	return c
}

func (c *PGConfig) UseBinariesIn(dir string) *PGConfig {
	c.BinDir = dir
	return c
}

func (c *PGConfig) DataDir(dir string) *PGConfig {
	c.Dir = dir
	return c
}

func (c *PGConfig) WithAdditionalArgs(args ...string) *PGConfig {
	c.AdditionalArgs = args
	return c
}

func (c *PGConfig) EnableFSync() *PGConfig {
	c.FSync = true
	return c
}

func (c *PGConfig) SetPassword(password string) *PGConfig {
	c.Password = password
	return c
}

func (c *PGConfig) SetDbName(dbName string) *PGConfig {
	c.DbName = dbName
	return c
}

func (c *PGConfig) Start() (*PG, error) {
	return start(c)
}
