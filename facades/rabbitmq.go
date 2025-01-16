package facades

import (
	"log"

	"github.com/orangbus/rabbitmq"
	"github.com/orangbus/rabbitmq/contracts"
)

func Rabbitmq() contracts.Rabbitmq {
	instance, err := rabbitmq.App.Make(rabbitmq.Binding)
	if err != nil {
		log.Println("mq初始化错误")
		log.Println(err)
		return nil
	}

	return instance.(contracts.Rabbitmq)
}
