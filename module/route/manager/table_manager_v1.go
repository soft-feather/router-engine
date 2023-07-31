package manager

import (
	"sync"
	"time"
)

type TableManagerV1 struct {
	sync.RWMutex
	routes map[string]Route
}

func NewTableManagerV1() *TableManagerV1 {
	t := &TableManagerV1{
		RWMutex: sync.RWMutex{},
		routes:  make(map[string]Route),
	}

	return t
}

func (t *TableManagerV1) AddRoute(route Route) {
	t.Lock()
	defer t.Unlock()

	t.routes[route.DestinationIp] = route
}

func (t *TableManagerV1) DeleteRoute(destination string) {
	t.Lock()
	defer t.Unlock()

	delete(t.routes, destination)
}

func (t *TableManagerV1) UpdateRoute(route Route) {
	t.Lock()
	defer t.Unlock()

	route.LastUpdateTime = time.Now()
	t.routes[route.DestinationIp] = route
}

func (t *TableManagerV1) GetRoute(destination string) (Route, bool) {
	t.RLock()
	defer t.RUnlock()

	res, ok := t.routes[destination]

	return res, ok
}

func (t *TableManagerV1) GetAllRoutes() []Route {
	t.RLock()
	defer t.RUnlock()

	res := make([]Route, 0, len(t.routes))
	for _, route := range t.routes {
		res = append(res, route)
	}

	return res
}

func (t *TableManagerV1) CleanAllRoutes() {
	t.Lock()
	defer t.Unlock()

	t.routes = make(map[string]Route)
}
