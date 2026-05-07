package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPathUsesConfiguredEnvValue(t *testing.T) {
	got := ExpandPath("$X/any", []string{"X=/Users/example/Projects/xhd2015"})
	want := "/Users/example/Projects/xhd2015/any"
	if got != want {
		t.Fatalf("ExpandPath() = %q, want %q", got, want)
	}
}

func TestCollapsePathUsesConfiguredEnvName(t *testing.T) {
	got := CollapsePath("/Users/example/Projects/xhd2015/any", []string{"X=/Users/example/Projects/xhd2015"})
	want := "$X/any"
	if got != want {
		t.Fatalf("CollapsePath() = %q, want %q", got, want)
	}
}

func TestLoadDefaultReadsLLSConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configFile, err := DefaultFile(true)
	if err != nil {
		t.Fatalf("DefaultFile: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(configFile, []byte(`{"envs":["X=/tmp/x"],"projects":["$X/any"]}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, gotFile, err := LoadDefault(false)
	if err != nil {
		t.Fatalf("LoadDefault: %v", err)
	}
	if gotFile != configFile {
		t.Fatalf("LoadDefault file = %q, want %q", gotFile, configFile)
	}
	if len(cfg.Envs) != 1 || cfg.Envs[0] != "X=/tmp/x" {
		t.Fatalf("unexpected envs: %#v", cfg.Envs)
	}
}
