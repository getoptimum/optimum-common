package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Name    string   `yaml:"name"`
	Count   int      `yaml:"count"`
	Enabled bool     `yaml:"enabled"`
	Items   []string `yaml:"items"`
	Ratio   float32  `yaml:"ratio"` // for testing floats
}

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	file := filepath.Join(dir, "cfg.yaml")
	require.NoError(t, os.WriteFile(file, []byte(content), 0o600))
	return file
}

func TestLoad_Unmarshal(t *testing.T) {
	yaml := "name: yaml\ncount: 1\nenabled: false\nitems: [a,b]\nratio: 0.5"
	path := writeTempYAML(t, yaml)

	cfg := testConfig{}
	defs := []FlagDef{
		{Name: "name", Value: &cfg.Name},
		{Name: "count", Value: &cfg.Count},
		{Name: "enabled", Value: &cfg.Enabled},
		{Name: "items", Value: &cfg.Items},
		{Name: "ratio", Value: &cfg.Ratio},
	}
	os.Args = []string{"cmd", "-config", path}
	require.NoError(t, Load(&cfg, defs))
	require.Equal(t, "yaml", cfg.Name)
	require.Equal(t, 1, cfg.Count)
	require.False(t, cfg.Enabled)
	require.Equal(t, []string{"a", "b"}, cfg.Items)
	require.InDelta(t, 0.5, cfg.Ratio, 0.0001)
}

func TestLoad_OverridePriority(t *testing.T) {
	yaml := "name: yaml\ncount: 1\nenabled: false"
	path := writeTempYAML(t, yaml)
	cfg := testConfig{}
	defs := []FlagDef{
		{Name: "name", Value: &cfg.Name},
		{Name: "count", Value: &cfg.Count},
		{Name: "enabled", Value: &cfg.Enabled},
	}
	t.Setenv("OPTIMUM_NAME", "env")
	t.Setenv("OPTIMUM_COUNT", "2")
	os.Args = []string{"cmd", "-config", path, "-count", "3", "-enabled", "true"}
	require.NoError(t, Load(&cfg, defs))
	require.Equal(t, "env", cfg.Name) // env overrides YAML
	require.Equal(t, 3, cfg.Count)    // flag overrides env
	require.True(t, cfg.Enabled)      // flag overrides YAML
}
