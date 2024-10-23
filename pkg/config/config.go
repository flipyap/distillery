package config

type Config struct {
	BinDir  string            `yaml:"bin_dir"`
	Aliases map[string]string `yaml:"aliases"`
}
