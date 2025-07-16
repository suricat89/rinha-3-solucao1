package conf

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type config struct {
	ProcessorDefaultBaseUrl  string `env:"PROCESSOR_DEFAULT_BASEURL,required"`
	ProcessorFallbackBaseUrl string `env:"PROCESSOR_FALLBACK_BASEURL,required"`
	QueueBufferSize          int    `env:"QUEUE_BUFFER_SIZE" envDefault:"10000"`
	Port                     int    `env:"PORT" envDefault:"9999"`
	ConsumerGoroutines       int    `env:"CONSUMER_GOROUTINES" envDefault:"20"`
	RedisHost                string `env:"REDIS_HOST,required"`
	RedisPort                int    `env:"REDIS_PORT" envDefault:"6379"`
}

var Env config

func init() {
	err := env.Parse(&Env)
	if err != nil {
		log.Fatalf("Error parsing envs. %v", err)
	}
}
