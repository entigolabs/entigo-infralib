package test

import (
	"os"

	"gopkg.in/yaml.v3"
)

const suiteConfigFile = "./testconfig/suites.yaml"

// SuiteConfig controls which test suites run in this pipeline execution.
// CI writes this file before tests run.
type SuiteConfig struct {
	Suites []string `yaml:"suites"`
}

// TODO: add kafka to array when ready for testing
var allSuites = []string{"zone", "postgresql", "cronjob", "repository", "s3bucket", "valkey", "webapp", "webaccess"}

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
	for _, s := range c.Suites {
		if s == suite {
			return true
		}
	}
	return false
}
