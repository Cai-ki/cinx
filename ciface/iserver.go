package ciface

type IServer interface {
	// 启动服务器
	Start()
	// 停止服务器
	Stop()
	// 执行业务服务
	Serve()
}
