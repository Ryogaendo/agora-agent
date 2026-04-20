package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	AgentID       string             `json:"agent_id"`
	EnvironmentID string             `json:"environment_id"`
	GitHubToken   string             `json:"github_token,omitempty"`
	AgoraPath     string             `json:"agora_path"`
	SystemPrompt  string             `json:"system_prompt,omitempty"` // optional custom system prompt for the agent
	Repos         map[string]string  `json:"repos"`                   // name -> github URL
	Skills        map[string]string  `json:"skills"`                  // name -> skill_id
	Routines      map[string]Routine `json:"routines"`                // name -> routine config
}

// Routine holds the trigger ID and token for a Claude Code routine.
type Routine struct {
	TriggerID   string `json:"trigger_id"`
	Token       string `json:"token"`
	Description string `json:"description,omitempty"`
}

func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".agora-agent.json")
}

func Load() (*Config, error) {
	path := DefaultConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(DefaultConfigPath(), data, 0600)
}

func defaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		AgoraPath: filepath.Join(home, "projects", "agora"),
		Repos:     map[string]string{},
		Skills:    map[string]string{},
	}
}
