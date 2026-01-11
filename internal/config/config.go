package config

import (
	"log"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Port                  string `env:"PORT" envDefault:"8080"`
	DefaultPriority       int64  `env:"DEFAULT_PRIORITY" envDefault:"0"`
	DefaultTTL            int64  `env:"DEFAULT_TTL" envDefault:"900"`
	DefaultNumberWorkers  int    `env:"DEFAULT_NUMBER_WORKER" envDefault:"4"`
	DefaultChanelCapacity int    `env:"DEFAULT_CHANEL_CAPACITY" envDefault:"10000"`
	REDIS_URL             string `env:"REDIS_URL" envDefault:"localhost:6379"`
}

func LoadConfig() *Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Unable to parse env: %v", err)
	}
	return &cfg
}
