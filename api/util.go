package api

import (
	"sync"
	"time"
)

var (
	limit   = make(map[string]int)
	limitMx sync.Mutex
)

func limitEndpoint(timer int, endpoint string) bool {
	limitMx.Lock()
	endpointLimit := 0
	switch endpoint {
	default:
		endpointLimit = 30
	}
	if limit[endpoint] > endpointLimit {
		limitMx.Unlock()
		return false
	}
	limit[endpoint] += 1
	limitMx.Unlock()
	go func(timer int) {
		time.Sleep(time.Duration(timer * 1e9))
		limitMx.Lock()
		limit[endpoint] -= 1
		limitMx.Unlock()
	}(timer)
	return true
}
