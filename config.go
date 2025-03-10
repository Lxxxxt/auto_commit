package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	ModelKey string `json:"model_key"`
	ApiKey   string `json:"api_key"`
	Platform string `json:"platform"`
}

func getConfig() Config {
	fullPath := getConfigPath()
	// 文件存在，打开文件进行后续操作
	file, err := os.OpenFile(fullPath, os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 这里可以添加代码来操作文件，比如读取或写入内容
	// 例如，读取文件内容
	content, err := os.ReadFile(fullPath)
	if err != nil {
		log.Fatal("读取配置文件失败:", err)
	}
	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Println("解析配置文件失败:", err)
		saveConfig(config)
	}
	return config
}
func getApiKey() string {
	config := getConfig()
	return config.ApiKey
}
func getModelKey() string {
	config := getConfig()
	return config.ModelKey
}
func getPlatform() string {
	config := getConfig()
	return config.Platform
}

func saveConfig(config Config) {
	fullPath := getConfigPath()
	// 以覆盖模式打开文件
	file, err := os.OpenFile(fullPath, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	jsonData, err := json.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Write(jsonData)
	if err != nil {
		log.Fatal(err)
	}
	// 保存并关闭文件
	err = file.Sync()
	if err != nil {
		log.Fatal(err)
	}

}
func setApiKey(key string) {
	config := getConfig()
	config.ApiKey = key
	saveConfig(config)
	fmt.Println("api key 已设置")
}
func setModelKey(key string) {
	config := getConfig()
	config.ModelKey = key
	saveConfig(config)
	fmt.Println("model key 已设置")
}
func setPlatform(platform string) {
	config := getConfig()
	config.Platform = platform
	saveConfig(config)
	fmt.Println("platform 已设置")
}
func getConfigPath() (fullPath string) {
	// 定义文件路径
	path := filepath.Join(".autoc", "autoc_config.json")

	// 获取当前用户的主目录路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// 构建完整的文件路径
	fullPath = filepath.Join(homeDir, path)

	// 检查目录是否存在，如果不存在则创建
	dir := filepath.Dir(fullPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	// 检查文件是否存在，如果不存在则创建
	_, err = os.Stat(fullPath)
	if os.IsNotExist(err) {
		file, err := os.Create(fullPath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	}
	return
}
