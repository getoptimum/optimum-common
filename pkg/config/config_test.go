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
	Host         string   `yaml:"host" env:"HOST" flag:"host" default:"localhost"`
	Port         int      `yaml:"port" env:"PORT" flag:"port" default:"8080"`
	Debug        bool     `yaml:"debug" env:"DEBUG" flag:"debug" default:"false"`
	Items        []string `yaml:"items" env:"ITEMS" flag:"items" default:"item1,item2"`
	DefaultParam string   `yaml:"default_param" env:"DEFAULT_PARAM" flag:"default_param" default:"defaultValue"`
}

func TestLoadPriority(t *testing.T) {
	// given
	dir := t.TempDir()
	p := filepath.Join(dir, "cfg.yaml")
	yamlData := `
host: yamlhost
port: 8080
debug: false
items:
  - a
  - b
`
	require.NoError(t, os.WriteFile(p, []byte(yamlData), 0o600))

	t.Run("yml should parse correctly", func(t *testing.T) {
		// when
		cfg := testConfig{}
		require.NoError(t, config.Load(&cfg, config.WithYAML(p)))

		// then
		require.Equal(t, "yamlhost", cfg.Host)
		require.Equal(t, 8080, cfg.Port)
		require.False(t, cfg.Debug)
		require.Equal(t, []string{"a", "b"}, cfg.Items)
		require.Equal(t, "defaultValue", cfg.DefaultParam)
	})

	t.Run("env should override yml", func(t *testing.T) {
		t.Setenv("HOST", "envhost")
		t.Setenv("PORT", "9090")
		t.Setenv("DEBUG", "true")
		t.Setenv("ITEMS", "x,y,z")

		// when
		cfg := testConfig{}
		require.NoError(t, config.Load(&cfg, config.WithYAML(p)))

		// then
		require.Equal(t, "envhost", cfg.Host)
		require.Equal(t, 9090, cfg.Port)
		require.True(t, cfg.Debug)
		require.Equal(t, []string{"x", "y", "z"}, cfg.Items)
		require.Equal(t, "defaultValue", cfg.DefaultParam)
	})

	t.Run("flags should override env and yml", func(t *testing.T) {
		t.Setenv("HOST", "envhost")
		t.Setenv("PORT", "9090")
		t.Setenv("DEBUG", "true")
		t.Setenv("ITEMS", "x,y,z")

		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("host", "", "")
		fs.Int("port", 0, "")
		fs.Bool("debug", false, "")
		fs.String("items", "", "")
		os.Args = []string{"cmd", "-host", "flaghost", "-debug=true", "-port=9010", "-items=m,n"}

		// when
		cfg := testConfig{}
		require.NoError(t, config.Load(&cfg, config.WithYAML(p), config.WithFlagSet(fs)))

		// then
		require.Equal(t, "flaghost", cfg.Host)
		require.Equal(t, 9010, cfg.Port)
		require.True(t, cfg.Debug)
		require.Equal(t, []string{"m", "n"}, cfg.Items)
	})
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
		Data map[string]string `env:"DATA"` // maps not supported
	}

	t.Setenv("DATA", "foo,bar")

	cfg := UnsupportedConfig{}
	err := config.Load(&cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported kind")
}
