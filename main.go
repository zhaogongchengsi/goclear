package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// Config 定义配置文件结构
type Config struct {
	Root  string   `json:"root"`
	Match []string `json:"match"`
}

type ClearDirs []Config

func main() {
	// 1. 从环境变量读取配置文件路径
	configPath := os.Getenv("CLEANER_CONFIG")
	log.Printf("配置文件路径: %s", configPath)
	if configPath == "" {
		log.Fatal("环境变量 CLEANER_CONFIG 未设置")
	}

	// 2. 读取并解析配置文件
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 使用 WaitGroup 等待所有协程完成
	var wg sync.WaitGroup

	// 遍历配置并启动协程
	for _, dir := range *config {
		wg.Add(1)
		go func(dir Config) {
			defer wg.Done()
			if err := cleanDirectories([]Config{dir}); err != nil {
				log.Printf("清理失败: %v", err)
			}
		}(dir)
	}

	// 等待所有协程完成
	wg.Wait()

	fmt.Println("清理完成!")
}

// 加载配置文件
func loadConfig(path string) (*ClearDirs, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	var config ClearDirs
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	return &config, nil
}

// 清理目录
func cleanDirectories(dirs []Config) error {
	for _, dir := range dirs {
		// 检查 root 目录是否存在
		if _, err := os.Stat(dir.Root); os.IsNotExist(err) {
			log.Printf("目录不存在: %s", dir.Root)
			continue
		}

		var matches []string

		for _, pattern := range dir.Match {
			match, err := filepath.Glob(filepath.Join(dir.Root, pattern))
			if err != nil {
				return fmt.Errorf("无法匹配模式 %s: %w", pattern, err)
			}
			matches = append(matches, match...)
		}

		for _, match := range matches {
			fmt.Printf("[清理] %s\n", match)
			if err := os.RemoveAll(match); err != nil {
				return fmt.Errorf("无法清理 %s: %w", match, err)
			}
		}
	}
	return nil
}
