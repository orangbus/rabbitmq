package test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func worker(id int, restartChan chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Worker %d started\n", id)

	// 模拟随机运行时间
	runTime := time.Duration(rand.Intn(10)+5) * time.Second
	select {
	case <-time.After(runTime): // 模拟正常运行
		fmt.Printf("Worker %d finished\n", id)
	case <-time.After(time.Duration(rand.Intn(5)) * time.Second): // 模拟异常退出
		fmt.Printf("Worker %d crashed!\n", id)
		restartChan <- id // 通知主协程需要重启
	}
}
func TestCache(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	const workerCount = 5
	restartChan := make(chan int, workerCount) // 用于通知需要重启的 worker
	var wg sync.WaitGroup

	// 启动 5 个 worker
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(i, restartChan, &wg)
	}

	// 监控 goroutine
	go func() {
		for id := range restartChan {
			fmt.Printf("Restarting worker %d...\n", id)
			wg.Add(1)
			go worker(id, restartChan, &wg)
		}
	}()

	// 保持主 goroutine 运行
	wg.Wait()
}
