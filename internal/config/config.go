package config

import (
	"log"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `env:"ENV" env-default:"local"`
	DatabaseURL string `env:"DATABASE_URL" env-required:"true"`
	RedisHost   string `env:"REDIS_HOST" env-default:"localhost:6379"`
	IsDebug bool `env:"IS_DEBUG" env-default:"false"`
	HTTP	struct {
			Port string `env:"HTTP_PORT" env-default:":8080"`
			Timeout time.Duration `env:"HTTP_TIMEOUT" env-default:"5s"`
	}
	Postgres struct {
		Host		string `env:"DB_HOST" env-required:"true"`
		Port		string `env:"DB_PORT" env-default:"5432"`
		User 		string `env:"DB_USER" env-required:"true"`
		Password 	string `env:"DB_PASSWORD" env-required:"true"`
		Name 		string `env:"DB_NAME" env-required:"true"`
	}
}

var (
	instance *Config
	once     sync.Once
)

// GetConfig читает .env и заполняет структуру один раз (Singleton)
func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		
		// Сначала пробуем прочитать .env (для локальной разработки)
		// Если файла нет, cleanenv вернет ошибку, которую мы просто проигнорируем
		_ = cleanenv.ReadConfig(".env", instance)
		
		// Затем читаем переменные окружения (они перекроют дефолты или данные из .env)
		// Если даже обязательные переменные (env-required) не найдены — тогда падаем
		if err := cleanenv.ReadEnv(instance); err != nil {
			log.Fatalf("Config error: %v", err)
		}
	})
	return instance
}