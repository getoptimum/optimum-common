package config_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/getoptimum/optimum-common/pkg/config"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Host  string `yaml:"host" env:"HOST" flag:"host"`
	Port  int    `yaml:"port" env:"PORT" flag:"port"`
	Debug bool   `yaml:"debug" env:"DEBUG" flag:"debug"`
}

func TestLoadPriority(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "cfg.yaml")
	yamlData := []byte("host: yamlhost\nport: 8080\ndebug: false\n")
	require.NoError(t, os.WriteFile(p, yamlData, 0o600))

	t.Setenv("HOST", "envhost")
	t.Setenv("PORT", "9090")

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("host", "", "")
	fs.Int("port", 0, "")
	fs.Bool("debug", false, "")
	os.Args = []string{"cmd", "-host", "flaghost", "-debug=true"}

	cfg := testConfig{}
	require.NoError(t, config.Load(&cfg, config.WithYAML(p), config.WithFlagSet(fs)))

	require.Equal(t, "flaghost", cfg.Host)
	require.Equal(t, 9090, cfg.Port)
	require.True(t, cfg.Debug)
}

func TestEnvPrefix(t *testing.T) {
	t.Setenv("APP_HOST", "prefixedhost")
	t.Setenv("APP_PORT", "4242")
	t.Setenv("APP_DEBUG", "true")

	cfg := testConfig{}
	require.NoError(t, config.Load(&cfg, config.WithEnvPrefix("APP")))

	require.Equal(t, "prefixedhost", cfg.Host)
	require.Equal(t, 4242, cfg.Port)
	require.True(t, cfg.Debug)
}

func TestNestedStruct(t *testing.T) {
	type NestedConfig struct {
		LogLevel string `env:"LOG_LEVEL" flag:"log-level"`
	}
	type ComplexConfig struct {
		AppName string `env:"APP_NAME" flag:"app-name"`
		Nested  NestedConfig
	}

	t.Setenv("APP_NAME", "myapp")
	t.Setenv("LOG_LEVEL", "debug")

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("app-name", "", "")
	fs.String("log-level", "", "")
	os.Args = []string{"cmd", "-log-level=info"}

	cfg := ComplexConfig{}
	require.NoError(t, config.Load(&cfg, config.WithFlagSet(fs)))

	require.Equal(t, "myapp", cfg.AppName)
	require.Equal(t, "info", cfg.Nested.LogLevel) // Flag overrides env
}

func TestInvalidTypeConversion(t *testing.T) {
	t.Setenv("PORT", "notanumber")

	cfg := testConfig{}
	err := config.Load(&cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid syntax")
}

func TestUnsupportedType(t *testing.T) {
	type UnsupportedConfig struct {
		Data []string `env:"DATA"` // slices not supported
	}

	t.Setenv("DATA", "foo,bar")

	cfg := UnsupportedConfig{}
	err := config.Load(&cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported kind")
}
