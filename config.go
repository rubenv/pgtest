package pgtest

type PGConfig struct {
	BinDir       string
	Dir          string
	IsPersistent bool
}

func NewConfig() *PGConfig {
	return &PGConfig{
		BinDir:       "",
		Dir:          "",
		IsPersistent: false,
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
