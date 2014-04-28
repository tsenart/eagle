package main

import "github.com/BurntSushi/toml"

// Config provides a simple config file format to configure eagle.
type Config struct {
	Name  string
	Tests map[string]struct {
		URL     string
		Address string
	} `toml:"test"`
}

func loadConfig(path string) (Config, error) {
	var c Config
	if _, err := toml.DecodeFile(path, &c); err != nil {
		return c, err
	}

	return c, nil
}
