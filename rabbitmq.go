package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/facades"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
	"time"
)

type Rabbitmq struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	prefetchCount int

	ctx       context.Context
	queueName string // 队列的名称
	key       string // 路由key
	exchange  string // 交换机

	mu sync.Mutex // 互斥锁，用于保护并发访问
}

func NewRabbitmq() (*Rabbitmq, error) {
	host := facades.Config().GetString("rabbitmq.host")
	port := facades.Config().GetInt("rabbitmq.port")
	name := facades.Config().GetString("rabbitmq.username")
	password := facades.Config().GetString("rabbitmq.password")
	vhost := facades.Config().GetString("rabbitmq.vhost")
	sdn := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", name, password, host, port, vhost)
	queueName := facades.Config().GetString("rabbitmq.queue")
	prefetchCount := facades.Config().GetInt("rabbitmq.prefetchCount", 10)

	var mq = &Rabbitmq{}
	var err error
	mq.conn, err = amqp.DialConfig(sdn, amqp.Config{Heartbeat: 60})
	if err != nil {
		return nil, err
	}
	mq.ctx = context.Background()
	mq.queueName = queueName
	mq.channel, err = mq.conn.Channel()
	mq.prefetchCount = prefetchCount
	if err != nil {
		return nil, err
	}
	return mq, nil
}

// 检查通道是否关闭
func (r *Rabbitmq) checkChannel() error {
	if r.conn.IsClosed() {
		return errors.New("rabbitmq 连接已经关闭")
	}
	if r.channel != nil && !r.channel.IsClosed() {
		return nil
	}
	var err error
	r.channel, err = r.conn.Channel()
	if err != nil {
		return errors.New(fmt.Sprintf("channel重连错误：%s", err.Error()))
	}
	log.Println("channel 已经重新连接")
	return nil
}

// 声明交换机
func (r *Rabbitmq) declareExchange(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	return r.channel.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args)
}

// 声明队列
func (r *Rabbitmq) declareQueue(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return r.channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}

func (c *Rabbitmq) Dlx() *Dlx {
	return NewDlx(c)
}

func (c *Rabbitmq) Close() {
	if err := c.channel.Close(); err != nil {
		log.Printf("rabbitmq channel close error: %v", err)
	}
	if err := c.conn.Close(); err != nil {
		log.Printf("rabbitmq conn close error: %v", err)
	}
}

// 统一发送消息
func (r *Rabbitmq) seed(data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.channel.PublishWithContext(r.ctx, "", r.queueName, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json", // text/plain
		Body:         body,
	})
}

// 接受消息
func consume(r *Rabbitmq, ch chan<- []byte) {
	notifyClose := r.channel.NotifyClose(make(chan *amqp.Error))
	msgs, err := r.channel.Consume(
		r.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		facades.Log().Errorf("接受消息失败：%s", err.Error())
		close(ch)
		return
	}
	//go func() {
	//	time.Sleep(time.Second * 2)
	//	r.channel.Close()
	//	log.Println("模拟通道关闭")
	//}()

	for {
		select {
		case <-notifyClose:
			facades.Log().Errorf("Channel auto closed, attempting to reconnect")
			// 尝试重新连接
			if err := r.checkChannel(); err != nil {
				facades.Log().Errorf("Failed to reconnect channel: %s", err.Error())
				r.Close()
				close(ch)
				return
			}
			// 重新声明队列
			if _, err := r.channel.QueueDeclare(r.queueName, true, false, false, false, nil); err != nil {
				facades.Log().Errorf("Failed to redeclare queue: %v", err)
				r.Close()
				close(ch)
				return
			}
			// 重新开始消费
			msgs, err = r.channel.Consume(
				r.queueName,
				"",
				false,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				facades.Log().Errorf("Failed to restart consuming: %v", err)
				r.Close()
				close(ch)
				return
			}
			notifyClose = r.channel.NotifyClose(make(chan *amqp.Error))
			facades.Log().Errorf("Successfully reconnected and restarted consuming")
		case msg, ok := <-msgs:
			if !ok {
				facades.Log().Errorf("消息通道已关闭，等待10秒...")
				time.Sleep(time.Second * 10)
				continue
			}
			if err := msg.Ack(false); err != nil {
				facades.Log().Errorf("消息确认失败,停留3秒:%s", err.Error())
				time.Sleep(time.Second * 3)
				continue
			}
			ch <- msg.Body
		}
	}
}

