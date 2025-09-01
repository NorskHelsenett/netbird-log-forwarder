package settings

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

func InitConfig(config_path string) (string, error) {
	abs, err := filepath.Abs(config_path)
	if err != nil {
		return "", fmt.Errorf("resolve config path: %w", err)
	}

	viper.SetConfigFile(abs)
	if err := viper.ReadInConfig(); err != nil {
		return "", fmt.Errorf("failed to read config %s: %w", abs, err)
	}

	return viper.ConfigFileUsed(), nil
}

func InitSecrets(secrets_path string) (string, error) {
	abs, err := filepath.Abs(secrets_path)
	if err != nil {
		return "", fmt.Errorf("resolve secrets path: %w", err)
	}

	viper.SetConfigFile(abs)
	if err := viper.MergeInConfig(); err != nil {
		return "", fmt.Errorf("failed to read secrets %s: %w", abs, err)
	}

	return viper.ConfigFileUsed(), nil
}
