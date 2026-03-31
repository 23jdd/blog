package handlers

import (
	"blog/internal/model"
	"blog/internal/redis"
	"blog/internal/sql"
	"blog/internal/types"
	"blog/internal/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func getUserID(ctx *gin.Context) (int, bool) { // 获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		return 0, false
	}
	id, ok := userID.(int) // 将用户ID转换为 int 类型
	return id, ok          // 返回用户ID
}

// CreateComment 发布评论/回复
//
//	@Summary		发布评论
//	@Description	对文章发布评论，parent_id=0 表示一级评论，非0表示回复评论
//	@Tags			interactions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int							true	"文章ID"
//	@Param			req	body		types.CreateCommentRequest	true	"评论内容"
//	@Success		200	{object}	map[string]interface{}
//	@Failure		400	{object}	types.ErrorResponse
//	@Failure		401	{object}	types.ErrorResponse
//	@Failure		500	{object}	types.ErrorResponse
//	@Router			/articles/{id}/comments [post]
func CreateComment(ctx *gin.Context) {
	articleID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	userID, ok := getUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未授权"})
		return
	}
	var req types.CreateCommentRequest
	if err = ctx.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Content) == "" {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}

	review := &model.Review{
		ArticleID: articleID,
		Content:   req.Content,
		AuthorID:  userID,
		IsDirect:  req.ParentID == 0,
		ParentID:  req.ParentID,
		Status:    "pending",
	}
	mapper := sql.NewReviewMapper(sql.GetDB())
	id, err := mapper.Create(review)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "创建评论失败"})
		return
	}
	_ = redis.IncHotByComment(articleID)
	ctx.JSON(http.StatusOK, gin.H{"id": id, "message": "创建评论成功"})
}

// DeleteComment 删除评论（仅作者）
//
//	@Summary		删除评论
//	@Description	删除当前用户自己的评论
//	@Tags			interactions
//	@Accept			json
//	@Produce		json
//	@Param			commentID	path		int	true	"评论ID"
//	@Success		200			{object}	types.SuccessResponse
//	@Failure		400			{object}	types.ErrorResponse
//	@Failure		401			{object}	types.ErrorResponse
//	@Failure		500			{object}	types.ErrorResponse
//	@Router			/articles/comments/{commentID} [delete]
func DeleteComment(ctx *gin.Context) {
	commentID, err := strconv.Atoi(ctx.Param("commentID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	userID, ok := getUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未授权"})
		return
	}
	mapper := sql.NewReviewMapper(sql.GetDB())
	if err = mapper.DeleteByIDAndAuthor(commentID, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "删除评论失败"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "删除评论成功"})
}

// AuditComment 评论审核
//
//	@Summary		审核评论
//	@Description	审核评论状态（pending/approved/rejected）
//	@Tags			interactions
//	@Accept			json
//	@Produce		json
//	@Param			commentID	path		int							true	"评论ID"
//	@Param			req			body		types.ReviewStatusRequest	true	"审核状态"
//	@Success		200			{object}	types.SuccessResponse
//	@Failure		400			{object}	types.ErrorResponse
//	@Failure		500			{object}	types.ErrorResponse
//	@Router			/articles/comments/{commentID}/status [patch]
func AuditComment(ctx *gin.Context) {
	commentID, err := strconv.Atoi(ctx.Param("commentID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	var req types.ReviewStatusRequest
	if err = ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	if req.Status != "pending" && req.Status != "approved" && req.Status != "rejected" {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "状态非法"})
		return
	}
	mapper := sql.NewReviewMapper(sql.GetDB())
	if err = mapper.UpdateStatus(commentID, req.Status); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "审核评论失败"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "审核评论成功"})
}

// ListArticleComments 获取文章评论
//
//	@Summary		获取文章评论列表
//	@Description	分页获取文章评论
//	@Tags			interactions
//	@Accept			json
//	@Produce		json
//	@Param			id			path		int	true	"文章ID"
//	@Param			page		query		int	false	"页码"
//	@Param			pageSize	query		int	false	"每页条数"
//	@Success		200			{array}		model.Review
//	@Failure		400			{object}	types.ErrorResponse
//	@Failure		500			{object}	types.ErrorResponse
//	@Router			/articles/{id}/comments [get]
func ListArticleComments(ctx *gin.Context) {
	articleID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	offset, limit := utils.ResolveOffsetLimit(page, pageSize)

	mapper := sql.NewReviewMapper(sql.GetDB())
	rows, err := mapper.FindByArticleIDWithPagination(articleID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "查询评论失败"})
		return
	}
	ctx.JSON(http.StatusOK, rows)
}

// LikeArticle 点赞文章（防重复）
//
//	@Summary		点赞文章
//	@Description	对文章点赞，重复点赞直接返回已点赞状态
//	@Tags			interactions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"文章ID"
//	@Success		200	{object}	types.LikeActionResponse
//	@Failure		400	{object}	types.ErrorResponse
//	@Failure		401	{object}	types.ErrorResponse
//	@Failure		500	{object}	types.ErrorResponse
//	@Router			/articles/{id}/likes [post]
func LikeArticle(ctx *gin.Context) {
	articleID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	userID, ok := getUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未授权"})
		return
	}
	mapper := sql.NewLikeMapper()
	exist, err := mapper.Exists(userID, articleID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "点赞失败"})
		return
	}
	if exist {
		ctx.JSON(http.StatusOK, types.LikeActionResponse{Liked: true})
		return
	}
	_, err = mapper.Create(&model.Like{UserID: userID, ArticleID: articleID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "点赞失败"})
		return
	}
	_ = redis.IncHotByLike(articleID)
	ctx.JSON(http.StatusOK, types.LikeActionResponse{Liked: true})
}

