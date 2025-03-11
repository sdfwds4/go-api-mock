package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {
	app, err := InitializeApplication()
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	app.Run()
}

type Application struct {
	cfg          *config
	routeManager *routeManager
	echo         *echo.Echo
}

func InitializeApplication() (*Application, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	routeManager := newRouteManager()
	if err := initRoutes(routeManager, cfg.APIPath); err != nil {
		return nil, fmt.Errorf("route init failed: %w", err)
	}

	e := setupEchoServer(routeManager)

	setupFileWatcher(routeManager, cfg.APIPath)

	return &Application{
		cfg:          cfg,
		routeManager: routeManager,
		echo:         e,
	}, nil
}

func (app *Application) Run() {
	address := fmt.Sprintf(":%d", app.cfg.Port)
	if err := app.echo.Start(address); err != nil && err != http.ErrServerClosed {
		app.echo.Logger.Fatal("Shutting down the server")
	}
}

func initRoutes(rm *routeManager, apiPath string) error {
	files, err := os.ReadDir(apiPath)
	if err != nil {
		log.Fatal(err)
	}

	// 按文件名排序加载
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			filePath := filepath.Join(apiPath, f.Name())
			routes, err := ParseFile(filePath)
			if err != nil {
				log.Printf("Error parsing %s: %v", filePath, err)
				continue
			}
			rm.UpdateFileRoutes(filePath, routes)
		}
	}
	return nil
}

func setupEchoServer(rm *routeManager) *echo.Echo {
	e := echo.New()
	e.Any("*", func(c echo.Context) error {
		// 获取请求信息
		method := c.Request().Method
		path := c.Request().URL.Path

		// 查找路由配置
		config := rm.GetConfig(method, path)
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
	return e
}

func setupFileWatcher(rm *routeManager, apiPath string) {
	handleChanges := func(files []string) {
		for _, filePath := range files {
			// 检查文件是否已被删除
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				rm.RemoveFile(filePath)
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

			rm.UpdateFileRoutes(filePath, routes)
			log.Printf("Updated routes from: %s", filePath)
		}
	}
	go watchDirectory(apiPath, time.Second, handleChanges)
}
