package main

import (
	"sync"
	"time"
)

type ResponseConfig struct {
	Code  int
	Delay time.Duration
	Body  interface{}
}

type RouteManager struct {
	mu     sync.RWMutex
	routes map[string]map[string]*ResponseConfig // method -> path -> config
}

func NewRouteManager() *RouteManager {
	return &RouteManager{
		routes: make(map[string]map[string]*ResponseConfig),
	}
}

func (rm *RouteManager) UpdateRoutes(newRoutes map[string]map[string]*ResponseConfig) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.routes = newRoutes
}

func (rm *RouteManager) GetConfig(method, path string) *ResponseConfig {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if methodMap, ok := rm.routes[method]; ok {
		if config, ok := methodMap[path]; ok {
			return config
		}
	}
	return nil
}
