package manager

import (
	"math"
	"net"
	"sort"
	"sync"
	"time"
)

type TableManagerV1 struct {
	sync.RWMutex
	routes  []Route
	routeId int64
}

func NewTableManagerV1() *TableManagerV1 {
	t := &TableManagerV1{
		RWMutex: sync.RWMutex{},
		routes:  make([]Route, 0, 8),
		routeId: 1,
	}

	return t
}

func (t *TableManagerV1) AddRoute(route Route) {
	t.Lock()
	defer t.Unlock()

	// 赋予唯一id
	route.Id = t.routeId
	t.routeId++

	// 预先计算目标网络地址，减少重复计算
	route.DestinationIp = route.DestinationIp.Mask(route.SubnetMask)

	t.routes = append(t.routes, route)

	// 按照子网掩码排序, 由大到小
	sort.Slice(t.routes, func(i, j int) bool {
		maskBitI, _ := t.routes[i].SubnetMask.Size()
		maskBitJ, _ := t.routes[j].SubnetMask.Size()

		return maskBitI > maskBitJ
	})
}

func (t *TableManagerV1) DeleteRoute(route Route) {
	t.Lock()
	defer t.Unlock()

	for i := range t.routes {
		if t.routes[i].Id == route.Id {
			t.routes = append(t.routes[:i], t.routes[i+1:]...)
		}
	}
}

func (t *TableManagerV1) GetAllRoutes() []Route {
	t.RLock()
	defer t.RUnlock()

	return t.routes[:]
}

func (t *TableManagerV1) CleanAllRoutes() {
	t.Lock()
	defer t.Unlock()

	t.routes = make([]Route, 0, 8)
}

func (t *TableManagerV1) UpdateRoute(route Route) {
	t.Lock()
	defer t.Unlock()

	// 根据id更新路由条目
	for i, r := range t.routes {
		if r.Id == route.Id {
			route.DestinationIp = route.DestinationIp.Mask(route.SubnetMask)
			route.LastUpdateTime = time.Now()
			t.routes[i] = route
		}
	}

	// 按照子网掩码排序, 由大到小
	sort.Slice(t.routes, func(i, j int) bool {
		maskBitI, _ := t.routes[i].SubnetMask.Size()
		maskBitJ, _ := t.routes[j].SubnetMask.Size()

		return maskBitI > maskBitJ
	})
}

func (t *TableManagerV1) DefaultCheckRoute(ip net.IP) Route {
	var (
		resRoute    Route
		maskMaxSize = math.MinInt
		minHopCount = math.MaxInt
	)

	for _, r := range t.routes {
		// 1、最长子网掩码原则
		size, _ := r.SubnetMask.Size()
		if size < maskMaxSize {
			break
		}

		if !ip.Mask(r.SubnetMask).Equal(r.DestinationIp) {
			continue
		}

		// 2、跳数最少原则
		if minHopCount <= r.HopCount {
			continue
		}

		maskMaxSize = size
		minHopCount = r.HopCount
		resRoute = r
	}

	return resRoute
}

func (t *TableManagerV1) CheckRouteYourShelf(f func([]Route) Route) Route {
	return f(t.routes)
}
