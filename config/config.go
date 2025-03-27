package config
import (
	"github.com/spf13/viper"
)
type Config struct {
	WeatherAPIKey string `mapstructure:"WEATHER_API_KEY"`
	WebServerPort string `mapstructure:"WEB_SERVER_PORT"`
}
func LoadConfig(path string) (*Config, error) {
	var cfg Config
	viper.SetConfigName(".env") 
	viper.SetConfigType("env")  
	viper.AddConfigPath(path)   
	viper.AutomaticEnv()        
	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok && err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
