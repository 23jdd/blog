package sql

import (
	"blog/internal/model"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

func ExampleReviewMapperUsage() {
	db, err := sqlx.Connect("mysql", "user:password@tcp(localhost:3306)/database")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	reviewMapper := NewReviewMapper(db)

	newReview := &model.Review{
		ArticleID: 1,
		Content:   "这是一篇很棒的文章！",
		AuthorID:  100,
		IsDirect:  true,
		ParentID:  0,
	}

	id, err := reviewMapper.Create(newReview)
	if err != nil {
		log.Printf("创建评论失败: %v", err)
		return
	}
	fmt.Printf("成功创建评论，ID: %d\n", id)

	review, err := reviewMapper.FindByID(int(id))
	if err != nil {
		log.Printf("查询评论失败: %v", err)
		return
	}
	fmt.Printf("查询到的评论: %+v\n", review)

	articles, err := reviewMapper.FindByArticleID(1)
	if err != nil {
		log.Printf("查询文章评论失败: %v", err)
		return
	}
	fmt.Printf("文章1的所有评论数量: %d\n", len(articles))

	replies, err := reviewMapper.FindByParentID(int(id))
	if err != nil {
		log.Printf("查询回复失败: %v", err)
		return
	}
	fmt.Printf("评论的回复数量: %d\n", len(replies))

	review.Content = "更新后的评论内容"
	err = reviewMapper.Update(review)
	if err != nil {
		log.Printf("更新评论失败: %v", err)
		return
	}
	fmt.Printf("成功更新评论\n")

	err = reviewMapper.DeleteByID(int(id))
	if err != nil {
		log.Printf("删除评论失败: %v", err)
		return
	}
	fmt.Printf("成功删除评论\n")

	directReviews, err := reviewMapper.FindDirectReviews(1)
	if err != nil {
		log.Printf("查询直接评论失败: %v", err)
		return
	}
	fmt.Printf("直接评论数量: %d\n", len(directReviews))

	count, err := reviewMapper.FindByArticleIDCount(1)
	if err != nil {
		log.Printf("查询评论数量失败: %v", err)
		return
	}
	fmt.Printf("文章1的评论总数: %d\n", count)

	pagedReviews, err := reviewMapper.FindByArticleIDWithPagination(1, 10, 0)
	if err != nil {
		log.Printf("分页查询失败: %v", err)
		return
	}
	fmt.Printf("第一页评论数量: %d\n", len(pagedReviews))

	startTime := time.Now().AddDate(0, -1, 0)
	endTime := time.Now()
	rangeReviews, err := reviewMapper.FindWithTimeRange(1, startTime, endTime)
	if err != nil {
		log.Printf("时间范围查询失败: %v", err)
		return
	}
	fmt.Printf("最近一个月的评论数量: %d\n", len(rangeReviews))

	latestReviews, err := reviewMapper.FindLatestByAuthorID(100, 5)
	if err != nil {
		log.Printf("查询最新评论失败: %v", err)
		return
	}
	fmt.Printf("用户100的最新5条评论数量: %d\n", len(latestReviews))

	popularReviews, err := reviewMapper.FindPopularReviews(10)
	if err != nil {
		log.Printf("查询热门评论失败: %v", err)
		return
	}
	fmt.Printf("热门评论数量: %d\n", len(popularReviews))

	batchReviews := []*model.Review{
		{ArticleID: 1, Content: "批量评论1", AuthorID: 100, IsDirect: true, ParentID: 0},
		{ArticleID: 1, Content: "批量评论2", AuthorID: 100, IsDirect: true, ParentID: 0},
		{ArticleID: 1, Content: "批量评论3", AuthorID: 100, IsDirect: true, ParentID: 0},
	}
	lastID, err := reviewMapper.BatchCreate(batchReviews)
	if err != nil {
		log.Printf("批量创建评论失败: %v", err)
		return
	}
	fmt.Printf("批量创建评论成功，最后ID: %d\n", lastID)

	exists, err := reviewMapper.ExistsByID(1)
	if err != nil {
		log.Printf("检查评论存在性失败: %v", err)
		return
	}
	fmt.Printf("评论1是否存在: %v\n", exists)

	recentReviews, err := reviewMapper.FindRecentReviews(10)
	if err != nil {
		log.Printf("查询最近评论失败: %v", err)
		return
	}
	fmt.Printf("最近10条评论数量: %d\n", len(recentReviews))

	totalCount, err := reviewMapper.FindAllCount()
	if err != nil {
		log.Printf("查询总评论数失败: %v", err)
		return
	}
	fmt.Printf("系统总评论数: %d\n", totalCount)
}
