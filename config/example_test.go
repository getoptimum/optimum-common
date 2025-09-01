package config_test

import (
	"fmt"
	"os"

	oc "github.com/getoptimum/optimum-common/config"
)

// Example demonstrating how a CLI application can load configuration
// values from a YAML file, environment variables, and command-line flags
// using the shared loader. The override priority is flags > env vars > YAML.
func ExampleLoad() {
	// AppConfig represents application settings.
	type AppConfig struct {
		Addr  string `yaml:"addr"`
		Debug bool   `yaml:"debug"`
	}

	cfg := AppConfig{
		Addr: ":8080", // default value
	}
	defs := []oc.FlagDef{
		{Name: "addr", Usage: "listen address", Value: &cfg.Addr},
		{Name: "debug", Usage: "enable debug mode", Value: &cfg.Debug},
	}

	// Simulate environment variable and flag input.
	if err := os.Setenv("OPTIMUM_DEBUG", "true"); err != nil {
		panic(err)
	}
	os.Args = []string{"app", "-addr", ":9090"}

	if err := oc.Load(&cfg, defs); err != nil {
		panic(err)
	}

	fmt.Println(cfg.Addr, cfg.Debug)
	// Output:
	// :9090 true
}
