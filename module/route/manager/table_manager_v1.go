package manager

import (
	"errors"
	"math"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TableManagerV1 struct {
	sync.RWMutex
	routes  []*Route
	routeId int64
}

func NewTableManagerV1() *TableManagerV1 {
	t := &TableManagerV1{
		RWMutex: sync.RWMutex{},
		routes:  make([]*Route, 0, 8),
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
	ipv4Num, err := ParseIpV4StringToNum(route.DestinationIp.String())
	if err == nil {
		route.IpV4Num = ipv4Num
	}

	t.routes = append(t.routes, &route)

	// 路由条目聚合
	t.RouteMerge()

	// 按照子网掩码排序, 由大到小
	t.SortBySubnetMask()
}

// RouteMerge 路由条目合并
// 当路由条目相等时：192.192.100.0/24 && 192.192.1.0/24 ---->192.192.0.0/16
func (t *TableManagerV1) RouteMerge() {
	// 少于两条无需合并
	if len(t.routes) <= 1 {
		return
	}

	res := make([]*Route, 0, len(t.routes))
	for i := 0; i < len(t.routes)-1; i++ {
		sourceRoute := t.routes[i]
		nextHop := sourceRoute.OutgoingIp

		for j := i + 1; j < len(t.routes); j++ {
			mergeRoute := t.routes[j]
			// 当前位置路由被已被清除则跳过
			if mergeRoute == nil {
				continue
			}

			// 下一跳不相等跳过合并
			if !nextHop.Equal(mergeRoute.OutgoingIp) {
				continue
			}

			// 下一跳相等尝试合并
			// 先选出最小的子网掩码作为最大校验位数
			maxBitNum := 0
			subBit1, _ := sourceRoute.SubnetMask.Size()
			subBit2, _ := mergeRoute.SubnetMask.Size()

			maxBitNum = subBit1
			if subBit2 < subBit1 {
				maxBitNum = subBit2
			}

			// 两个目标IP先截取剩下最小的子网掩码位数
			subBitNum := 32 - maxBitNum
			sourceIpNum := sourceRoute.IpV4Num >> subBitNum
			mergeIpNum := mergeRoute.IpV4Num >> subBitNum

			// 循环比较，直到地址相同为止，或者子网掩码被减少至0位时，则为无法合并
			for ; subBitNum > 0; subBitNum-- {
				if sourceIpNum-mergeIpNum == 0 {
					break
				}

				sourceIpNum = sourceIpNum >> 1
				mergeIpNum = mergeIpNum >> 1
			}

			// 如果子网掩码位数为0则无法合并
			if subBitNum == 0 {
				continue
			}

			// 子网掩码不为0则进行条目合并
			sourceRoute.IpV4Num = sourceIpNum
			sourceRoute.DestinationIp = ParseNumToIpv4(sourceIpNum)
			sourceRoute.SubnetMask = net.CIDRMask(subBitNum, 32)
			sourceRoute.LastUpdateTime = time.Now()
			sourceRoute.RouteSource = "merge"

			// 清除被合并成功的路由
			t.routes[j] = nil
		}

		// 提交当前已合并完成的路由条目
		res = append(res, sourceRoute)
	}

	t.routes = res
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

	res := make([]Route, len(t.routes))
	for _, route := range t.routes {
		res = append(res, *route)
	}

	return res
}

func (t *TableManagerV1) CleanAllRoutes() {
	t.Lock()
	defer t.Unlock()

	t.routes = make([]*Route, 0, 8)
}

func (t *TableManagerV1) UpdateRoute(route Route) {
	t.Lock()
	defer t.Unlock()

	// 根据id更新路由条目
	for i, r := range t.routes {
		if r.Id == route.Id {
			route.DestinationIp = route.DestinationIp.Mask(route.SubnetMask)
			ipv4Num, err := ParseIpV4StringToNum(route.DestinationIp.String())
			if err == nil {
				route.IpV4Num = ipv4Num
			}

			route.LastUpdateTime = time.Now()
			t.routes[i] = &route

			break
		}
	}

	// 路由聚合
	t.RouteMerge()

	// 子网掩码排序
	t.SortBySubnetMask()
}

func (t *TableManagerV1) SortBySubnetMask() {
	// 按照子网掩码排序, 由大到小
	sort.Slice(t.routes, func(i, j int) bool {
		maskBitI, _ := t.routes[i].SubnetMask.Size()
		maskBitJ, _ := t.routes[j].SubnetMask.Size()

		return maskBitI > maskBitJ
	})
}

func (t *TableManagerV1) DefaultCheckRoute(ip net.IP) Route {
	var (
		resRoute    *Route
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

	return *resRoute
}

func (t *TableManagerV1) CheckRouteYourShelf(ip net.IP, f func([]Route, net.IP) Route) Route {
	return f(t.GetAllRoutes(), ip)
}

func ParseIpV4StringToNum(ip string) (uint32, error) {
	split := strings.Split(ip, ".")
	if len(split) != 4 {
		return 0, errors.New("非ipv4地址")
	}

	errInvalidIP := errors.New("IPv4地址解析失败")

	res := 0
	num, err := strconv.Atoi(split[0])
	if err != nil {
		return 0, err
	}
	if num < 0 || num > 255 {
		return 0, errInvalidIP
	}
	res += num << 24

	num, err = strconv.Atoi(split[1])
	if err != nil {
		return 0, err
	}
	if num < 0 || num > 255 {
		return 0, errInvalidIP
	}
	res += num << 16

	num, err = strconv.Atoi(split[2])
	if err != nil {
		return 0, err

	}
	if num < 0 || num > 255 {
		return 0, errInvalidIP
	}
	res += num << 8

	num, err = strconv.Atoi(split[3])
	if err != nil {
		return 0, err
	}
	if num < 0 || num > 255 {
		return 0, errInvalidIP
	}
	res += num

	return uint32(res), nil
}

func ParseNumToIpv4(num uint32) net.IP {
	ipBytes := make([]byte, 4)

	for i := 0; i < 4; i++ {
		ipBytes[3-i] = byte(num)
		num = num >> 8
	}

	return net.IPv4(ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3])
}
