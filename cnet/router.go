package cnet

import "github.com/Cai-ki/cinx/ciface"

type BaseRouter struct{}

func (br *BaseRouter) PreHandle(req ciface.IRequest)  {}
func (br *BaseRouter) Handle(req ciface.IRequest)     {}
func (br *BaseRouter) PostHandle(req ciface.IRequest) {}
