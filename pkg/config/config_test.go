package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigNewYAML(t *testing.T) {
	cfg, err := New("testdata/base.yaml")
	assert.NoError(t, err)

	assert.Equal(t, "/home/test/.distillery", cfg.Path)
	assert.Equal(t, "/home/test/.cache", cfg.CachePath)

	aliases := &Aliases{
		"dist": &Alias{
			Name:    "ekristen/distillery",
			Version: "latest",
		},
		"aws-nuke": &Alias{
			Name:    "ekristen/aws-nuke",
			Version: "3.29.3",
		},
	}

	assert.EqualValues(t, aliases, cfg.Aliases)
}

func TestConfigNewTOML(t *testing.T) {
	cfg, err := New("testdata/base.toml")
	assert.NoError(t, err)

	assert.Equal(t, "/home/test/.distillery", cfg.Path)
	assert.Equal(t, "/home/test/.cache", cfg.CachePath)

	aliases := &Aliases{
		"dist": &Alias{
			Name:    "ekristen/distillery",
			Version: "latest",
		},
		"aws-nuke": &Alias{
			Name:    "ekristen/aws-nuke",
			Version: "3.29.3",
		},
	}

	assert.EqualValues(t, aliases, cfg.Aliases)
}