// UnlikeArticle 取消点赞
//
//	@Summary		取消点赞
//	@Description	取消当前用户对文章的点赞
//	@Tags			interactions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"文章ID"
//	@Success		200	{object}	types.LikeActionResponse
//	@Failure		400	{object}	types.ErrorResponse
//	@Failure		401	{object}	types.ErrorResponse
//	@Failure		500	{object}	types.ErrorResponse
//	@Router			/articles/{id}/likes [delete]
func UnlikeArticle(ctx *gin.Context) {
	articleID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	userID, ok := getUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未授权"})
		return
	}
	if err = sql.NewLikeMapper().Delete(userID, articleID); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "取消点赞失败"})
		return
	}
	_ = redis.DecHotByLike(articleID)
	ctx.JSON(http.StatusOK, types.LikeActionResponse{Liked: false})
}

// CollectArticle 收藏文章（防重复）
//
//	@Summary		收藏文章
//	@Description	收藏文章，重复收藏返回已收藏
//	@Tags			interactions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"文章ID"
//	@Success		200	{object}	types.SuccessResponse
//	@Failure		400	{object}	types.ErrorResponse
//	@Failure		401	{object}	types.ErrorResponse
//	@Failure		404	{object}	types.ErrorResponse
//	@Failure		500	{object}	types.ErrorResponse
//	@Router			/articles/{id}/collections [post]
func CollectArticle(ctx *gin.Context) {
	articleID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	userID, ok := getUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未授权"})
		return
	}

	article, err := sql.NewArticleMapper().FindByID(articleID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, types.ErrorResponse{Message: "文章不存在"})
		return
	}

	mapper := sql.NewCollectMapperDefault()
	exist, err := mapper.Exists(userID, articleID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "收藏失败"})
		return
	}
	if exist {
		ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "已收藏"})
		return
	}
	_, err = mapper.Create(&model.Collect{
		UserID:       userID,
		ArticleID:    articleID,
		ArticleTitle: article.Title,
		AuthorID:     article.AuthorID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "收藏失败"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "收藏成功"})
}

// UnCollectArticle 取消收藏
//
//	@Summary		取消收藏
//	@Description	取消当前用户对文章的收藏
//	@Tags			interactions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"文章ID"
//	@Success		200	{object}	types.SuccessResponse
//	@Failure		400	{object}	types.ErrorResponse
//	@Failure		401	{object}	types.ErrorResponse
//	@Failure		500	{object}	types.ErrorResponse
//	@Router			/articles/{id}/collections [delete]
func UnCollectArticle(ctx *gin.Context) {
	articleID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	userID, ok := getUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未授权"})
		return
	}
	if err = sql.NewCollectMapperDefault().Delete(userID, articleID); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "取消收藏失败"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "取消收藏成功"})
}

// ListMyCollections 查询用户收藏列表
//
//	@Summary		获取我的收藏列表
//	@Description	分页查询当前用户收藏的文章
//	@Tags			interactions
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int	false	"页码"
//	@Param			pageSize	query		int	false	"每页条数"
//	@Success		200			{array}		model.Collect
//	@Failure		401			{object}	types.ErrorResponse
//	@Failure		500			{object}	types.ErrorResponse
//	@Router			/interactions/my-collections [get]
func ListMyCollections(ctx *gin.Context) {
	userID, ok := getUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未授权"})
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	offset, limit := utils.ResolveOffsetLimit(page, pageSize)
	rows, err := sql.NewCollectMapperDefault().ListByUser(userID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "查询收藏失败"})
		return
	}
	ctx.JSON(http.StatusOK, rows)
}

// FollowUser 关注用户
//
//	@Summary		关注用户
//	@Description	当前用户关注目标用户，写入 Redis followers 集合
//	@Tags			interactions
//	@Accept			json
//	@Produce		json
//	@Param			targetID	path		int	true	"目标用户ID"
//	@Success		200			{object}	types.SuccessResponse
//	@Failure		400			{object}	types.ErrorResponse
//	@Failure		401			{object}	types.ErrorResponse
//	@Failure		500			{object}	types.ErrorResponse
//	@Router			/interactions/follow/{targetID} [post]
func FollowUser(ctx *gin.Context) {
	followerID, ok := getUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未授权"})
		return
	}
	targetID, err := strconv.Atoi(ctx.Param("targetID"))
	if err != nil || targetID <= 0 {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	if followerID == targetID {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "不能关注自己"})
		return
	}

	feedService := redis.NewFeedService(redis.Client)
	if err = feedService.AddFollower(targetID, followerID); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "关注失败"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "关注成功"})
}

// UnfollowUser 取消关注用户
//
//	@Summary		取消关注用户
//	@Description	当前用户取消关注目标用户，从 Redis followers 集合中移除
//	@Tags			interactions
//	@Accept			json
//	@Produce		json
//	@Param			targetID	path		int	true	"目标用户ID"
//	@Success		200			{object}	types.SuccessResponse
//	@Failure		400			{object}	types.ErrorResponse
//	@Failure		401			{object}	types.ErrorResponse
//	@Failure		500			{object}	types.ErrorResponse
//	@Router			/interactions/follow/{targetID} [delete]
func UnfollowUser(ctx *gin.Context) {
	followerID, ok := getUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.ErrorResponse{Message: "未授权"})
		return
	}
	targetID, err := strconv.Atoi(ctx.Param("targetID"))
	if err != nil || targetID <= 0 {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "参数错误"})
		return
	}
	if followerID == targetID {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Message: "不能取消关注自己"})
		return
	}

	feedService := redis.NewFeedService(redis.Client)
	if err = feedService.RemoveFollower(targetID, followerID); err != nil {
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Message: "取消关注失败"})
		return
	}
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "取消关注成功"})
}
