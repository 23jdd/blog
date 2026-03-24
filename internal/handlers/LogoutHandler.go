package handlers

import (
	"blog/internal/redis"
	"blog/internal/types"
	"blog/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary		Logout
// @Description	Logout
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			req	body		types.RefreshTokenRequest	false	"可选，携带 refresh_token 用于同步撤销"
// @Success		200	{object}	types.SuccessResponse
// @Failure		401	{object}	types.ErrorResponse
// @Router			/auth/logout [post]
func LogoutHandler(ctx *gin.Context) {
	exp, exists := ctx.Get("exp") // 获取token过期时间
	if !exists {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "未授权",
		})
		return
	}
	jti, exists := ctx.Get("jti") // 获取 jti
	if !exists {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "未授权",
		})
		return
	}

	expTime, ok := exp.(time.Time) // 将过期时间转换为 time.Time 类型
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "未授权",
		})
		return
	}

	accessJTI, ok := jti.(string) // 将 jti 转换为 string 类型
	if !ok || accessJTI == "" {   // 如果 jti 转换为 string 类型失败，或者 jti 为空，返回 401 错误
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "未授权",
		})
		return
	}

	ttl := time.Until(expTime)
	if ttl > 0 { // 如果 ttl 大于 0，将 access token jti 加入黑名单
		_ = redis.BlacklistAccessToken(accessJTI, ttl) // 将 access token jti 加入黑名单
	}
	// 如果 refresh token 存在，删除刷新令牌
	var req types.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
		if refreshClaims, err := utils.ValidateRefreshToken(req.RefreshToken); err == nil {
			_ = redis.DeleteRefreshToken(refreshClaims.ID) // 删除刷新令牌
		}
	}
	// 如果登出成功，返回 200 状态码
	ctx.JSON(http.StatusOK, types.SuccessResponse{
		Message: "登出成功",
	})
}
