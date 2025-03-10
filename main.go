package main

import (
	"log"
	"net/http"
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
	routeManager := NewRouteManager()

	// 初始加载路由
	loadRoutes := func() {
		newRoutes, err := ParseDirectory(cfg.APIPath)
		if err != nil {
			log.Printf("Error parsing directory: %v", err)
			return
		}
		routeManager.UpdateRoutes(newRoutes)
		log.Println("Routes reloaded")
	}
	loadRoutes()

	// 启动文件监听
	go WatchDirectory(cfg.APIPath, loadRoutes)

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
