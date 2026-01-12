package config

import (
	"log"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Port                  string   `env:"PORT" envDefault:"8080"`
	DefaultPriority       int64    `env:"DEFAULT_PRIORITY" envDefault:"0"`
	DefaultTTL            int64    `env:"DEFAULT_TTL" envDefault:"900"`
	DefaultNumberWorkers  int      `env:"DEFAULT_NUMBER_WORKER" envDefault:"4"`
	DefaultChanelCapacity int      `env:"DEFAULT_CHANEL_CAPACITY" envDefault:"10000"`
	RedisUrl              string   `env:"REDIS_URL" envDefault:"localhost:6379"`
	KafkaBrokers          []string `env:"KAFKA_BROKERS" envSeparator:"," envDefault:"localhost:9092"`
	KafkaTopic            string   `env:"KAFKA_TOPIC" envDefault:"notification_topic"`
	KafkaConsumerGroup    string   `env:"KAFKA_CONSUMER_GROUP" envDefault:"notification_dispatcher_group"`
}

func LoadConfig() *Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Unable to parse env: %v", err)
	}
	return &cfg
}