/*
发送一个普通消息，由默认的队列接受消息
name: 队列名称，设为空字符串 "" 时由 RabbitMQ 自动生成唯一名称。
durable: 是否持久化，设为 false 表示非持久化（重启后丢失）。
autoDelete: 设为 false 表示不是自动删除（这里由连接关闭触发删除）。
exclusive: 设为 true 表示该队列是独占的，只有当前连接可以访问。当连接关闭时，队列会被自动删除。
noWait: 是否不等待服务器确认，设为 false 表示等待确认
*/
func (r *Rabbitmq) Msg(data any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.checkChannel(); err != nil {
		return err
	}

	_, err := r.declareQueue(r.queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}
	return r.seed(data)
}

/*
发布订阅模式：消息发送给交换机，由交换机路由到队列，由队列接受消息，可以由多个消费者消息
*/
func (r *Rabbitmq) Publish(exchangeName string, data interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.checkChannel(); err != nil {
		return err
	}

	// 1、定义交换机
	err := r.declareExchange(exchangeName, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	marshal, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := r.checkChannel(); err != nil {
		return err
	}
	return r.channel.PublishWithContext(r.ctx, exchangeName, "", false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         marshal,
	})
}

/*
路由模式：消息发送给交换机，但是可以携带一个key，消费者也会有一个key去接受对应keu的消息
举例：系统产生了错误消息，但是要求 A消费者只能接受 key=info 的消息，B消费者只能接受 key=debug 的消息， ....
*/
func (r *Rabbitmq) Routing(exchangeName, key string, data interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.checkChannel(); err != nil {
		return err
	}

	// 1、定义 direct 类型的交换机
	if err := r.declareExchange(exchangeName, "direct", true, false, false, false, nil); err != nil {
		return err
	}
	marshal, err2 := json.Marshal(data)
	if err2 != nil {
		return err2
	}
	if err := r.checkChannel(); err != nil {
		return err
	}

	return r.channel.PublishWithContext(r.ctx, exchangeName, key, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        marshal,
	})
}

func (r *Rabbitmq) Topic(exchangeName, key string, data interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.checkChannel(); err != nil {
		return err
	}
	// 1、定义交换机
	if err := r.declareExchange(exchangeName, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	marshal, err2 := json.Marshal(data)
	if err2 != nil {
		return err2
	}

	if err := r.checkChannel(); err != nil {
		return err
	}
	return r.channel.PublishWithContext(r.ctx, exchangeName, key, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        marshal,
	})
}

