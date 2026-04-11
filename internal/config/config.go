package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Redis    RedisConfig
	Firebase FirebaseConfig
}

type ServerConfig struct {
	Port           string
	AllowedOrigins string
}

type RedisConfig struct {
	Addr string
}

type FirebaseConfig struct {
	CredentialsPath string
}

// Load reads configuration from the .env file and environment variables.
// Environment variables take precedence over the .env file.
func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Warning: failed to read .env file, relying on system envs")
	}

	return &Config{
		Server: ServerConfig{
			Port:           getEnv("PORT", "8080"),
			AllowedOrigins: getEnv("ALLOWED_ORIGINS", ""),
		},
		Redis: RedisConfig{
			Addr: getEnv("REDIS_ADDR", "localhost:6379"),
		},
		Firebase: FirebaseConfig{
			// Leave empty on Cloud Run to use Application Default Credentials.
			CredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if viper.IsSet(key) {
		return viper.GetString(key)
	}
	return defaultValue
}
