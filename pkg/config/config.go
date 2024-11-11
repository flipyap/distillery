package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/pelletier/go-toml/v2"

	"github.com/ekristen/distillery/pkg/common"
)

type Config struct {
	// Path - path to store the configuration files, this path is set by default based on the operating system type
	// and your user's home directory. Typically, this is set to $HOME/.distillery
	Path string `yaml:"path" toml:"path"`

	// BinPath - path to create symlinks for your binaries, this path is set by default based on the operating system type
	// This is the path that is added to your PATH environment variable. Typically, this is set to $HOME/.distillery/bin
	// This allows you to override the location for symlinks. For example, you can instead put them all in /usr/local/bin
	BinPath string `yaml:"bin_path" toml:"bin_path"`

	// CachePath - path to store cache files, this path is set by default based on the operating system type
	CachePath string `yaml:"cache_path" toml:"cache_path"`

	// DefaultSource - the default source to use when installing binaries, this defaults to GitHub
	DefaultSource string `yaml:"default_source" toml:"default_source"`

	// AutomaticAliases - automatically create aliases for any binary that is installed
	AutomaticAliases bool `yaml:"automatic_aliases" toml:"automatic_aliases"`

	// Aliases - Allow for creating shorthand aliases for source locations that you use frequently. A good example
	// of this is `distillery` -> `ekristen/distillery`
	Aliases *Aliases `yaml:"aliases" toml:"aliases"`

	// Language - the language to use for the output of the application
	Language string `yaml:"language" toml:"language"`
}

func (c *Config) GetCachePath() string {
	return filepath.Join(c.CachePath, common.NAME)
}

func (c *Config) GetMetadataPath() string {
	return filepath.Join(c.CachePath, common.NAME, "metadata")
}

func (c *Config) GetDownloadsPath() string {
	return filepath.Join(c.CachePath, common.NAME, "downloads")
}

func (c *Config) GetOptPath() string {
	return filepath.Join(c.Path, "opt")
}

func (c *Config) GetAlias(name string) *Alias {
	if c.Aliases == nil {
		return nil
	}

	for short, alias := range *c.Aliases {
		if short == name {
			return alias
		}
	}

	return nil
}

func (c *Config) MkdirAll() error {
	paths := []string{c.BinPath, c.GetOptPath(), c.CachePath, c.GetMetadataPath(), c.GetDownloadsPath()}

	for _, path := range paths {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

// Load - load the configuration file
func (c *Config) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if strings.HasSuffix(path, ".yaml") {
		return yaml.Unmarshal(data, c)
	} else if strings.HasSuffix(path, ".toml") {
		return toml.Unmarshal(data, c)
	}

	return fmt.Errorf("unknown configuration file suffix")
}

// New - create a new configuration object
func New(path string) (*Config, error) {
	cfg := &Config{}
	if err := cfg.Load(path); err != nil {
		return cfg, err
	}

	if cfg.Language == "" {
		cfg.Language = "en"
	}

	if cfg.DefaultSource == "" {
		cfg.DefaultSource = "github"
	}

	if cfg.Path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return cfg, err
		}
		cfg.Path = filepath.Join(homeDir, fmt.Sprintf(".%s", common.NAME))
	}

	if cfg.CachePath == "" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return cfg, err
		}
		cfg.CachePath = cacheDir
	}

	if cfg.BinPath == "" {
		cfg.BinPath = filepath.Join(cfg.Path, "bin")
	}

	return cfg, nil
}
