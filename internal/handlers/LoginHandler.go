package handlers

import (
	"blog/internal/types"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary Login
// @Description Login
// @Tags auth
// @Accept json
// @Produce json
// @Param req body types.LoginRequest true "Login Request"
// @Success 200 {object} types.LoginResponse
// @Failure 400 {object} types.ErrorResponse
// @Router /login [post]
func LoginHandler(c *gin.Context) {
	var req types.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "Param error",
		}) // 400 错误
		return
	}
	//TODO : 验证用户名和密码
	if req.Username != "admin" || req.Password != "123456" {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "unauthorized",
		}) // 401 错误
		return
	}
	// 登录成功，返回token
	c.JSON(http.StatusOK, types.LoginResponse{
		Token: "123456",
	}) // 200 成功
}

// @Summary Register
// @Description Register
// @Tags auth
// @Accept json
// @Produce json
// @Param req body types.LoginRequest true "Register Request"
// @Success 200 {object} types.LoginResponse
// @Failure 400 {object} types.ErrorResponse
// @Router /register [post]
func RegisterHandler(c *gin.Context) {
	var req types.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "Param error",
		}) // 400 错误
		return
	}
	//TODO Register
	// 注册成功，返回token
	c.JSON(http.StatusOK, types.LoginResponse{
		Token: "123456",
	}) // 200 成功
}
func JudgeToken(c *gin.Context) {
	_, exists := c.Get("token") // 从上下文获取token
	if !exists {
		c.JSON(401, gin.H{
			"message": "unauthorized",
		})
		return
	}
	c.Redirect(http.StatusFound, "/main") // 重定向到主页面
}
