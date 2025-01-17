package rabbitmq

import (
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Dlx struct {
	client       *Rabbitmq
	queueName    string
	exchangeName string
	key          string
	routingKey   string

	deadExchangeName string
	deadKey          string
	deadQueue        string
	deadRoutingKey   string
}

func NewDlx(r *Rabbitmq) *Dlx {
	return &Dlx{client: r, deadExchangeName: "dlx.exchange", deadKey: "dlx.key"}
}

func (d *Dlx) SetOption(exchangeName, queueName, key, routingKey string) *Dlx {
	d.queueName = exchangeName
	d.exchangeName = queueName
	d.key = key
	d.routingKey = routingKey
	return d
}
func (d *Dlx) SetDeadOption(exchangeName, queueName, key, routingKey string) *Dlx {
	d.deadQueue = exchangeName
	d.deadExchangeName = queueName
	d.deadKey = key
	d.deadRoutingKey = routingKey
	return d
}

func (d *Dlx) PushMsg(exchangeName string, msg any, timeout int) error {
	d.exchangeName = exchangeName
	d.queueName = "dlx.queue"

	// 1、声明队列
	_, err := d.client.channel.QueueDeclare(d.queueName, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "dlx.exchange", // 指定死信交换机
		"x-dead-letter-routing-key": "dlx.key",      // 指定死信routing-key
		"x-message-ttl":             timeout,        // 消息过期时间,毫秒
	})
	if err != nil {
		return err
	}

	// 2、声明交换机
	err = d.client.channel.ExchangeDeclare(d.exchangeName, amqp.ExchangeDirect, true, false, false, false, nil)
	if err != nil {
		return err
	}

	// 3、绑定交换机
	err = d.client.channel.QueueBind(d.queueName, d.key, d.exchangeName, false, nil)
	if err != nil {
		return err
	}

	// 4、声明死信队列
	_, err = d.client.channel.QueueDeclare(d.deadQueue, true, false, false, false, nil)
	// 5、声明死信交换机
	if err = d.client.channel.ExchangeDeclare(d.deadExchangeName, amqp.ExchangeDirect, true, false, false, false, nil); err != nil {
		return err
	}
	// 6、绑定死信交换机
	if err = d.client.channel.QueueBind(d.deadQueue, d.deadRoutingKey, d.deadExchangeName, false, nil); err != nil {
		return err
	}
	// 发送消息
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return d.client.channel.Publish(d.exchangeName, d.routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         data,
	})
}

func (d *Dlx) ConsumeMsg(queueName string) (<-chan amqp.Delivery, error) {
	d.queueName = queueName
	return d.client.channel.Consume(d.queueName, "", false, false, false, false, nil)
}

func (d *Dlx) ConsumeDlxMsg(deadQueueName string) (<-chan amqp.Delivery, error) {
	d.deadQueue = deadQueueName
	return d.client.channel.Consume(d.deadQueue, "", false, false, false, false, nil)
}
