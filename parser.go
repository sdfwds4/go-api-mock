package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RequestResponse struct {
	Request struct {
		URL    string `json:"url"`
		Method string `json:"method"`
	} `json:"request"`
	Response struct {
		Code  int         `json:"code"`
		Delay int         `json:"delay"`
		Body  interface{} `json:"body"`
	} `json:"response"`
}

func ParseDirectory(dir string) (map[string]map[string]*ResponseConfig, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	routes := make(map[string]map[string]*ResponseConfig)

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			filePath := filepath.Join(dir, f.Name())
			log.Println("Found api file: ", filePath, "...")
			data, err := os.ReadFile(filePath)
			if err != nil {
				log.Println("Cannot read api file: ", filePath)
				continue // 跳过错误文件
			}

			var items []RequestResponse
			if err := json.Unmarshal(data, &items); err != nil {
				log.Println("Cannot unmarshal api file: ", filePath)
				continue // 跳过解析失败文件
			}

			for _, item := range items {
				method := item.Request.Method
				path := item.Request.URL

				if _, ok := routes[method]; !ok {
					routes[strings.ToLower(method)] = make(map[string]*ResponseConfig)
				}

				routes[strings.ToLower(method)][path] = &ResponseConfig{
					Code:  item.Response.Code,
					Delay: time.Duration(item.Response.Delay) * time.Millisecond,
					Body:  item.Response.Body,
				}
				log.Println("Load method: ", path)
			}
		}
	}
	return routes, nil
}

// 新增单个文件解析方法
func ParseFile(filePath string) (map[string]map[string]*ResponseConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var items []RequestResponse
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	routes := make(FileRoutes)

	for _, item := range items {
		method := strings.ToUpper(item.Request.Method)
		path := item.Request.URL

		if _, ok := routes[method]; !ok {
			routes[method] = make(map[string]*ResponseConfig)
		}

		routes[method][path] = &ResponseConfig{
			Code:  item.Response.Code,
			Delay: time.Duration(item.Response.Delay) * time.Millisecond,
			Body:  item.Response.Body,
		}
	}
	return routes, nil
}
