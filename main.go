package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {
	// 加载配置
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// 初始化路由管理
	routeManager := newRouteManager()

	// 初始化加载（保持文件加载顺序）
	initRoutes := func() {
		files, err := os.ReadDir(cfg.APIPath)
		if err != nil {
			log.Fatal(err)
		}

		// 按文件名排序加载
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name() < files[j].Name()
		})

		for _, f := range files {
			if filepath.Ext(f.Name()) == ".json" {
				filePath := filepath.Join(cfg.APIPath, f.Name())
				routes, err := ParseFile(filePath)
				if err != nil {
					log.Printf("Error parsing %s: %v", filePath, err)
					continue
				}
				routeManager.UpdateFileRoutes(filePath, routes)
			}
		}
	}
	initRoutes()

	// 文件变化处理函数
	handleFileChanges := func(files []string) {
		for _, filePath := range files {
			// 检查文件是否已被删除
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				routeManager.RemoveFile(filePath)
				log.Printf("Removed routes for deleted file: %s", filePath)
				continue
			}

			// 只处理JSON文件
			if filepath.Ext(filePath) != ".json" {
				continue
			}

			// 解析并更新路由
			routes, err := ParseFile(filePath)
			if err != nil {
				log.Printf("Error reloading %s: %v", filePath, err)
				continue
			}

			routeManager.UpdateFileRoutes(filePath, routes)
			log.Printf("Updated routes from: %s", filePath)
		}
	}

	// 启动带防抖的文件监听（1秒）
	go WatchDirectory(cfg.APIPath, time.Second, handleFileChanges)

	// 创建Echo实例
	e := echo.New()

	// 注册通用处理程序
	e.Any("*", func(c echo.Context) error {
		// 获取请求信息
		method := c.Request().Method
		path := c.Request().URL.Path

		// 查找路由配置
		config := routeManager.GetConfig(method, path)
		if config == nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Not found"})
		}

		// 应用延迟
		if config.Delay > 0 {
			time.Sleep(config.Delay)
		}

		// 返回响应
		return c.JSON(config.Code, config.Body)
	})

	// 启动服务器
	e.Logger.Fatal(e.Start(":" + strconv.Itoa(cfg.Port)))
}
