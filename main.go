package main

import (
	middle "blog/internal/Middle"
	"blog/internal/config"
	"blog/internal/etcd"
	"blog/internal/handlers"
	"blog/internal/redis"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// @title		Blog API
// @version	1.0
// @host		localhost:8080
// @BasePath	/
func MustInitConfig() *config.Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("Error reading config file: %v", err))
	}
	cfg, err := config.LoadFromViper()
	if err != nil {
		panic(fmt.Sprintf("Error parsing config file: %v", err))
	}
	return cfg
}

func main() {

	cfg := MustInitConfig()            // 初始化配置
	fmt.Printf("cfg: %v", cfg)         // 打印配置
	etcdClient := etcd.NewEtcdClient() // 初始化 etcd 客户端
	etcdClient.Watch("limit.read", func(v string) {
		viper.Set("limit.read", v)
	})
	etcdClient.Watch("limit.rate", func(v string) {
		viper.Set("limit.rate", v)
	})
	// 兼容旧 etcd 键
	etcdClient.Watch("read_limit", func(v string) {
		viper.Set("limit.read", v)
	})
	etcdClient.Watch("rate", func(v string) {
		viper.Set("limit.rate", v)
	})

	fmt.Printf("etcdClient: %v", etcdClient)                                          // 打印 etcd 客户端
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666) // 初始化日志文件
	if err != nil {
		panic(fmt.Sprintf("Error opening log file: %v", err))
	}
	defer logFile.Close()  // 关闭日志文件
	log.SetOutput(logFile) // 设置日志输出
	redis.InitRedis()      // 初始化 redis
	defer func() {
		_ = redis.CloseRedis() // 关闭 redis
	}()
	port := cfg.Port    // 获取端口
	router := gin.New() // 初始化 gin 路由
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: logFile,
	})) // 设置日志输出
	router.Use(gin.Recovery())
	router.Use(middle.CrosMiddleware())                                     // swagger 中间件
	router.Use(middle.ReadLimitMiddlerWare(cfg.Limit.Read, cfg.Limit.Rate)) // 读取限制中间件
	router.Use(middle.RequestLogMiddleware())                               // 请求日志中间件
	router.Static(cfg.Upload.URLPrefix, "./"+cfg.Upload.Dir)                // 静态文件

	// 前端静态资源托管（生产部署用）
	frontDist := "./front/dist"
	if _, err := os.Stat(frontDist + "/index.html"); err == nil {
		router.GET("/", func(ctx *gin.Context) { ctx.File(frontDist + "/index.html") })
		// Vite 默认把打包产物放到 /assets 下，这里只托管 assets 目录即可
		router.Static("/assets", frontDist+"/assets")
	} else if _, err := os.Stat("./front/index.html"); err == nil {
		// 未构建 dist 时，仍然可以托管开发版页面（保证你能直接从 HTTP 访问）
		router.GET("/", func(ctx *gin.Context) { ctx.File("./front/index.html") })
		router.GET("/app.js", func(ctx *gin.Context) { ctx.File("./front/app.js") })
	} else {
		router.GET("/", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{
				"message": "Hello World",
			})
		}) // 根路径（开发/未构建前端时）
	}

	authGroup := router.Group("/auth") // 认证路由组
	{
		authGroup.POST("/login", handlers.LoginHandler)                          // 登录
		authGroup.POST("/register", handlers.RegisterHandler)                    // 注册
		authGroup.POST("/refresh", handlers.RefreshTokenHandler)                 // 刷新 token
		authGroup.GET("/verification/send", handlers.SendVerificationCode)       // 发送验证码
		authGroup.GET("/judgeToken", middle.AuthMiddleware, handlers.JudgeToken) // 判断 token 是否有效
		authGroup.POST("/logout", middle.AuthMiddleware, handlers.LogoutHandler) // 登出
	}

	articleGroup := router.Group("/articles") // 文章路由组
	{
		// 固定路径须注册在 /:id 之前，否则会被当成数字 id 解析失败
		articleGroup.GET("/search", handlers.SearchArticles)                      // 搜索文章
		articleGroup.GET("/hot", handlers.GetHotArticles)                         // 获取热门文章
		articleGroup.GET("/by-tag", handlers.GetArticlesByTag)                    // 获取标签文章
		articleGroup.GET("/author/:authid", handlers.GetArticlesByAuthorID)       // 获取作者文章
		articleGroup.GET("/category/:categoryID", handlers.GetArticlesByCategory) // 获取分类文章
		articleGroup.GET("/:id", handlers.GetArticleByID)                         // 获取文章
		articleGroup.GET("/:id/comments", handlers.ListArticleComments)           // 获取文章评论
		articleGroup.GET("/:id/stats", handlers.GetArticleStats)                  // 获取文章统计

		articleGroup.Use(middle.AuthMiddleware)                                  // 认证中间件
		articleGroup.POST("", handlers.CreateArticle)                            // 创建文章
		articleGroup.PUT("/:id", handlers.UpdateArticle)                         // 更新文章
		articleGroup.PATCH("/:id/status", handlers.UpdateArticleStatus)          // 更新文章状态
		articleGroup.DELETE("/:id", handlers.DeleteArticle)                      // 删除文章
		articleGroup.POST("/:id/comments", handlers.CreateComment)               // 创建评论
		articleGroup.DELETE("/comments/:commentID", handlers.DeleteComment)      // 删除评论
		articleGroup.PATCH("/comments/:commentID/status", handlers.AuditComment) // 审核评论
		articleGroup.POST("/:id/likes", handlers.LikeArticle)                    // 点赞文章
		articleGroup.DELETE("/:id/likes", handlers.UnlikeArticle)                // 取消点赞文章
		articleGroup.POST("/:id/collections", handlers.CollectArticle)           // 收藏文章
		articleGroup.DELETE("/:id/collections", handlers.UnCollectArticle)       // 取消收藏文章
	}

	interactionGroup := router.Group("/interactions", middle.AuthMiddleware)
	{
		interactionGroup.GET("/my-collections", handlers.ListMyCollections) // 获取我的收藏
		interactionGroup.GET("/feed", handlers.GetMyFeed)                   // 获取我的 Feed
		interactionGroup.POST("/follow/:targetID", handlers.FollowUser)     // 关注用户
		interactionGroup.DELETE("/follow/:targetID", handlers.UnfollowUser) // 取消关注用户
	       // 获取我关注的人
	}

	fileGroup := router.Group("/file", middle.AuthMiddleware) // 文件路由组
	{
		fileGroup.POST("/setPersonImage", handlers.SetPersonImage)   // 设置个人头像
		fileGroup.POST("/uploadArticle", handlers.UpLoadArticleFile) // 上传文章文件

	}

	if cfg.Feature.EnableMarkdownAPI {
		router.POST("/markdown", handlers.MarkdownHandler) // markdown 转 html
	}

	draftGroup := router.Group("/drafts", middle.AuthMiddleware) // 草稿路由组
	{
		draftGroup.POST("", handlers.SaveDraft)         // 保存草稿
		draftGroup.GET("", handlers.ListDrafts)         // 草稿列表
		draftGroup.PUT("/:id", handlers.UpdateDraft)    // 修改草稿
		draftGroup.DELETE("/:id", handlers.DeleteDraft) // 删除草稿
	}

	configGroup := router.Group("/config", middle.AuthMiddleware) // 配置路由组
	{
		configGroup.GET("", handlers.GetDynamicConfig)     // 获取动态配置
		configGroup.POST("", handlers.UpdateDynamicConfig) // 更新动态配置
	}
	userGroup := router.Group("/user", middle.AuthMiddleware) // 用户路由组
	{
		userGroup.GET("/info", handlers.GetUserInfoHandler)       // 获取用户信息
		userGroup.POST("/update", handlers.UpdateUserInfoHandler) // 更新用户信息
		userGroup.DELETE("/delete", handlers.DeleteUserHandler)   // 删除用户
	}
	fmt.Printf("Server starting on port %d", port) // 启动服务器

	router.Run(fmt.Sprintf(":%d", port)) // 运行服务器
}
