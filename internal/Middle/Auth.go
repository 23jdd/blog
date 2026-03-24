package middle

import (
	"blog/internal/redis"
	"blog/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization") // 获取 Authorization 头
	if authHeader == "" {
		c.JSON(401, gin.H{
			"message": "unauthorized",
		})
		c.Abort()
		return
	} // 如果 Authorization 头为空，返回 401 错误
	// Bearer Token 格式
	parts := strings.SplitN(authHeader, " ", 2) // 分割 Authorization 头
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.JSON(401, gin.H{
			"message": "unauthorized",
		})
		c.Abort()
		return
	} // 如果 Authorization 头格式不正确，返回 401 错误

	token := parts[1]                               // 获取 token
	claims, err := utils.ValidateAccessToken(token) // 验证 token
	if err != nil {
		c.JSON(401, gin.H{
			"message": "unauthorized",
		})
		c.Abort()
		return
	} // 如果 token 格式不正确，返回 401 错误

	if claims.ID != "" { // 如果 token 的 ID 不为空，则检查 token 是否在黑名单中
		blacklisted, redisErr := redis.IsAccessTokenBlacklisted(claims.ID)
		if redisErr != nil || blacklisted {
			c.JSON(401, gin.H{
				"message": "unauthorized",
			})
			c.Abort()
			return
		}
	} else {
		c.JSON(401, gin.H{
			"message": "unauthorized",
		})
		c.Abort()
		return
	} // 如果 token 的 ID 为空，返回 401 错误
	c.Set("userID", claims.UserID)      // 设置 userID
	c.Set("token", token)               // 设置 token
	c.Set("exp", claims.ExpiresAt.Time) // 设置 exp
	c.Set("jti", claims.ID)             // 设置 jti
	c.Next()
}
