package config

import "github.com/go-yaml/yaml"

type Config struct {
	Name string `yaml:"name"`
}

func NewFromString(s string) (*Config, error) {
	c := new(Config)

	err := yaml.Unmarshal([]byte(s), &c)

	if err != nil {
		return nil, err
	}

	return c, nil
}
