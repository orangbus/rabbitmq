package contracts

type Rabbitmq interface {
	Msg(msg any) error
	Publish(exchangeName string, data interface{}) error // 工作队列
	Routing(exchangeName, key string, data interface{}) error
	Topic(exchangeName, key string, data interface{}) error

	Consume()
	ConsumePublish(exchangeName string) error
	ConsumeRouting(exchangeName, key string) error
	ConsumeTopic(exchangeName, key string) error
}
