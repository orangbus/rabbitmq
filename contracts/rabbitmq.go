package contracts

import (
	"github.com/orangbus/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

type Rabbitmq interface {
	Msg(msg any) error
	Publish(exchangeName string, data interface{}) error // 工作队列
	Routing(exchangeName, key string, data interface{}) error
	Topic(exchangeName, key string, data interface{}) error

	ConsumeMsg() (<-chan amqp091.Delivery, error)
	ConsumePublish(exchangeName string) (<-chan amqp091.Delivery, error)
	ConsumeRouting(exchangeName, key string) (<-chan amqp091.Delivery, error)
	ConsumeTopic(exchangeName, key string) (<-chan amqp091.Delivery, error)

	Dlx() *rabbitmq.Dlx
}
