package test

import (
	"fmt"
	"github.com/goravel/framework/facades"
	"github.com/orangbus/rabbitmq/bootstrap"
	rabbit "github.com/orangbus/rabbitmq/facades"
	"testing"
)

func init() {
	bootstrap.Boot()
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
	for i := 0; i < 100; i++ {
		err := rabbit.Rabbitmq().Msg(fmt.Sprintf("消息id:%d", i))
		if err != nil {
			t.Log(err.Error())
			return
		}
	}
	t.Log("done")
}

func TestConsume(t *testing.T) {
	rabbit.Rabbitmq().Consume()
}
