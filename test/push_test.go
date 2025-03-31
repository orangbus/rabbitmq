package test

import (
	"fmt"
	"github.com/orangbus/rabbitmq/facades"
	"sync"
	"testing"
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
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			msgs, err := facades.Rabbitmq().ConsumeMsg()
			if err != nil {
				t.Log(err.Error())
				return
			}
			for msg := range msgs {
				t.Log(fmt.Sprintf("%d---%s", i, string(msg.Body)))
				//time.Sleep(time.Second)
				msg.Ack(false)
			}
		}(i)
	}
	wg.Wait()
	select {}
}
