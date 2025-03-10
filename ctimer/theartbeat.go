package ctimer

import (
	"fmt"
	"time"

	"github.com/Cai-ki/cinx/ciface"
	"github.com/Cai-ki/cinx/cnet"
	"github.com/Cai-ki/cinx/crouter"
)

func StartHeartbeat(conn ciface.IConnection) {
	conn.SetProperty("lastActiveTime", time.Now())
	ticker := time.NewTicker(5 * time.Second) // 5秒发送一次心跳
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := conn.SendBuffMsg(crouter.MsgIDHeartbeatRequest, []byte("ping"))
			if err != nil {
				fmt.Println("[Cinx] Send heartbeat request error:", err)
				return
			}
		case <-conn.(*cnet.Connection).ExitBuffChan:
			return
		}
	}
}

func StartHeartbeatChecker(conn ciface.IConnection) {
	// 启动一个协程定期检查连接活性
	for {
		time.Sleep(10 * time.Second) // 10秒检查一次

		// 获取最后一次活跃时间
		lastActiveTime, ok := conn.GetProperty("lastActiveTime")
		if !ok {
			continue
		}

		// 超过 15 秒未收到消息，判定超时
		if time.Since(lastActiveTime.(time.Time)) > 15*time.Second {
			fmt.Println("Connection timeout, closing...")
			conn.Stop()
			return
		}
	}
}
