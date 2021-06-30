package pgtest

type PGConfig struct {
	Folder     string
	Dir        string
	Persistent bool
}

func NewConfig() *PGConfig {
	return &PGConfig{
		Folder:     "",
		Dir:        "",
		Persistent: false,
	}
}

func (c *PGConfig) WithPersistent() *PGConfig {
	c.Persistent = true
	return c
}

func (c *PGConfig) WithBinDir(dir string) *PGConfig {
	c.Folder = dir
	return c
}

func (c *PGConfig) WithDataDir(dir string) *PGConfig {
	c.Dir = dir
	return c
}
