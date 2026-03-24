package handlers

import (
	"blog/internal/types"
	"blog/internal/utils"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary		MarkdownHandler
// @Description	MarkdownHandler
// @Tags			markdown
// @Accept			text/html
// @Produce		json
// @Param			markdown	body		string	true	"Markdown"
// @Success		200			{object}	types.SuccessResponse
// @Failure		400			{object}	types.ErrorResponse
// @Router			/markdown [post]
func MarkdownHandler(ctx *gin.Context) {
	markdown, err := io.ReadAll(ctx.Request.Body) // 读取 markdown 内容
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "参数错误",
		})
		return
	}
	html := utils.MarkdownToHtml(string(markdown)) // 将 markdown 转换为 html
	ctx.JSON(http.StatusOK, gin.H{
		"html": html,
	}) // 返回 html 内容
}
