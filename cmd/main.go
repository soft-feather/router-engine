package main

import (
	"github.com/soft-feather/router-engine/module/packet"
	"github.com/soft-feather/router-engine/module/route/table/manager"
)

// ModuleType
// 一个模块要实现这些函数 Shutdown() 不需要返回 error, 因为调用 Shutdown 主程序不需要知道这个信息
type ModuleType interface {
	Init() error
	Shutdown()
}

// Register
// 将模块注册到主程序中进行维护
// Register 是主程序的加载模块的工具
// Register 使用方式 :
//  1. r.Init()
//  2. r.RegisterModule(moduleInstants)
//  3. r.Run()
//
// ...
// 当需要关闭程序时 ,  r.Shutdown()
type Register struct {
	moduleList    []ModuleType
	runModuleList []ModuleType
}

func (r *Register) Init() {
	r.runModuleList = make([]ModuleType, 0, 10)
	r.moduleList = make([]ModuleType, 0, 10)
}

func (r *Register) RegisterModule(module ModuleType) {
	r.moduleList = append(r.moduleList, module)
}

func (r *Register) Run() {
	var err error

	for _, module := range r.moduleList {
		if err := module.Init(); err != nil {
			// TODO:错误日志输出

			break
		}

		// 启动成功加入到已启动模块列表
		r.runModuleList = append(r.runModuleList, module)
	}

	// 启动错误则关闭已启动的模块
	if err != nil {
		r.Shutdown()
	}
}

func (r *Register) Shutdown() {
	for _, module := range r.runModuleList {
		module.Shutdown()
	}
}

func main() {
	r := &Register{}

	// 模块注册
	r.RegisterModule(new(packet.Server))
	r.RegisterModule(new(manager.Service))

	// 启动
	r.Run()
}
