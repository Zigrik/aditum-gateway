package balancer

import (
	"sync/atomic"
)

type RoundRobin struct {
	counter uint64
}

func (rr *RoundRobin) Next(instances []string) string {
	if len(instances) == 0 {
		return ""
	}
	// атомарно увеличиваем счётчик и берём по модулю
	next := atomic.AddUint64(&rr.counter, 1) - 1
	return instances[next%uint64(len(instances))]
}
