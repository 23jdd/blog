package redis

import (
	"context"
	"fmt"
	"strconv"
)

const (
	hotKey         = "hot:articles"
	articleViewKey = "article:%d:views"
)

// IncHotByRead 增加阅读热度
func IncHotByRead(articleID int) error {
	return Client.ZIncrBy(context.Background(), hotKey, 1, strconv.Itoa(articleID)).Err()
}

// IncHotByLike 增加点赞热度
func IncHotByLike(articleID int) error {
	return Client.ZIncrBy(context.Background(), hotKey, 3, strconv.Itoa(articleID)).Err()
}

// DecHotByLike 取消点赞时降低热度
func DecHotByLike(articleID int) error {
	return Client.ZIncrBy(context.Background(), hotKey, -3, strconv.Itoa(articleID)).Err()
}

// IncHotByComment 增加评论热度
func IncHotByComment(articleID int) error {
	return Client.ZIncrBy(context.Background(), hotKey, 2, strconv.Itoa(articleID)).Err()
}

// IncArticleView 增加文章访问量
func IncArticleView(articleID int) (int64, error) {
	return Client.Incr(context.Background(), fmt.Sprintf(articleViewKey, articleID)).Result()
}

// GetArticleView 获取文章访问量
func GetArticleView(articleID int) (int64, error) {
	return Client.Get(context.Background(), fmt.Sprintf(articleViewKey, articleID)).Int64()
}
func GetArticleUserView(userID int) (int64, error) {
	//return Client.Get(context.Background(), fmt.Sprintf(articleUserViewKey, userID)).Int64()
	return 0, nil
}

// GetHot 获取热榜文章ID列表
func GetHot(sum int) ([]int, error) {
	if sum <= 0 {
		sum = 10
	}
	raw, err := Client.ZRevRange(context.Background(), hotKey, 0, int64(sum-1)).Result() // 获取热门文章ID列表
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(raw)) // 创建热门文章ID列表
	for _, v := range raw {
		id, convErr := strconv.Atoi(v) // 将热门文章ID转换为 int 类型
		if convErr != nil {
			continue
		}
		ids = append(ids, id) // 将热门文章ID添加到热门文章ID列表
	}
	return ids, nil // 返回热门文章ID列表
}
