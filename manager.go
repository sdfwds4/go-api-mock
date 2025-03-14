package main

import (
	"sort"
	"sync"
	"time"
)

type responseConfig struct {
	Code  int
	Delay time.Duration
	Body  interface{}
}

type FileRoutes map[string]map[string]*responseConfig // method -> path -> config
type routeManager struct {
	mu         sync.RWMutex
	fileRoutes map[string]FileRoutes // 文件路径 -> 路由配置
}

func newRouteManager() *routeManager {
	return &routeManager{
		fileRoutes: make(map[string]FileRoutes),
	}
}

// 更新单个文件的路由配置
func (rm *routeManager) UpdateFileRoutes(file string, routes FileRoutes) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.fileRoutes[file] = routes
}

// 删除指定文件的路由配置
func (rm *routeManager) RemoveFile(file string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.fileRoutes, file)
}

// 获取路由配置（按文件名排序，后加载的文件优先级更高）
func (rm *routeManager) GetConfig(method, path string) *responseConfig {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// 获取排序后的文件名列表
	files := make([]string, 0, len(rm.fileRoutes))
	for f := range rm.fileRoutes {
		files = append(files, f)
	}
	sort.Strings(files)

	// 逆序查找，后加载的文件优先
	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]
		if routes, ok := rm.fileRoutes[file]; ok {
			if methodRoutes, ok := routes[method]; ok {
				if config, ok := methodRoutes[path]; ok {
					return config
				}
			}
		}
	}
	return nil
}
