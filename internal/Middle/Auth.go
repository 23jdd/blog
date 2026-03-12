package middle

import (
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	// 从请求头中获取token
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(401, gin.H{
			"message": "unauthorized",
		}) // 401 错误
		c.Abort()
		return
	}
	// 验证token
	if token != "123456" {
		c.JSON(401, gin.H{
			"message": "unauthorized",
		}) // 401 错误
		c.Abort()
		return
	}
	// token验证通过，将token存储到上下文
	c.Set("token", token)
	c.Next()

}
