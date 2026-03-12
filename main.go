package main

import (
	"blog/internal/config"
	"blog/internal/middle"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// @title Blog API
// @version 1.0
// @host localhost:8080
// @BasePath /
func MustInitConfig() *config.Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("read_limit", 100)
	viper.SetDefault("rate", 1)
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("Error reading config file: %v", err))
	}
	port := viper.GetInt("server.port")
	return &config.Config{
		Port:      port,
		ReadLimit: viper.GetInt64("read_limit"),
		Rate:      viper.GetInt64("rate"),
	}
}

func main() {
	cfg := MustInitConfig()
	port := cfg.Port
	router := gin.Default()
	router.Use(middle.ReadLimitMiddlerWare(cfg.ReadLimit, cfg.Rate) // 读取限制中间件
	log.Printf("Server starting on port %d", port)
	router.Run(fmt.Sprintf(":%d", port))
}
