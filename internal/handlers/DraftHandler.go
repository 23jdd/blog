package handlers

import (
	stdsql "database/sql"
	"errors"
	"net/url"

	"blog/internal/model"
	"blog/internal/sql"
	"blog/internal/types"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func normalizeContentURL(v1, v2 string) string {
	v := strings.TrimSpace(v1)
	if v == "" {
		v = strings.TrimSpace(v2)
	}
	return v
}

func isValidContentURL(raw string) bool {
	if strings.HasPrefix(raw, "/") {
		return true
	}
	u, err := url.ParseRequestURI(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

func getAuthUserID(ctx *gin.Context) (int, bool) {
	userID, ok := ctx.Get("userID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未授权"})
		return 0, false
	}
	authorID, ok := userID.(int)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未授权"})
		return 0, false
	}
	return authorID, true
}

// SaveDraft 保存草稿
//
//	@Summary		保存草稿
//	@Description	保存当前登录用户的草稿
//	@Tags			drafts
//	@Accept			json
//	@Produce		json
//	@Param			draft	body		types.DraftSaveRequest	true	"草稿内容"
//	@Success		200		{object}	types.SuccessResponse
//	@Failure		400		{object}	types.ErrorResponse
//	@Failure		401		{object}	types.ErrorResponse
//	@Failure		500		{object}	types.ErrorResponse
//	@Router			/drafts [post]
func SaveDraft(ctx *gin.Context) {
	var req types.DraftSaveRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	contentURL := strings.TrimSpace(req.ContentURL)
	contentText := strings.TrimSpace(req.Content)
	if req.Title == "" || (contentURL == "" && contentText == "") {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "标题和内容不能为空（content_url 或 content 至少一个）"})
		return
	}
	if len([]rune(req.Title)) > 255 {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "标题长度不能超过255个字符"})
		return
	}
	authorID, ok := getAuthUserID(ctx)
	if !ok {
		return
	}
	if contentURL == "" {
		var err error
		contentURL, err = saveMarkdownContentForUser(authorID, contentText)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "保存内容文件失败"})
			return
		}
	}
	contentURL = normalizeContentURL(contentURL, "")
	if !isValidContentURL(contentURL) {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "content_url 非法，必须是 / 开头或 http/https URL"})
		return
	}

	mapper := sql.NewDraftMapper()
	id, err := mapper.Create(&model.Draft{
		Title:    req.Title,
		Content:  contentURL,
		AuthorID: authorID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "保存草稿失败"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{
		Message: "保存草稿成功",
		Data:    types.DraftIDResponse{ID: id},
	})
}

// UpdateDraft 更新草稿
//
//	@Summary		更新草稿
//	@Description	更新当前登录用户自己的草稿
//	@Tags			drafts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"草稿ID"
//	@Param			draft	body		types.DraftSaveRequest	true	"草稿内容"
//	@Success		200		{object}	types.SuccessResponse
//	@Failure		400		{object}	types.ErrorResponse
//	@Failure		401		{object}	types.ErrorResponse
//	@Failure		403		{object}	types.ErrorResponse
//	@Failure		404		{object}	types.ErrorResponse
//	@Failure		500		{object}	types.ErrorResponse
//	@Router			/drafts/{id} [put]
func UpdateDraft(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	var req types.DraftSaveRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	contentURL := strings.TrimSpace(req.ContentURL)
	contentText := strings.TrimSpace(req.Content)
	if req.Title == "" || (contentURL == "" && contentText == "") {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "标题和内容不能为空（content_url 或 content 至少一个）"})
		return
	}
	if len([]rune(req.Title)) > 255 {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "标题长度不能超过255个字符"})
		return
	}
	authorID, ok := getAuthUserID(ctx)
	if !ok {
		return
	}
	if contentURL == "" {
		contentURL, err = saveMarkdownContentForUser(authorID, contentText)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "保存内容文件失败"})
			return
		}
	}
	contentURL = normalizeContentURL(contentURL, "")
	if !isValidContentURL(contentURL) {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "content_url 非法，必须是 / 开头或 http/https URL"})
		return
	}
	mapper := sql.NewDraftMapper()

	exists, err := mapper.FindByID(id)
	if err != nil {
		if errors.Is(err, stdsql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, types.ErrorResponse{Message: "草稿不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "查询草稿失败"})
		return
	}
	if exists.AuthorID != authorID {
		ctx.JSON(http.StatusForbidden, types.ErrorResponse{Message: "无权限操作他人草稿"})
		return
	}

	if err = mapper.Update(&model.Draft{
		ID:       id,
		Title:    req.Title,
		Content:  contentURL,
		AuthorID: authorID,
	}); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "更新草稿失败"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "更新草稿成功"})
}

// DeleteDraft 删除草稿
//
//	@Summary		删除草稿
//	@Description	删除当前登录用户自己的草稿
//	@Tags			drafts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"草稿ID"
//	@Success		200	{object}	types.SuccessResponse
//	@Failure		400	{object}	types.ErrorResponse
//	@Failure		401	{object}	types.ErrorResponse
//	@Failure		403	{object}	types.ErrorResponse
//	@Failure		404	{object}	types.ErrorResponse
//	@Failure		500	{object}	types.ErrorResponse
//	@Router			/drafts/{id} [delete]
func DeleteDraft(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id <= 0 {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	authorID, ok := getAuthUserID(ctx)
	if !ok {
		return
	}
	mapper := sql.NewDraftMapper()

	exists, err := mapper.FindByID(id)
	if err != nil {
		if errors.Is(err, stdsql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, types.ErrorResponse{Message: "草稿不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "查询草稿失败"})
		return
	}
	if exists.AuthorID != authorID {
		ctx.JSON(http.StatusForbidden, types.ErrorResponse{Message: "无权限操作他人草稿"})
		return
	}

	if err = mapper.Delete(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "删除草稿失败"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "删除草稿成功"})
}

// ListDrafts 获取当前用户草稿列表
//
//	@Summary		草稿列表
//	@Description	获取当前登录用户的草稿列表
//	@Tags			drafts
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	types.SuccessResponse
//	@Failure		401	{object}	types.ErrorResponse
//	@Failure		500	{object}	types.ErrorResponse
//	@Router			/drafts [get]
func ListDrafts(ctx *gin.Context) {
	authorID, ok := getAuthUserID(ctx)
	if !ok {
		return
	}
	mapper := sql.NewDraftMapper()
	list, err := mapper.FindAll(authorID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "查询草稿失败"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{
		Message: "查询草稿成功",
		Data:    types.DraftListResponse{List: list},
	})
}
