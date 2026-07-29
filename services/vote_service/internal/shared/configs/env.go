package configs

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Configs struct{
    Api ApiConfig
	Db DatabaseConfig
}

type ApiConfig struct {
   Addr string	
}

type DatabaseConfig struct {
	Url string
}


const ENV_PREFIX = "SERVER_"

func InitConfigs() Configs {
	// Fail silently for production
	_ = godotenv.Load()
	
	return Configs{
		Api: ApiConfig{
			Addr: getEnv(ENV_PREFIX+"ADDR", ":8080"),
		},
		Db: DatabaseConfig{
			Url: getEnv(ENV_PREFIX+"DB_URL", "postgresql://postgres:password@polls-maker-db:5432/polls_maker?sslmode=disable"),
		},
	}
}

func getEnv(key string, defaultValue string) string {

	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}

func getEnvFromInt(key string, defaultValue int) int {

	if value, ok := os.LookupEnv(key); ok {
		num, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}

		return num
	}

	return defaultValue
}