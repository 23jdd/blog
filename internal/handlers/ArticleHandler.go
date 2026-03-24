package handlers

import (
	"blog/internal/model"
	"blog/internal/redis"
	"blog/internal/sql"
	"blog/internal/types"
	"blog/internal/utils"
	stdsql "database/sql"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func normalizeArticleContentURL(v1, v2 string) string {
	v := strings.TrimSpace(v1)
	if v == "" {
		v = strings.TrimSpace(v2)
	}
	return v
}

func isValidArticleContentURL(raw string) bool {
	if strings.HasPrefix(raw, "/") {
		return true
	}
	u, err := url.ParseRequestURI(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

// @Summary		创建文章
// @Description	创建新的文章
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			article	body		types.CreateArticleRequest	true	"文章信息"
// @Success		200		{object}	types.CreateArticleResponse
// @Failure		400		{object}	types.ErrorResponse
// @Failure		500		{object}	types.ErrorResponse
// @Router			/articles [post]
func CreateArticle(ctx *gin.Context) {
	var req types.CreateArticleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "参数错误",
		})
		return
	}

	if req.Status == "" {
		req.Status = "published"
	}
	if req.Status != "draft" && req.Status != "published" && req.Status != "offline" {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "状态非法"})
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未授权"})
		return
	}
	authorID, ok := userID.(int)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未授权"})
		return
	}

	title := strings.TrimSpace(req.Title)
	contentURL := normalizeArticleContentURL(req.ContentURL, req.Content)

	// 支持「发布时选择草稿」：传 draft_id 时从草稿带出标题与内容URL
	if req.DraftID > 0 {
		draft, derr := sql.NewDraftMapper().FindByID(req.DraftID)
		if derr != nil {
			if errors.Is(derr, stdsql.ErrNoRows) {
				ctx.JSON(http.StatusNotFound, types.ErrorResponse{Message: "草稿不存在"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "查询草稿失败"})
			return
		}
		if draft.AuthorID != authorID {
			ctx.JSON(http.StatusForbidden, types.ErrorResponse{Message: "无权限发布他人草稿"})
			return
		}
		if title == "" {
			title = strings.TrimSpace(draft.Title)
		}
		if contentURL == "" {
			contentURL = strings.TrimSpace(draft.Content)
		}
	}
	if title == "" || contentURL == "" {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "标题和content_url不能为空"})
		return
	}
	if !isValidArticleContentURL(contentURL) {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "content_url 非法，必须是 / 开头或 http/https URL"})
		return
	}

	article := &model.Article{
		Title:      title,
		Content:    contentURL,
		AuthorID:   authorID,
		Status:     req.Status,
		CategoryID: req.CategoryID,
		Tags:       req.Tags,
		CoverURL:   req.CoverURL,
	}

	articleMapper := sql.NewArticleMapper()
	id, err := articleMapper.Insert(article)
	if err == nil {
		feedService := redis.NewFeedService(redis.Client)
		_ = feedService.AddArticleToFeed(authorID, int(id), time.Now())
		ctx.JSON(http.StatusOK, types.CreateArticleResponse{ID: id})
		return
	}
	ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
		Message: "创建文章失败",
	})

}

// @Summary		获取文章详情
// @Description	根据文章ID获取文章详情
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			id	path		int	true	"文章ID"
// @Success		200	{object}	model.Article
// @Failure		400	{object}	types.ErrorResponse
// @Failure		500	{object}	types.ErrorResponse
// @Router			/articles/{id} [get]
func GetArticleByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "参数错误",
		})
		return
	}
	articleMapper := sql.NewArticleMapper()
	if article, err := articleMapper.FindByID(id); err == nil {
		_, _ = redis.IncArticleView(id)
		_ = redis.IncHotByRead(id)
		ctx.JSON(http.StatusOK, article)
		return
	}
	ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
		Message: "文章查询失败",
	})
}

// @Summary		更新文章
// @Description	更新文章基础信息
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			id		path		int							true	"文章ID"
// @Param			article	body		types.UpdateArticleRequest	true	"更新信息"
// @Success		200		{object}	types.SuccessResponse
// @Failure		400		{object}	types.ErrorResponse
// @Failure		500		{object}	types.ErrorResponse
// @Router			/articles/{id} [put]
func UpdateArticle(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}

	var req types.UpdateArticleRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	title := strings.TrimSpace(req.Title)
	contentURL := normalizeArticleContentURL(req.ContentURL, req.Content)
	if title == "" || contentURL == "" {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "标题和content_url不能为空"})
		return
	}
	if !isValidArticleContentURL(contentURL) {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "content_url 非法，必须是 / 开头或 http/https URL"})
		return
	}

	err = sql.NewArticleMapper().UpdateByID(id, &model.Article{
		Title:      title,
		Content:    contentURL,
		CategoryID: req.CategoryID,
		Tags:       req.Tags,
		CoverURL:   req.CoverURL,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "更新文章失败"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "更新文章成功"})
}

