package handlers

import (
	"blog/internal/etcd"
	"blog/internal/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type DynamicConfigRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// UpdateDynamicConfig 动态更新配置并写入etcd
//
//	@Summary		更新动态配置
//	@Description	将配置写入 etcd 并同步到当前服务进程
//	@Tags			config
//	@Accept			json
//	@Produce		json
//	@Param			req	body		DynamicConfigRequest	true	"动态配置"
//	@Success		200	{object}	types.SuccessResponse
//	@Failure		400	{object}	types.ErrorResponse
//	@Failure		500	{object}	types.ErrorResponse
//	@Router			/config [post]
func UpdateDynamicConfig(ctx *gin.Context) {
	var req DynamicConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Key == "" {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}

	cli := etcd.NewEtcdClient()
	if err := cli.Put(req.Key, req.Value); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "写入配置失败"})
		return
	}
	viper.Set(req.Key, req.Value)
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "配置更新成功"})
}

// GetDynamicConfig 查询动态配置
//
//	@Summary		查询动态配置
//	@Description	按 key 查询 etcd 中的配置值
//	@Tags			config
//	@Accept			json
//	@Produce		json
//	@Param			key	query		string	true	"配置键"
//	@Success		200	{object}	map[string]string
//	@Failure		400	{object}	types.ErrorResponse
//	@Failure		500	{object}	types.ErrorResponse
//	@Router			/config [get]
func GetDynamicConfig(ctx *gin.Context) {
	key := ctx.Query("key")
	if key == "" {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	cli := etcd.NewEtcdClient()
	val, err := cli.Get(key)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "读取配置失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"key": key, "value": val})
}
