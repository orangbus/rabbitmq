package test

import (
	"fmt"
	"github.com/goravel/framework/facades"
	"github.com/orangbus/rabbitmq/bootstrap"
	rabbit "github.com/orangbus/rabbitmq/facades"
	"log"
	"math/rand"
	"testing"
	"time"
)

func init() {
	bootstrap.Boot()
}

var (
	msg_open     = false
	publish_open = true
	routing_open = false
	top_open     = false

	publish_exchangeName = "publish_exchange"
	routing_exchangeName = "routing_exchange"
	top_exchangeName     = "top_exchange"

	routing_key = "routing_key"
	top_key     = "log.*.*"
)

func randomVal(list []string) string {
	randomIndex := rand.Intn(len(list))
	return list[randomIndex]
}

func TestConfig(t *testing.T) {
	host := facades.Config().GetString("rabbitmq.host")
	port := facades.Config().GetInt("rabbitmq.port")
	name := facades.Config().GetString("rabbitmq.username")
	password := facades.Config().GetString("rabbitmq.password")
	vhost := facades.Config().GetString("rabbitmq.vhost")
	sdn := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", name, password, host, port, vhost)
	t.Log(sdn)
}

func TestMsg(t *testing.T) {
	if msg_open {
		go func() {
			var total int64
			for {
				total++
				err := rabbit.Rabbitmq().Msg(fmt.Sprintf("%d 测试mq消息", total))
				if err != nil {
					log.Printf("rabbitmq 普通消息发送失败：%s", err.Error())
				}
				log.Printf("rabbitmq 普通消息发送成功：%d", total)
				time.Sleep(time.Second)
			}
		}()
	}

	// 订阅模式：这里面的消息可以有多个消费者消费
	if publish_open {
		go func() {
			var total2 int64
			for {
				total2++
				err := rabbit.Rabbitmq().Publish(publish_exchangeName, fmt.Sprintf("%d 订阅消息", total2))
				if err != nil {
					log.Printf("rabbitmq 订阅消息发送失败：%s", err.Error())
				}
				log.Printf("rabbitmq 订阅消息发送成功：%d", total2)
				if total2%2 == 0 {
					time.Sleep(time.Millisecond * 300)
				}
			}
		}()
	}

	// 路由模式：
	if routing_open {
		go func() {
			var total3 int64
			for {
				total3++
				if total3%2 == 0 {
					routing_key = "info"
				} else {
					routing_key = "error"
				}
				log.Printf("total3:%d -> %d -> %s", total3, total3%2, routing_key)
				err := rabbit.Rabbitmq().Routing(routing_exchangeName, routing_key, fmt.Sprintf("%d 【key:%s】路由消息", total3, routing_key))
				if err != nil {
					log.Printf("rabbitmq 路由消息发送失败：%s", err.Error())
				}
				log.Printf("[key:%s]rabbitmq 路由消息发送成功：%d", routing_key, total3)
				time.Sleep(time.Second)
			}
		}()
	}

	// 主题模式
	if top_open {
		go func() {
			var total4 int64
			//country := []string{"country1", "country2"}
			province := []string{"province1", "province2"}
			city := []string{"city1", "city2"}

			key := fmt.Sprintf("%s.%s.%s", "log", randomVal(province), randomVal(city))
			for {
				total4++
				err := rabbit.Rabbitmq().Topic(top_exchangeName, key, fmt.Sprintf("%d 【key:%s】主题消息", total4, key))
				if err != nil {
					log.Printf("rabbitmq 主题消息发送失败：%s", err.Error())
				}
				log.Printf("[key:%s]rabbitmq 主题消息发送成功：%d", key, total4)
				time.Sleep(time.Millisecond * 500)
			}
		}()
	}

	select {}
}

func TestConsume(t *testing.T) {
	msgs, err := rabbit.Rabbitmq().ConsumeMsg()
	if err != nil {
		t.Log(err.Error())
		return
	}

	var forever chan struct{}
	go func() {
		for data := range msgs {
			log.Printf("接收到mq的普通消息是：%s", string(data.Body))
			err := data.Ack(false)
			if err != nil {
				log.Printf("ack error: %s", err.Error())
			}
			time.Sleep(time.Second)
		}
	}()
	<-forever
}

func TestConsumePublish(t *testing.T) {
	msgs, err := rabbit.Rabbitmq().ConsumePublish(publish_exchangeName)
	if err != nil {
		t.Log(err.Error())
		return
	}

	forever := make(chan bool)
	go func() {
		for msg := range msgs {
			log.Printf("[%s]订阅消息：%s", publish_exchangeName, string(msg.Body))
			err := msg.Ack(false)
			if err != nil {
				log.Printf("ack error: %s", err.Error())
			}
			time.Sleep(time.Second)
		}
	}()
	<-forever

}
func TestConsumeRouting(t *testing.T) {
	msgs, err := rabbit.Rabbitmq().ConsumeRouting(routing_exchangeName, routing_key)
	if err != nil {
		t.Log(err.Error())
		return
	}
	forever := make(chan bool)
	go func() {
		for msg := range msgs {
			log.Printf("[%s:%s]路由消息：%s", routing_exchangeName, routing_key, string(msg.Body))
			err := msg.Ack(false)
			if err != nil {
				log.Printf("ack error: %s", err.Error())
			}
			time.Sleep(time.Second)
		}
	}()
	<-forever
}
func TestConsumeTopic(t *testing.T) {
	msgs, err := rabbit.Rabbitmq().ConsumeTopic(top_exchangeName, top_key)
	if err != nil {
		t.Log(err.Error())
		return
	}
	forever := make(chan bool)
	go func() {
		for msg := range msgs {
			log.Printf("[%s:%s]主题消息：%s", top_exchangeName, top_key, string(msg.Body))
			err := msg.Ack(false)
			if err != nil {
				log.Printf("ack error: %s", err.Error())
			}
			time.Sleep(time.Second)
		}
	}()
	<-forever
}