func (r *Rabbitmq) ReceiverTopic(exchangeName, key string) (<-chan amqp.Delivery, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.checkChannel(); err != nil {
		return nil, err
	}
	// 1、定义交换机
	err := r.declareExchange(exchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	// 2、定义消息队列
	q, err := r.declareQueue(r.queueName, true, false, true, false, nil)
	if err != nil {
		return nil, err
	}

	// 3、绑定:队列名称 key 交换机
	if err := r.channel.QueueBind(r.queueName, key, exchangeName, false, nil); err != nil {
		return nil, err
	}

	return r.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
}

func (r *Rabbitmq) ConsumeMsg() (<-chan []byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查通道是否OK
	if err := r.checkChannel(); err != nil {
		return nil, err
	}
	// 申请队列
	if _, err := r.declareQueue(r.queueName, true, false, false, false, nil); err != nil {
		return nil, err
	}
	if err := r.channel.Qos(r.prefetchCount, 0, false); err != nil {
		return nil, err
	}

	ch := make(chan []byte, 1)
	go consume(r, ch)
	return ch, nil
}

/*
*
接受订阅模式的消息,多个消费者收到的消息是一样的（类似把一则消息广播给多个人，每个人收到的消息是一致的）
*/
func (r *Rabbitmq) ConsumePublish(exchangeName string, queueName ...string) (<-chan []byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	customQueueName := ""
	durable := false // 指定了队列的时候就持久化
	if len(queueName) > 0 {
		customQueueName = queueName[0]
		durable = true
	}
	r.queueName = customQueueName
	if err := r.checkChannel(); err != nil {
		return nil, err
	}
	// 1、声明交换机
	err := r.declareExchange(exchangeName, "fanout", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	// 2、声明一个队列，队名的名称会随机生成
	q, err := r.declareQueue(customQueueName, durable, false, !durable, false, nil)
	if err != nil {
		return nil, err
	}
	if len(queueName) > 0 {
		r.queueName = q.Name
	}
	if err := r.channel.Qos(r.prefetchCount, 0, false); err != nil {
		return nil, err
	}
	// 3、交换机绑定上面创建的队列
	err = r.channel.QueueBind(
		q.Name,       // queue name：这里的队名名称会随机生成
		"",           // routing key
		exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// 4、消费消息
	ch := make(chan []byte, 1)
	go consume(r, ch)
	return ch, nil
}

/*
*
接受路由消息：当前消费者只会消费当前交换机产生指定key的消息
*/
func (r *Rabbitmq) ConsumeRouting(exchangeName, key string, queueName ...string) (<-chan []byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	customQueueName := ""
	durable := false // 指定了队列的时候就持久化
	if len(queueName) > 0 {
		customQueueName = queueName[0]
		durable = true
	}
	if err := r.checkChannel(); err != nil {
		return nil, err
	}
	// 1、声明交换机
	err := r.declareExchange(exchangeName, "direct", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	// 2、声明一个队列，队名的名称会随机生成
	q, err := r.declareQueue(customQueueName, durable, false, !durable, false, nil)
	if err != nil {
		return nil, err
	}
	if len(queueName) > 0 {
		r.queueName = q.Name
	}

	// 3、交换机绑定上面创建的队列
	err = r.channel.QueueBind(
		q.Name,       // queue name：这里的队名名称会随机生成
		key,          // routing key
		exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	if err := r.channel.Qos(r.prefetchCount, 0, false); err != nil {
		return nil, err
	}
	// 4、消费消息
	ch := make(chan []byte, 1)
	go consume(r, ch)
	return ch, nil
}

/*
*
可以精准的获取消息类型
china.yunnan.kunming (中国.云南.昆明)
*: 匹配一个单词,例如，如果队列绑定的Routing Key是user.*，那么它将匹配user.123、user.abc等Routing Key，但不会匹配user.123.456。
#:匹配零个或多个单词,如果队列绑定的Routing Key是user.#，那么它将匹配user、user.123、user.abc以及user.123.456等Routing Key。
*/
func (r *Rabbitmq) ConsumeTopic(exchangeName, key string, queueName ...string) (<-chan []byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	customQueueName := ""
	durable := false // 指定了队列的时候就持久化
	if len(queueName) > 0 {
		customQueueName = queueName[0]
		durable = true
	}

	if err := r.checkChannel(); err != nil {
		return nil, err
	}
	// 1、声明交换机
	err := r.declareExchange(exchangeName, "topic", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	// 2、声明一个队列，队名的名称会随机生成
	q, err := r.declareQueue(customQueueName, durable, false, !durable, false, nil)
	if err != nil {
		return nil, err
	}
	if len(queueName) > 0 {
		r.queueName = q.Name
	}

	// 3、交换机绑定上面创建的队列
	err = r.channel.QueueBind(
		q.Name,       // queue name：这里的队名名称会随机生成
		key,          // routing key
		exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	if err := r.channel.Qos(r.prefetchCount, 0, false); err != nil {
		return nil, err
	}
	// 4、消费消息
	ch := make(chan []byte, 1)
	go consume(r, ch)
	return ch, nil
}
