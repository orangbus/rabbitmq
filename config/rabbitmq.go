package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("rabbitmq", map[string]any{
		"host":          config.Env("RABBITMQ_HOST", "localhost"),
		"port":          config.Env("RABBITMQ_PORT", 5672),
		"username":      config.Env("RABBITMQ_USERNAME", "guest"),
		"password":      config.Env("RABBITMQ_PASSWORD", "guest"),
		"vhost":         config.Env("RABBITMQ_VHOST", "/"),
		"queue":         config.Env("RABBITMQ_QUEUE", "dev"),
		"prefetchCount": config.Env("RABBITMQ_PREFETCH_COUNT", 10),
		"dead": map[string]any{
			"exchange":    config.Env("RABBITMQ_DEAD_EXCHANGE", "dead.exchange"),
			"queue":       config.Env("RABBITMQ_DEAD_QUEUE", "dead.queue"),
			"routing_key": config.Env("RABBITMQ_DEAD_ROUTING_KEY", "dead.routing_key"),
		},
	})
}
