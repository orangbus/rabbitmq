package test

import (
	"fmt"
	"github.com/orangbus/rabbitmq/facades"
	"log"
	"sync"
	"testing"
	"time"
)

func TestPushMsg(t *testing.T) {
	number := 10
	var wg sync.WaitGroup
	var wg2 sync.WaitGroup
	wg2.Add(number)
	for i := 0; i < number; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				if err := facades.Rabbitmq().Msg(fmt.Sprintf("[%d]this is push msg", j)); err != nil {
					t.Log(err.Error())
				}
			}
			wg2.Done()
			t.Log(fmt.Sprintf("%d over", i))
		}()
	}
	wg.Wait()
	wg2.Wait()
}

func TestConsumerPush(t *testing.T) {
	for i := 0; i < 2; i++ {
		go func() {
			msgs, err := facades.Rabbitmq().ConsumeMsg()
			if err != nil {
				log.Print(err.Error())
				return
			}
			for {
				data := <-msgs
				log.Printf(fmt.Sprintf("%s", string(data)))
				time.Sleep(time.Second)
			}
			//for msg := range msgs {
			//	log.Printf(fmt.Sprintf("%s", string(msg.Body)))
			//	time.Sleep(time.Second)
			//	if err := msg.Ack(false); err != nil {
			//		log.Printf("消息处理失败：%s", err.Error())
			//	}
			//}
		}()
	}

	select {}
}
