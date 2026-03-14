package main

import (
	"blog/internal/config"
	"blog/internal/middle"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log"
	"os"
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
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("Error opening log file: %v", err))
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	port := cfg.Port
	router := gin.New()
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: logFile,
	}))
	router.Use(gin.Recovery())
	router.Use(middle.ReadLimitMiddlerWare(cfg.ReadLimit, cfg.Rate)) // 读取限制中间件
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "Hello World",
		})
	})
	log.Printf("Server starting on port %d", port)
	router.Run(fmt.Sprintf(":%d", port))
}
