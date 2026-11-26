package test

import (
	"os"
	"slices"

	"gopkg.in/yaml.v3"
)

const suiteConfigFile = "./testconfig/suites.yaml"

// SuiteConfig controls which test suites run in this pipeline execution.
// CI writes this file before tests run.
type SuiteConfig struct {
	Suites []string `yaml:"suites"`
}

var allSuites = []string{"zone", "postgresql", "cronjob", "repository", "s3bucket", "valkey", "mariadb", "rabbitmq", "webapp", "webaccess"}

func loadSuiteConfig() SuiteConfig {
	data, err := os.ReadFile(suiteConfigFile)
	if err != nil {
		return SuiteConfig{Suites: allSuites}
	}
	var cfg SuiteConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return SuiteConfig{Suites: allSuites}
	}
	return cfg
}

func (c SuiteConfig) Has(suite string) bool {
	return slices.Contains(c.Suites, suite)
}

func (c SuiteConfig) Only(suites []string) SuiteConfig {
	if suites == nil {
		return c
	}
	var kept []string
	for _, s := range suites {
		if c.Has(s) {
			kept = append(kept, s)
		}
	}
	return SuiteConfig{Suites: kept}
}
