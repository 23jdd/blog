package redis

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	feedKeyPrefix     = "feed:user:"
	followerKeyPrefix = "followers:"
	defaultLimit      = 10
)

var ctx = context.Background()

type FeedService struct {
	client *redis.Client
}

// 创建FeedService
func NewFeedService(client *redis.Client) *FeedService {
	return &FeedService{
		client: client,
	}
}

func getFeedKey(userID int) string {
	return fmt.Sprintf("%s%d", feedKeyPrefix, userID)
}

func getFollowerKey(userID int) string {
	return fmt.Sprintf("%s%d", followerKeyPrefix, userID)
}

// AddFollower 建立关注关系：followerID 关注 targetUserID。
func (fs *FeedService) AddFollower(targetUserID int, followerID int) error {
	key := getFollowerKey(targetUserID)
	err := fs.client.SAdd(ctx, key, strconv.Itoa(followerID)).Err()
	if err != nil {
		log.Printf("Failed to add follower %d for user %d: %v", followerID, targetUserID, err)
		return fmt.Errorf("failed to add follower: %w", err)
	}
	return nil
}

// RemoveFollower 取消关注关系：followerID 取消关注 targetUserID。
func (fs *FeedService) RemoveFollower(targetUserID int, followerID int) error {
	key := getFollowerKey(targetUserID)
	err := fs.client.SRem(ctx, key, strconv.Itoa(followerID)).Err()
	if err != nil {
		log.Printf("Failed to remove follower %d for user %d: %v", followerID, targetUserID, err)
		return fmt.Errorf("failed to remove follower: %w", err)
	}
	return nil
}

// GetFollowerIDs 返回指定作者的粉丝用户ID列表。
func (fs *FeedService) GetFollowerIDs(authorID int) ([]int, error) {
	key := getFollowerKey(authorID)
	members, err := fs.client.SMembers(ctx, key).Result()
	if err != nil {
		log.Printf("Failed to get followers for user %d: %v", authorID, err)
		return nil, fmt.Errorf("failed to get follower ids: %w", err)
	}
	if len(members) == 0 {
		return []int{}, nil
	}

	followerIDs := make([]int, 0, len(members))
	for _, v := range members {
		id, convErr := strconv.Atoi(v)
		if convErr != nil {
			log.Printf("Failed to parse follower id %s: %v", v, convErr)
			continue
		}
		followerIDs = append(followerIDs, id)
	}
	return followerIDs, nil
}

func (fs *FeedService) AddArticleToFeed(userID int, articleID int, timestamp time.Time) error {
	key := getFeedKey(userID)          // "feed:user:1"
	score := float64(timestamp.Unix()) // 时间戳
	member := strconv.Itoa(articleID)  // "1"
	err := fs.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Err()
	// 添加文章到Feed中如果失败，返回 500 错误
	if err != nil {
		log.Printf("Failed to add article %d to user %d feed: %v", articleID, userID, err)
		return fmt.Errorf("failed to add article to feed: %w", err)
	}

	log.Printf("Successfully added article %d to user %d feed at %v", articleID, userID, timestamp)
	return nil
}

func (fs *FeedService) GetLatestArticleIDs(userID int, limit int) ([]int, error) {
	if limit <= 0 {
		limit = defaultLimit
	}

	key := getFeedKey(userID)

	start := int64(0)
	stop := int64(limit - 1)

	result, err := fs.client.ZRevRange(ctx, key, start, stop).Result()
	if err != nil {
		log.Printf("Failed to get latest articles for user %d: %v", userID, err)
		return nil, fmt.Errorf("failed to get latest articles: %w", err)
	}

	if len(result) == 0 {
		log.Printf("No articles found in feed for user %d", userID)
		return []int{}, nil
	}

	articleIDs := make([]int, 0, len(result))
	for _, idStr := range result {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Failed to parse article ID %s: %v", idStr, err)
			continue
		}
		articleIDs = append(articleIDs, id)
	}

	log.Printf("Successfully retrieved %d article IDs for user %d", len(articleIDs), userID)
	return articleIDs, nil
}

func (fs *FeedService) AddArticlesToFeedBatch(userID int, articles []ArticleInfo) error {
	if len(articles) == 0 {
		log.Printf("No articles to add to feed for user %d", userID)
		return nil
	}

	key := getFeedKey(userID)
	zs := make([]redis.Z, 0, len(articles))

	for _, article := range articles {
		member := strconv.Itoa(article.ID)
		score := float64(article.Timestamp.Unix())
		zs = append(zs, redis.Z{
			Score:  score,
			Member: member,
		})
	}

	err := fs.client.ZAdd(ctx, key, zs...).Err()
	if err != nil {
		log.Printf("Failed to batch add %d articles to user %d feed: %v", len(articles), userID, err)
		return fmt.Errorf("failed to batch add articles to feed: %w", err)
	}

	log.Printf("Successfully batch added %d articles to user %d feed", len(articles), userID)
	return nil
}

