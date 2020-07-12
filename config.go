package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// config represents the configuration
type config struct {
	Opencast struct {
		URL      string `yaml:"url"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"opencast"`

	Server struct {
		Address string `yaml:"address"`
	}
}

// newConfig reads in a file and create a new configuration with its content.
func newConfig(filename string) (*config, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read %v, %w", filename, err)
	}

	c := &config{}
	err = yaml.Unmarshal(bs, c)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal config file, %w", err)
	}

	return c, err
}
