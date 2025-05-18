package config

import (
	"errors"
	"fmt"
	"github.com/adrg/xdg"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

const (
	appName = "ApiQube"
)

type CLIConfig struct {
	UI struct {
		TimestampColor string `mapstructure:"timestamp_color" yaml:"timestamp_color" json:"timestampColor"`
		LogColor       string `mapstructure:"log_color" yaml:"log_color" json:"logColor"`
		SuccessColor   string `mapstructure:"success_color" yaml:"success_color" json:"successColor"`
		WarnColor      string `mapstructure:"warn_color" yaml:"warn_color" json:"warnColor"`
		ErrorColor     string `mapstructure:"error_color" yaml:"error_color" json:"errorColor"`
		DebugColor     string `mapstructure:"debug_color" yaml:"debug_color" json:"debugColor"`
		InfoColor      string `mapstructure:"info_color" yaml:"info_color" json:"infoColor"`
	} `mapstructure:"ui" yaml:"ui" json:"ui"`
}

func InitConfig() (*CLIConfig, error) {
	configPath := filepath.Join(xdg.ConfigHome, appName)

	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.SetEnvPrefix("QUBE")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			if err = createDefaultConfig(configPath); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
		}
	}

	var cfg CLIConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal error: %w", err)
	}

	return &cfg, nil
}

func createDefaultConfig(configPath string) error {
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	cfgFile := filepath.Join(configPath, "config.yaml")

	defaultCfg := CLIConfig{}
	_ = defaultCfg

	file, err := os.Create(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if err = viper.WriteConfigAs(cfgFile); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	fmt.Printf("Created default config at: %s\n", cfgFile)
	return nil
}
