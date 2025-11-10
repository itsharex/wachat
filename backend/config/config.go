package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration
type Config struct {
	AI       *AIConfig       `mapstructure:"ai"`
	Binaries *BinariesConfig `mapstructure:"binaries"`
}

// AIConfig holds AI service configuration
type AIConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
}

// BinariesConfig holds binary manager configuration
type BinariesConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	UseEmbedded  bool     `mapstructure:"use_embedded"` // true: use embedded, false: use local bin/ directory
	BinPath      string   `mapstructure:"bin_path"`     // local bin directory path (default: "./bin")
	StartupOrder []string `mapstructure:"startup_order"`
}

// IsEnabled returns whether binary manager is enabled
func (c *BinariesConfig) IsEnabled() bool {
	return c != nil && c.Enabled
}

// IsUseEmbedded returns whether to use embedded mode
func (c *BinariesConfig) IsUseEmbedded() bool {
	return c != nil && c.UseEmbedded
}

// GetBinPath returns the bin directory path
func (c *BinariesConfig) GetBinPath() string {
	if c == nil {
		return ""
	}
	return c.BinPath
}

// GetStartupOrder returns the startup order list
func (c *BinariesConfig) GetStartupOrder() []string {
	if c == nil {
		return nil
	}
	return c.StartupOrder
}

var globalConfig *Config

// findProjectRoot tries to find project root by looking for go.mod
func findProjectRoot(startDir string) (string, error) {
	dir := startDir
	for i := 0; i < 2; i++ { // Limit search depth
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break // Reached filesystem root
		}
		dir = parentDir
	}
	return "", fmt.Errorf("go.mod not found")
}

// Load loads configuration from yaml file
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Priority 1: Environment variable WACHAT_CONFIG_PATH
	if configPath := os.Getenv("WACHAT_CONFIG_PATH"); configPath != "" {
		log.Printf("Using config path from environment: %s", configPath)
		viper.AddConfigPath(configPath)
	}

	// Priority 2: Current working directory
	if cwd, err := os.Getwd(); err == nil {
		log.Printf("Current working directory: %s", cwd)
		viper.AddConfigPath(cwd)

		// Try to find project root from cwd
		if projectRoot, err := findProjectRoot(cwd); err == nil {
			log.Printf("Found project root from cwd: %s", projectRoot)
			viper.AddConfigPath(projectRoot)
		}
	}

	// Priority 3: Executable directory
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)

		// Resolve symlinks (important for development)
		if realPath, err := filepath.EvalSymlinks(execPath); err == nil {
			execDir = filepath.Dir(realPath)
		}

		log.Printf("Executable directory: %s", execDir)
		viper.AddConfigPath(execDir)

		// Try to find project root from executable directory
		if projectRoot, err := findProjectRoot(execDir); err == nil {
			log.Printf("Found project root from executable: %s", projectRoot)
			viper.AddConfigPath(projectRoot)
		}
	}

	// Priority 4: User home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		configHome := filepath.Join(homeDir, ".config", "wachat")
		viper.AddConfigPath(configHome)
	}

	// Fallback: current directory
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("ai.base_url", "https://api.openai.com/v1")
	viper.SetDefault("ai.model", "gpt-3.5-turbo")
	viper.SetDefault("binaries.enabled", false)
	viper.SetDefault("binaries.use_embedded", false) // Default to local mode
	viper.SetDefault("binaries.bin_path", "./bin")
	viper.SetDefault("binaries.startup_order", []string{})

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			log.Println("Warning: config.yaml not found, using defaults")
		}
	} else {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	// Unmarshal to struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get returns the global config instance
func Get() *Config {
	if globalConfig == nil {
		log.Fatal("Config not loaded. Call config.Load() first")
	}
	return globalConfig
}

// GetAIConfig returns AI configuration
func GetAIConfig() *AIConfig {
	cfg := Get()
	return cfg.AI
}
