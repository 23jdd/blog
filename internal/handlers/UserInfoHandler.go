package handlers

import (
	"blog/internal/model"
	"blog/internal/sql"
	"blog/internal/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary		GetUserInfo
// @Description	GetUserInfo
// @Tags			auth
// @Accept			json
// @Produce		json
// @Success		200	{object}	types.UserInfoResponse
// @Failure		401	{object}	types.ErrorResponse
// @Failure		500	{object}	types.ErrorResponse
// @Router			/user/info [get]
func GetUserInfoHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("userID") //  通过 AuthMiddleware 获取用户ID
	if !exists {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "unauthorized",
		})
		return
	}
	id, ok := userID.(int) // 将用户ID转换为 int 类型
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "unauthorized",
		})
		return
	}
	m := sql.NewUserMapper()
	user, err := m.GetUserInfoByID(id) // 通过用户ID获取用户信息
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "Get user info failed",
		})
		return
	}
	ctx.JSON(http.StatusOK, user) // 返回用户信息
}

// @Summary		UpdateUserInfo
// @Description	UpdateUserInfo
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			age		formData	string	true	"Age"
// @Param			gender	formData	string	true	"Gender"
// @Success		200		{object}	types.SuccessResponse
// @Failure		401		{object}	types.ErrorResponse
// @Failure		500		{object}	types.ErrorResponse
// @Router			/user/update [post]
func UpdateUserInfoHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("userID") // 通过 AuthMiddleware 获取用户ID
	if !exists {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "unauthorized",
		})
		return
	}
	//  age=18&gender=male
	age := ctx.PostForm("age")                                  // 获取年龄
	gender := ctx.PostForm("gender")                            // 获取性别
	ageVal, genderVal := sql.BuildNullableUserInfo(age, gender) // 构建可空用户信息
	if ageVal.Valid == false || genderVal.Valid == false {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "age or gender is required",
		})
		return
	} // 如果年龄或性别为空，返回 400 错误
	u := model.User{
		Age:    ageVal,
		Gender: genderVal,
	}
	id, ok := userID.(int)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "unauthorized",
		})
		return
	}
	m := sql.NewUserMapper()
	// 更新用户信息如果失败，返回 500 错误
	err := m.UpdateUserInfo(id, u)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "Update user info failed",
		})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{
		Message: "Update user info success",
	}) // 返回更新用户信息成功
}

// @Summary		DeleteUser
// @Description	DeleteUser
// @Tags			auth
// @Accept			json
// @Produce		json
// @Success		200	{object}	types.SuccessResponse
// @Failure		401	{object}	types.ErrorResponse
// @Failure		500	{object}	types.ErrorResponse
// @Router			/user/delete [delete]
func DeleteUserHandler(ctx *gin.Context) {
	userID, exists := ctx.Get("userID") // 通过 AuthMiddleware 获取用户ID
	if !exists {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "unauthorized",
		})
		return
	}
	exp, exists := ctx.Get("exp")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "unauthorized",
		})
		return
	}
	id, ok := userID.(int)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "unauthorized",
		})
		return
	}
	_, ok = exp.(time.Time)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Message: "unauthorized",
		})
		return
	}
	err := sql.NewUserMapper().Delete(id) // 删除用户信息如果失败，返回 500 错误
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Message: "Delete user failed",
		})
		return
	}

	ctx.JSON(http.StatusOK, types.SuccessResponse{
		Message: "Delete user success",
	})
	// TODO  将 access token jti 加入黑名单
}
