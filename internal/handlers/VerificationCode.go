package handlers

import (
	send "blog/internal/Email"
	"blog/internal/redis"
	"blog/internal/types"
	"blog/internal/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary		SendEmailVerifyCode
// @Description	SendEmailVerifyCode
// @Tags			verification
// @Accept			json
// @Produce		json
// @Param			email	query		string	true	"Email"
// @Param			code	query		string	true	"Code"
// @Success		200		{object}	types.SuccessResponse
// @Failure		400		{object}	types.ErrorResponse
// @Failure		500		{object}	types.ErrorResponse
// @Router			/verification/send [get]
func SendVerificationCode(ctx *gin.Context) {
	email := ctx.Query("email") // 获取邮箱
	if !utils.EmailCheck(email) {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "邮箱格式不正确",
		})
		return
	} // 如果邮箱格式不正确，返回 400 错误
	code := utils.GenerateRandomCode(6)
	err := send.SendPlain([]string{email}, "验证码", fmt.Sprintf("您的验证码是：%s", code))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "发送验证码失败",
		})
		return
	} // 如果发送验证码失败，返回 500 错误
	s := redis.NewSession(email)
	s.Set("code", code)
	s.Save(5 * time.Minute)
	ctx.JSON(http.StatusOK, types.SuccessResponse{
		Message: "发送验证码成功",
	})
}
