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
