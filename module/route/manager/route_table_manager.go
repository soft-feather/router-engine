package manager

import "time"

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
	DestinationIp  string    // 目标地址
	SubnetMask     string    // 子网掩码
	NextHopAddress string    // 下一跳的IP
	OutgoingIp     string    // 出口IP
	HopCount       int       // 到达所需跳数
	RouteSource    string    // 路由来源
	LastUpdateTime time.Time // 更新时间
	RouteStatus    string    // 路由状态
	Priority       int       // 优先级/权重
}

type TableManager interface {
	AddRoute(route Route)                      // 添加路由
	DeleteRoute(destination string)            // 删除路由
	UpdateRoute(route Route)                   // 修改路由
	GetRoute(destination string) (Route, bool) // 获取路由
	GetAllRoutes() []Route                     // 获取全部路由
	CleanAllRoutes()                           // 删除全部路由
}
