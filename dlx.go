package rabbitmq

import (
	"encoding/json"
	"github.com/goravel/framework/facades"
	"github.com/mitchellh/mapstructure"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Dlx struct {
	client       *Rabbitmq
	queueName    string
	exchangeName string
	routingKey   string

	deadExchangeName string
	deadQueue        string
	deadRoutingKey   string
	timeout          int
}

func NewDlx(r *Rabbitmq) *Dlx {
	setting := facades.Config().Get("rabbitmq.dead")
	var dead map[string]string
	if err := mapstructure.Decode(setting, &dead); err != nil {
		panic(err)
	}
	d := Dlx{client: r}
	d.deadExchangeName = dead["exchange"]
	d.deadQueue = dead["queue"]
	d.deadRoutingKey = dead["routing_key"]
	return &d
}

func (d *Dlx) SetOption(exchangeName, queueName string) *Dlx {
	d.exchangeName = exchangeName
	d.queueName = queueName
	return d
}

func (d *Dlx) PushMsg(routingKey string, msg any) error {
	d.routingKey = routingKey

	// 1、声明正常队列
	_, err := d.client.channel.QueueDeclare(d.queueName, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    d.deadExchangeName, // 指定死信交换机
		"x-dead-letter-routing-key": d.deadRoutingKey,   // 指定死信routing-key
		"x-message-ttl":             10 * 1000,          // 消息过期时间,毫秒
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
	err = d.client.channel.QueueBind(d.queueName, d.routingKey, d.exchangeName, false, nil)
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
func (d *Dlx) PushMsgWithTimeout(routingKey string, msg any, delay int) error {
	d.routingKey = routingKey

	// 1、声明正常队列
	_, err := d.client.channel.QueueDeclare(d.queueName, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    d.deadExchangeName, // 指定死信交换机
		"x-dead-letter-routing-key": d.deadRoutingKey,   // 指定死信routing-key
	})
	if err != nil {
		return err
	}

	// 2、声明交换机:插件的交换机名称https://github.com/rabbitmq/rabbitmq-delayed-message-exchange
	err = d.client.channel.ExchangeDeclare(d.exchangeName, "x-delayed-message", true, false, false, false, amqp.Table{
		"x-delayed-type": "direct",
	})
	if err != nil {
		return err
	}

	// 3、绑定交换机
	err = d.client.channel.QueueBind(d.queueName, d.routingKey, d.exchangeName, false, nil)
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
		Headers:      amqp.Table{"x-delay": delay * 1000},
	})
}

func (d *Dlx) ConsumeMsg(queueName string) (<-chan amqp.Delivery, error) {
	return d.client.channel.Consume(queueName, "", false, false, false, false, nil)
}

func (d *Dlx) ConsumeDlxMsg(deadQueueName string) (<-chan amqp.Delivery, error) {
	return d.client.channel.Consume(deadQueueName, "", false, false, false, false, nil)
}