func (fs *FeedService) GetArticleCount(userID int) (int64, error) {
	key := getFeedKey(userID)

	count, err := fs.client.ZCard(ctx, key).Result()
	if err != nil {
		log.Printf("Failed to get article count for user %d: %v", userID, err)
		return 0, fmt.Errorf("failed to get article count: %w", err)
	}

	log.Printf("User %d has %d articles in feed", userID, count)
	return count, nil
}

func (fs *FeedService) RemoveArticleFromFeed(userID int, articleID int) error {
	key := getFeedKey(userID)
	member := strconv.Itoa(articleID)

	err := fs.client.ZRem(ctx, key, member).Err()
	if err != nil {
		log.Printf("Failed to remove article %d from user %d feed: %v", articleID, userID, err)
		return fmt.Errorf("failed to remove article from feed: %w", err)
	}

	log.Printf("Successfully removed article %d from user %d feed", articleID, userID)
	return nil
}

func (fs *FeedService) ClearUserFeed(userID int) error {
	key := getFeedKey(userID)

	err := fs.client.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Failed to clear feed for user %d: %v", userID, err)
		return fmt.Errorf("failed to clear user feed: %w", err)
	}

	log.Printf("Successfully cleared feed for user %d", userID)
	return nil
}

func (fs *FeedService) GetFeedWithPagination(userID int, page int, pageSize int) ([]int, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultLimit
	}

	key := getFeedKey(userID)

	totalCount, err := fs.client.ZCard(ctx, key).Result()
	if err != nil {
		log.Printf("Failed to get total article count for user %d: %v", userID, err)
		return nil, 0, fmt.Errorf("failed to get total article count: %w", err)
	}

	start := int64((page - 1) * pageSize)
	stop := int64(page*pageSize - 1)

	result, err := fs.client.ZRevRange(ctx, key, start, stop).Result()
	if err != nil {
		log.Printf("Failed to get paginated articles for user %d: %v", userID, err)
		return nil, 0, fmt.Errorf("failed to get paginated articles: %w", err)
	}

	articleIDs := make([]int, 0, len(result))
	for _, idStr := range result {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Failed to parse article ID %s: %v", idStr, err)
			continue
		}
		articleIDs = append(articleIDs, id)
	}

	log.Printf("Successfully retrieved %d article IDs for user %d (page %d, page size %d)", len(articleIDs), userID, page, pageSize)
	return articleIDs, totalCount, nil
}

func (fs *FeedService) GetArticlesInRange(userID int, startTime, endTime time.Time) ([]int, error) {
	key := getFeedKey(userID)

	minScore := float64(startTime.Unix())
	maxScore := float64(endTime.Unix())

	result, err := fs.client.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: strconv.FormatFloat(minScore, 'f', -1, 64),
		Max: strconv.FormatFloat(maxScore, 'f', -1, 64),
	}).Result()

	if err != nil {
		log.Printf("Failed to get articles in range for user %d: %v", userID, err)
		return nil, fmt.Errorf("failed to get articles in range: %w", err)
	}

	articleIDs := make([]int, 0, len(result))
	for _, idStr := range result {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Failed to parse article ID %s: %v", idStr, err)
			continue
		}
		articleIDs = append(articleIDs, id)
	}

	log.Printf("Successfully retrieved %d article IDs for user %d in time range %v to %v", len(articleIDs), userID, startTime, endTime)
	return articleIDs, nil
}

func (fs *FeedService) TrimOldArticles(userID int, keepCount int) error {
	if keepCount <= 0 {
		return fmt.Errorf("keep count must be positive")
	}

	key := getFeedKey(userID)

	err := fs.client.ZRemRangeByRank(ctx, key, 0, int64(-keepCount-1)).Err()
	if err != nil {
		log.Printf("Failed to trim old articles for user %d: %v", userID, err)
		return fmt.Errorf("failed to trim old articles: %w", err)
	}

	log.Printf("Successfully trimmed old articles for user %d, keeping latest %d", userID, keepCount)
	return nil
}

func (fs *FeedService) HealthCheck() error {
	_, err := fs.client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Redis health check failed: %v", err)
		return fmt.Errorf("redis connection failed: %w", err)
	}

	log.Printf("Redis health check passed")
	return nil
}
