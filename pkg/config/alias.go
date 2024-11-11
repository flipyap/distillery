package config

import (
	"strings"

	"github.com/ekristen/distillery/pkg/common"
)

type Aliases map[string]*Alias

type Alias struct {
	Name    string `yaml:"name" toml:"name"`
	Version string `yaml:"version" toml:"version"`
	Bin     string `yaml:"bin" toml:"bin"`
}

func (a *Alias) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string
	if unmarshal(&value) == nil {
		p := strings.Split(value, "@")
		a.Name = p[0]
		a.Version = common.Latest
		if len(p) > 1 {
			a.Version = p[1]
		}
		return nil
	}

	type alias Alias
	aux := (*alias)(a)
	if err := unmarshal(aux); err != nil {
		return err
	}

	return nil
}

func (a *Alias) UnmarshalText(b []byte) error {
	*a = Alias{
		Name:    string(b),
		Version: "latest",
	}
	return nil
}
