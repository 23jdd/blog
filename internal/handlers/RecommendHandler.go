package handlers

import (
	"blog/internal/redis"
	"blog/internal/sql"
	"blog/internal/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetMyFeed 获取当前用户个性化时间线
//
//	@Summary		获取个性化Feed
//	@Description	分页获取当前用户的Feed文章ID列表
//	@Tags			recommend
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int	false	"页码"
//	@Param			pageSize	query		int	false	"每页条数"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		401			{object}	types.ErrorResponse
//	@Failure		500			{object}	types.ErrorResponse
//	@Router			/interactions/feed [get]
func GetMyFeed(ctx *gin.Context) {
	userID, ok := getUserID(ctx) // 获取用户ID
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未授权"})
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1")) // 获取页码，如果为空，则默认为 1
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10")) // 获取每页条数，如果为空，则默认为 10
	feedService := redis.NewFeedService(redis.Client) // 创建FeedService
	ids, total, err := feedService.GetFeedWithPagination(userID, page, pageSize) // 获取Feed文章ID列表
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "获取Feed失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"article_ids": ids, // 返回Feed文章ID列表
		"total":       total, // 返回总条数
		"page":        page, // 返回页码
		"page_size":   pageSize, // 返回每页条数
	})
}

// GetHotArticles 获取热门文章榜单
//
//	@Summary		获取热门文章榜单
//	@Description	按热度分数返回热门文章ID列表
//	@Tags			recommend
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int	false	"返回条数"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		500		{object}	types.ErrorResponse
//	@Router			/articles/hot [get]
func GetHotArticles(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10")) // 获取 limit 参数，如果为空，则默认为 10
	ids, err := redis.GetHot(limit) // 获取热门文章ID列表
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "获取热榜失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"article_ids": ids}) // 返回热门文章ID列表
}

// GetArticleStats 获取文章统计信息
//
//	@Summary		获取文章统计信息
//	@Description	返回文章访问量、点赞数、评论数
//	@Tags			recommend
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"文章ID"
//	@Success		200	{object}	map[string]interface{}
//	@Failure		400	{object}	types.ErrorResponse
//	@Router			/articles/{id}/stats [get]
func GetArticleStats(ctx *gin.Context) {
	articleID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	viewCount, err := redis.GetArticleView(articleID)
	if err != nil {
		viewCount = 0
	}
	likeCount, _ := sql.NewLikeMapper().CountByArticle(articleID) // 获取文章点赞数
	commentCount, _ := sql.NewReviewMapper(sql.GetDB()).CountByArticleID(articleID) // 获取文章评论数
	ctx.JSON(http.StatusOK, gin.H{
		"article_id":    articleID,
		"view_count":    viewCount, // 返回文章访问量
		"like_count":    likeCount, // 返回文章点赞数
		"comment_count": commentCount, // 返回文章评论数
	})
}
