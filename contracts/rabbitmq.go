package contracts

import (
	"github.com/orangbus/rabbitmq"
)

type Rabbitmq interface {
	Msg(msg any) error
	Publish(exchangeName string, data interface{}) error // 工作队列
	Routing(exchangeName, key string, data interface{}) error
	Topic(exchangeName, key string, data interface{}) error

	ConsumeMsg() (<-chan []byte, error)
	ConsumePublish(exchangeName string, queueName ...string) (<-chan []byte, error)
	ConsumeRouting(exchangeName, key string, queueName ...string) (<-chan []byte, error)
	ConsumeTopic(exchangeName, key string, queueName ...string) (<-chan []byte, error)

	Dlx() *rabbitmq.Dlx
}
