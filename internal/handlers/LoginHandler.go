package handlers

import (
	"blog/internal/Log"
	"blog/internal/model"
	"blog/internal/redis"
	"blog/internal/sql"
	"blog/internal/types"
	"blog/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary		Login
// @Description	Login
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			req	body		types.LoginRequest	true	"Login Request"
// @Success		200	{object}	types.LoginResponse
// @Failure		400	{object}	types.ErrorResponse
// @Failure		401	{object}	types.ErrorResponse
// @Router			/auth/login [post]
func LoginHandler(c *gin.Context) {
	var req types.LoginRequest // 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "参数错误",
		})
		return
	} // 如果参数错误，返回 400 错误

	blocked, _, err := redis.IsLoginBlocked(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "登录风控检查失败",
		})
		return
	} // 如果登录风控检查失败，返回 500 错误
	if blocked {
		c.JSON(http.StatusTooManyRequests, types.ErrorResponse{
			Message: "登录失败次数过多，请稍后再试",
		})
		return
	} // 如果登录风控检查失败，返回 401 错误

	user, err := sql.NewUserMapper().FindByUsername(req.Username)
	if err != nil {
		_, _ = redis.IncreaseLoginFailCount(req.Username)
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "用户不存在",
		})
		return
	}
	if !utils.CheckPassword(req.Password, user.Password) {
		_, _ = redis.IncreaseLoginFailCount(req.Username)
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "密码错误",
		})
		return
	}

	_ = redis.ClearLoginFailCount(req.Username) // 清除登录失败计数
	// 生成访问令牌
	accessToken, err := utils.GenerateAccessToken(user.ID, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "生成访问令牌失败",
		})
		return
	}
	// 生成刷新令牌
	refreshToken, refreshClaims, err := utils.GenerateRefreshToken(user.ID, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "生成刷新令牌失败",
		})
		return
	}

	ttl := time.Until(refreshClaims.ExpiresAt.Time)
	// 保存刷新令牌
	if err = redis.SaveRefreshToken(refreshClaims.ID, user.ID, ttl); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "保存刷新令牌失败",
		})
		return
	}

	c.JSON(http.StatusOK, types.LoginResponse{
		Code:         200,
		Message:      "success",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// @Summary		Register
// @Description	Register
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			req	body		types.RegisterRequest	true	"Register Request"
// @Success		200	{object}	types.LoginResponse
// @Failure		400	{object}	types.ErrorResponse
// @Router			/auth/register [post]
func RegisterHandler(c *gin.Context) {
	var req types.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "参数错误",
		})
		return
	}
	s, err := redis.GetSession(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "获取session失败",
		})
		return
	}
	code, err := s.Get("code")
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "获取验证码失败",
		})
		return
	}
	if code != req.Code {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "验证码错误",
		})
		return
	}
	// 如果参数错误，返回 400 错误
	if !utils.Lengthcheck(6, req.Username, req.Password) {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "用户名或密码长度不足",
		})
		return
	} // 如果用户名或密码长度不足，返回 400 错误
	password, err := utils.HashPassword(req.Password) // 加密密码
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "密码加密失败",
		})
		return
	} // 如果密码加密失败，返回 500 错误
	userID, err := sql.NewUserMapper().Insert(&model.User{
		Username: req.Username,
		Password: password,
	}) // 如果注册失败，返回 500 错误
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "注册失败",
		})
		return
	}
	accessToken, err := utils.GenerateAccessToken(int(userID), req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "生成访问令牌失败",
		})
		return
	}
	// 如果生成访问令牌失败，返回 500 错误

	refreshToken, refreshClaims, err := utils.GenerateRefreshToken(int(userID), req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "生成刷新令牌失败",
		})
		return
	}
	// 如果生成刷新令牌失败，返回 500 错误
	ttl := time.Until(refreshClaims.ExpiresAt.Time)
	if err = redis.SaveRefreshToken(refreshClaims.ID, int(userID), ttl); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "保存刷新令牌失败",
		})
		return
	} // 保存刷新令牌
	// 如果保存刷新令牌失败，返回 500 错误
	c.JSON(http.StatusOK, types.LoginResponse{
		Code:         200,
		Message:      "success",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}) // 如果注册成功，返回 200 状态码

}

//	@Summary		RefreshToken
//	@Description	使用 refresh_token 刷新访问令牌
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			req	body		types.RefreshTokenRequest	true	"Refresh Token Request"
//	@Success		200	{object}	types.LoginResponse
//	@Failure		400	{object}	types.ErrorResponse
//	@Failure		401	{object}	types.ErrorResponse
//	@Router			/auth/refresh [post]
//
// 刷新访问令牌 生成新的访问令牌和刷新令牌
func RefreshTokenHandler(c *gin.Context) {
	var req types.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "参数错误",
		})
		return
	}
	// 验证刷新令牌
	oldClaims, err := utils.ValidateRefreshToken(req.RefreshToken) // 验证刷新令牌
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "刷新令牌无效",
		})
		return
	}
	// 如果刷新令牌无效，返回 401 错误
	// 验证刷新令牌是否有效
	ok, err := redis.VerifyRefreshToken(oldClaims.ID)
	if err != nil || !ok {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "刷新令牌已失效",
		})
		return
	} // 如果刷新令牌无效，返回 401 错误

	_ = redis.DeleteRefreshToken(oldClaims.ID) // 删除刷新令牌

	accessToken, err := utils.GenerateAccessToken(oldClaims.UserID, oldClaims.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "生成访问令牌失败",
		})
		return
	}

	refreshToken, newRefreshClaims, err := utils.GenerateRefreshToken(oldClaims.UserID, oldClaims.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "生成刷新令牌失败",
		})
		return
	}
	ttl := time.Until(newRefreshClaims.ExpiresAt.Time)
	if err = redis.SaveRefreshToken(newRefreshClaims.ID, oldClaims.UserID, ttl); err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "保存刷新令牌失败",
		})
		return
	}

	c.JSON(http.StatusOK, types.LoginResponse{
		Code:         200,
		Message:      "success",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

//	@Summary		JudgeToken
//	@Description	JudgeToken
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	types.SuccessResponse
//	@Failure		401	{object}	types.ErrorResponse
//	@Router			/auth/judgeToken [get]
//
// 判断 token 是否有效
func JudgeToken(ctx *gin.Context) {
	token, exists := ctx.Get("token")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "unauthorized",
		})
		return
	}
	Log.ZLog.Info("JudgeToken", zap.String("token", token.(string))) // 记录token日志
	ctx.JSON(http.StatusOK, types.SuccessResponse{
		Code:    http.StatusOK,
		Message: "token valid",
	}) // token 有效时返回 JSON，避免前端 fetch 跟随 302 导致跨域
}
