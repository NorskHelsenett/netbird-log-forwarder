package settings

import (
	"fmt"

	"github.com/spf13/viper"
)

func InitConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}
	return nil
}
