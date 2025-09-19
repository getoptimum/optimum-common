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

	require.NoError(t, os.Setenv("HOST", "envhost"))
	require.NoError(t, os.Setenv("PORT", "9090"))
	// ensure env cleanup
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("HOST"))
		require.NoError(t, os.Unsetenv("PORT"))
	})

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
	require.NoError(t, os.Setenv("APP_HOST", "prefixedhost"))
	require.NoError(t, os.Setenv("APP_PORT", "4242"))
	require.NoError(t, os.Setenv("APP_DEBUG", "true"))
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("APP_HOST"))
		require.NoError(t, os.Unsetenv("APP_PORT"))
		require.NoError(t, os.Unsetenv("APP_DEBUG"))
	})

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

	require.NoError(t, os.Setenv("APP_NAME", "myapp"))
	require.NoError(t, os.Setenv("LOG_LEVEL", "debug"))
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("APP_NAME"))
		require.NoError(t, os.Unsetenv("LOG_LEVEL"))
	})

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
	require.NoError(t, os.Setenv("PORT", "notanumber"))
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("PORT"))
	})

	cfg := testConfig{}
	err := config.Load(&cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid syntax")
}

func TestUnsupportedType(t *testing.T) {
	type UnsupportedConfig struct {
		Data []string `env:"DATA"` // slices not supported
	}

	require.NoError(t, os.Setenv("DATA", "foo,bar"))
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("DATA"))
	})

	cfg := UnsupportedConfig{}
	err := config.Load(&cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported kind")
}