// @Summary		删除文章
// @Description	删除文章
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			id	path		int	true	"文章ID"
// @Success		200	{object}	types.SuccessResponse
// @Failure		400	{object}	types.ErrorResponse
// @Failure		500	{object}	types.ErrorResponse
// @Router			/articles/{id} [delete]
func DeleteArticle(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	if err = sql.NewArticleMapper().DeleteByID(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "删除文章失败"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "删除文章成功"})
}

// @Summary		更新文章状态
// @Description	设置文章状态（草稿/发布/下架）
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			id	path		int									true	"文章ID"
// @Param			req	body		types.UpdateArticleStatusRequest	true	"状态信息"
// @Success		200	{object}	types.SuccessResponse
// @Failure		400	{object}	types.ErrorResponse
// @Failure		500	{object}	types.ErrorResponse
// @Router			/articles/{id}/status [patch]
func UpdateArticleStatus(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	var req types.UpdateArticleStatusRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	if req.Status != "draft" && req.Status != "published" && req.Status != "offline" {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "状态非法"})
		return
	}
	if err = sql.NewArticleMapper().UpdateStatusByID(id, req.Status); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "更新文章状态失败"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "更新文章状态成功"})
}

// @Summary		按作者查询文章
// @Description	按作者ID查询文章列表
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			authid	path		int	true	"作者ID"
// @Success		200		{array}		model.Article
// @Failure		400		{object}	types.ErrorResponse
// @Failure		500		{object}	types.ErrorResponse
// @Router			/articles/author/{authid} [get]
func GetArticlesByAuthorID(ctx *gin.Context) {
	authid, err := strconv.Atoi(ctx.Param("authid"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{
			Message: "参数错误",
		})
		return
	}
	articleMapper := sql.NewArticleMapper()
	if articles, err := articleMapper.FindByAuthorID(authid); err == nil {
		ctx.JSON(http.StatusOK, articles)
		return
	}
	ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{
		Message: "文章查询失败",
	})
}

// @Summary		按标签查询文章
// @Description	按标签查询文章列表（支持分页）
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			tag			query		string	true	"标签"
// @Param			page		query		int		false	"页码"
// @Param			pageSize	query		int		false	"每页条数"
// @Success		200			{array}		model.Article
// @Failure		400			{object}	types.ErrorResponse
// @Failure		500			{object}	types.ErrorResponse
// @Router			/articles/by-tag [get]
func GetArticlesByTag(ctx *gin.Context) {
	tag := strings.TrimSpace(ctx.Query("tag"))
	if tag == "" {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "标签不能为空"})
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	offset, limit := utils.ResolveOffsetLimit(page, pageSize)

	articleMapper := sql.NewArticleMapper()
	if articles, err := articleMapper.FindByTag(tag, limit, offset); err == nil {
		ctx.JSON(http.StatusOK, articles)
		return
	}
	ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "文章查询失败"})
}

// @Summary		按分类查询文章
// @Description	按分类ID查询文章列表（支持分页）
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			categoryID	path		int	true	"分类ID"
// @Param			page		query		int	false	"页码"
// @Param			pageSize	query		int	false	"每页条数"
// @Success		200			{array}		model.Article
// @Failure		400			{object}	types.ErrorResponse
// @Failure		500			{object}	types.ErrorResponse
// @Router			/articles/category/{categoryID} [get]
func GetArticlesByCategory(ctx *gin.Context) {
	categoryID, err := strconv.Atoi(ctx.Param("categoryID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	offset, limit := utils.ResolveOffsetLimit(page, pageSize)

	articleMapper := sql.NewArticleMapper()
	if articles, err := articleMapper.FindByCategoryID(categoryID, limit, offset); err == nil {
		ctx.JSON(http.StatusOK, articles)
		return
	}
	ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "文章查询失败"})
}

// @Summary		搜索文章
// @Description	搜索文章列表（支持分页）
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			keyword		query		string	true	"关键词"
// @Param			page		query		int		false	"页码"
// @Param			pageSize	query		int		false	"每页条数"
// @Success		200			{array}		model.Article
// @Failure		400			{object}	types.ErrorResponse
// @Failure		500			{object}	types.ErrorResponse
// @Router			/articles/search [get]
func SearchArticles(ctx *gin.Context) {
	keyword := strings.TrimSpace(ctx.Query("keyword"))
	if keyword == "" {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "关键词不能为空"})
		return
	}

}
