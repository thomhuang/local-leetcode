package assistant

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultModel is the current DeepSeek model (DeepSeek-V4.1-Flash). The
	// legacy deepseek-chat/deepseek-reasoner names were retired in July 2026.
	DefaultModel           = "deepseek-flash"
	DefaultBaseURL         = "https://api.deepseek.com"
	DefaultReasoningEffort = "medium"
	DefaultMaxTokens       = 8192
	DefaultTemperature     = 1.0

	configFileName = "config.json"
	promptFileName = "system_prompt.md"
)

// Config holds the assistant settings that live in config.json.
type Config struct {
	Model           string  `json:"model"`
	BaseURL         string  `json:"base_url"`
	ReasoningEffort string  `json:"reasoning_effort"`
	Thinking        bool    `json:"thinking"`
	MaxTokens       int     `json:"max_tokens"`
	Temperature     float64 `json:"temperature"`
}

// DefaultConfig returns the settings used when config.json is missing or
// incomplete.
func DefaultConfig() Config {
	return Config{
		Model:           DefaultModel,
		BaseURL:         DefaultBaseURL,
		ReasoningEffort: DefaultReasoningEffort,
		Thinking:        true,
		MaxTokens:       DefaultMaxTokens,
		Temperature:     DefaultTemperature,
	}
}

// Load reads the assistant settings and system prompt from dir. A missing
// config.json is created with defaults, and missing individual config fields
// keep their default value. The system prompt is read as-is and must exist.
func Load(dir string) (Config, string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Config{}, "", fmt.Errorf("create assistant directory: %w", err)
	}

	cfg, err := loadConfig(dir)
	if err != nil {
		return Config{}, "", err
	}

	prompt, err := loadSystemPrompt(dir)
	if err != nil {
		return Config{}, "", err
	}
	return cfg, prompt, nil
}

func loadConfig(dir string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(filepath.Join(dir, configFileName))
	if errors.Is(err, fs.ErrNotExist) {
		if err := SaveConfig(dir, cfg); err != nil {
			return cfg, err
		}
		return cfg, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read assistant config: %w", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse assistant config: %w", err)
	}
	return cfg, nil
}

// SaveConfig writes the settings to config.json in dir.
func SaveConfig(dir string, cfg Config) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create assistant directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode assistant config: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, configFileName), append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write assistant config: %w", err)
	}
	return nil
}

func loadSystemPrompt(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, promptFileName))
	if err != nil {
		return "", fmt.Errorf("read system prompt: %w", err)
	}
	return string(data), nil
}
