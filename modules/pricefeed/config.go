package pricefeed

import (
	"gopkg.in/yaml.v3"

	"github.com/elesto-dao/bdjuno/types"
)

// Config contains the configuration about the pricefeed module
type Config struct {
	Tokens []types.Token `yaml:"tokens"`
}

// NewConfig returns a new Config instance
func NewConfig(tokens []types.Token) *Config {
	return &Config{
		Tokens: tokens,
	}
}

func ParseConfig(bz []byte) (*Config, error) {
	type T struct {
		Config *Config `yaml:"pricefeed"`
	}
	var cfg T
	err := yaml.Unmarshal(bz, &cfg)
	return cfg.Config, err
}
