package facades

import (
	"github.com/orangbus/rabbitmq"
	"github.com/orangbus/rabbitmq/contracts"
	"log"
)

func Rabbitmq() contracts.Rabbitmq {
	instance, err := rabbitmq.App.Make(rabbitmq.Binding)
	if err != nil {
		log.Printf("rabbitmq make error: %v", err)
		return nil
	}

	return instance.(contracts.Rabbitmq)
}
