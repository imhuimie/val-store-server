package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadConfig 从.env文件加载环境变量
func LoadConfig() error {
	// 尝试加载.env文件，如果文件不存在则忽略错误
	err := godotenv.Load()
	if err != nil {
		log.Printf("未找到.env文件，将使用系统环境变量: %v", err)
	}
	return nil
}

// GetEnv 获取环境变量，如果不存在则返回默认值
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
