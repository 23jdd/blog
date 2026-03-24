package handlers

import (
	"blog/internal/email"
	"blog/internal/redis"
	"blog/internal/types"
	"blog/internal/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const emailVerifySessionPrefix = "email_verify:"

func emailVerifySessionKey(addr string) string {
	return emailVerifySessionPrefix + addr
}

// @Summary		EmailVerifyHandler
// @Description	EmailVerifyHandler
// @Tags			email
// @Accept			json
// @Produce		json
// @Param			email	query		string	true	"??"
// @Success		200		{object}	types.SuccessResponse
// @Failure		400		{object}	types.ErrorResponse
// @Failure		500		{object}	types.ErrorResponse
// @Router			/email/verify [get]
func EmailVerifyHandler(ctx *gin.Context) {
	m, err := SendEmailVerifyCode(ctx.Query("email"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "???????",
		})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{
		Message: "???????",
		Data:    m,
	})
}
func SendEmailVerifyCode(addr string) (string, error) {
	code := utils.GenerateRandomCode(6)
	session := redis.NewSession(emailVerifySessionKey(addr))
	session.Map["code"] = code
	if err := session.Save(10 * time.Minute); err != nil {
		return "", err
	}
	subject := "???"
	body := fmt.Sprintf("???????%s?10 ??????", code)
	if err := email.SendPlain([]string{addr}, subject, body); err != nil {
		return "", err
	}
	return addr, nil
}

func VerifyEmailVerifyCode(email, code string) (bool, error) {
	session, err := redis.GetSession(emailVerifySessionKey(email))
	if err != nil {
		return false, err
	}
	codeVal, err := session.Get("code") // ?????
	if err != nil {
		return false, err
	}
	if codeVal != code { // ??????
		return false, nil
	}
	return true, nil
}
