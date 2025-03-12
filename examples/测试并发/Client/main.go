package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Cai-ki/cinx/ciface"
	"github.com/Cai-ki/cinx/cnet"
)

type helloRouter struct {
	cnet.BaseRouter
	f func()
}

func (h *helloRouter) Handle(request ciface.IRequest) {
	h.f()
	fmt.Println(request.GetConn().RemoteAddr(), "received:", string(request.GetData()))
}

const (
	concurrency  = 200    // 并发连接数
	requestCount = 100000 // 总请求数
)

func main() {
	var wg sync.WaitGroup
	success := atomic.Int32{}
	startTime := time.Now()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wg2 := sync.WaitGroup{}
			client := cnet.NewClient("Client", "tcp4", "127.0.0.1", 7777)
			client.AddRouter(0, &helloRouter{f: func() {
				success.Add(1)
				wg2.Done()
			}})
			client.Start()
			// fmt.Println(client.Conn().(*cnet.Connection).IsClosed.Load())
			ct := requestCount / concurrency
			for ct > 0 {
				wg2.Add(1)
				client.Conn().SendMsg(0, []byte("hello"))
				ct--
			}
			wg2.Wait()
			client.Stop()
		}()
	}
	wg.Wait()

	duration := time.Since(startTime).Seconds()
	fmt.Printf("QPS: %.2f\n", float64(success.Load())/duration)
	fmt.Printf("总连接数: %d\n", concurrency)
}
