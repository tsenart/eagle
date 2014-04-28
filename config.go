package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config provides a simple config file format to configure eagle.
type Config struct {
	Name  string
	Path  string
	Rate  int
	Tests map[string]configLayer `toml:"test"`
}

type configLayer struct {
	URL     string
	Address string
}

// NewConfig parses the file at the given path and returns a Config object.
func NewConfig(path string) (Config, error) {
	var c Config
	if _, err := toml.DecodeFile(path, &c); err != nil {
		return c, err
	}

	return c, nil
}

// Endpoints returns a list of all service endpoints. In case both URL and
// Address are configured, only URL is used.
func (t configLayer) Endpoints() ([]string, error) {
	if t.URL != "" {
		return []string{t.URL}, nil
	}

	if t.Address != "" {
		_, addrs, err := net.LookupSRV("", "", t.Address)
		if err != nil {
			return []string{}, err
		}

		e := make([]string, 0, len(addrs))
		for _, addr := range addrs {
			host := strings.Trim(addr.Target, ".")
			e = append(e, fmt.Sprintf("http://%s:%d/", host, addr.Port))
		}

		return e, nil
	}

	return []string{}, nil
}
