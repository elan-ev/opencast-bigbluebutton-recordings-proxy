package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

// config represents the configuration
type config struct {
	Opencast struct {
		URL             string        `yaml:"url"`
		Username        string        `yaml:"username"`
		Password        string        `yaml:"password"`
		CacheExpiration time.Duration `yaml:"cache_expiration"`
		RequestTimeout  time.Duration `yaml:"request_timeout"`
	} `yaml:"opencast"`

	BigBlueButton struct {
		Secret string `yaml:"secret"`
	} `yaml:"bigbluebutton"`

	Server struct {
		Address string `yaml:"address"`
	} `yaml:"server"`
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
