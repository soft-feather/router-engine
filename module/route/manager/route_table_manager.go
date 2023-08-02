package manager

import (
	"net"
	"time"
)

type Service struct {
	TableManager
}

func (s *Service) Init() error {
	tableManagerV1 := NewTableManagerV1()

	s.TableManager = tableManagerV1

	return nil
}

func (s *Service) Shutdown() {
}

type Route struct {
	Id                int64         //路由表id
	DestinationIp     net.IP        // 目标地址
	SubnetMask        net.IPMask    // 子网掩码
	NextHopAddress    net.IP        // 下一跳的IP
	OutgoingIp        net.IP        // 出口IP
	OutgoingInterface string        // 出口网卡
	HopCount          int           // 到达所需跳数
	RouteSource       string        // 路由来源
	LastUpdateTime    time.Time     // 更新时间
	RouteStatus       int           // 路由状态
	Priority          int           // 优先级/权重
	Bandwidth         int64         // 带宽（单位：KB）
	Load              int           // 负载
	Delay             time.Duration // 时延
}

type TableManager interface {
	AddRoute(route Route)                          // 添加路由
	UpdateRoute(route Route)                       // 修改路由
	DeleteRoute(route Route)                       // 删除路由
	GetAllRoutes() []Route                         // 获取全部路由
	CleanAllRoutes()                               // 删除全部路由
	DefaultCheckRoute(ip net.IP) Route             // 路由匹配
	CheckRouteYourShelf(func([]Route) Route) Route // 自定义路由匹配规则
}
