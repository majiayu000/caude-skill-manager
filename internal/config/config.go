package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the global configuration
type Config struct {
	SkillsDir string `json:"skills_dir"`
	Registry  string `json:"registry"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		SkillsDir: filepath.Join(homeDir, ".claude", "skills"),
		Registry:  "github",
	}
}

// GetSkillsDir returns the skills directory path
func GetSkillsDir() string {
	cfg := Load()
	return cfg.SkillsDir
}

// ConfigPath returns the path to config file
func ConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".skrc")
}

// Load loads configuration from file
func Load() *Config {
	cfg := DefaultConfig()

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return cfg
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		fmt.Fprintln(os.Stderr, "Warning: failed to parse config file, using defaults:", err)
		return cfg
	}
	return cfg
}

// Save saves configuration to file
func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0644)
}

// EnsureSkillsDir creates the skills directory if it doesn't exist
func EnsureSkillsDir() error {
	dir := GetSkillsDir()
	return os.MkdirAll(dir, 0755)
}
