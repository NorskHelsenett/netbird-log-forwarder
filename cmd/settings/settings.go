package settings

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

func InitConfig(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve config path: %w", err)
	}

	viper.SetConfigFile(abs)
	if err := viper.ReadInConfig(); err != nil {
		return "", fmt.Errorf("failed to read config %s: %w", abs, err)
	}

	return viper.ConfigFileUsed(), nil
}
