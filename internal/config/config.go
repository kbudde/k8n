package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Watcher []Watcher `yaml:"watcher"`
}

type Watcher struct {
	Name       string `yaml:"name"`
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Selector   string `yaml:"selector"`
	Namespace  string `yaml:"namespace"`
}

func FromYaml(yml []byte) (*Config, error) {
	var cfg Config

	if err := yaml.Unmarshal(yml, &cfg); err != nil {
		return nil, err
	}

	if cfg.APIVersion != "k8n.budd.ee/v1beta" {
		return nil, fmt.Errorf("invalid apiVersion %s", cfg.APIVersion)
	}

	if cfg.Kind != "config" {
		return nil, fmt.Errorf("invalid kind %s", cfg.Kind)
	}

	return &cfg, nil
}

func FromYamlFile(filename string) (*Config, error) {
	yml, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return FromYaml(yml)
}
