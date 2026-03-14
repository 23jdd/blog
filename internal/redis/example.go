package redis

import (
	"fmt"
	"log"
	"time"
)

func ExampleUsage() {
	feedService := NewFeedService(Client)

	err := feedService.HealthCheck()
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	userID := 123

	articles := []ArticleInfo{
		{ID: 1, Timestamp: time.Now().Add(-2 * time.Hour)},
		{ID: 2, Timestamp: time.Now().Add(-1 * time.Hour)},
		{ID: 3, Timestamp: time.Now()},
	}

	err = feedService.AddArticlesToFeedBatch(userID, articles)
	if err != nil {
		log.Printf("Failed to add articles: %v", err)
		return
	}

	latestArticles, err := feedService.GetLatestArticleIDs(userID, 10)
	if err != nil {
		log.Printf("Failed to get latest articles: %v", err)
		return
	}

	fmt.Printf("Latest articles for user %d: %v\n", userID, latestArticles)

	count, err := feedService.GetArticleCount(userID)
	if err != nil {
		log.Printf("Failed to get article count: %v", err)
		return
	}

	fmt.Printf("Total articles: %d\n", count)

	pageArticles, total, err := feedService.GetFeedWithPagination(userID, 1, 2)
	if err != nil {
		log.Printf("Failed to get paginated articles: %v", err)
		return
	}

	fmt.Printf("Page 1 articles: %v, Total: %d\n", pageArticles, total)

	startTime := time.Now().Add(-3 * time.Hour)
	endTime := time.Now()
	rangeArticles, err := feedService.GetArticlesInRange(userID, startTime, endTime)
	if err != nil {
		log.Printf("Failed to get articles in range: %v", err)
		return
	}

	fmt.Printf("Articles in range: %v\n", rangeArticles)

	err = feedService.RemoveArticleFromFeed(userID, 1)
	if err != nil {
		log.Printf("Failed to remove article: %v", err)
		return
	}

	err = feedService.TrimOldArticles(userID, 10)
	if err != nil {
		log.Printf("Failed to trim articles: %v", err)
		return
	}

	err = feedService.ClearUserFeed(userID)
	if err != nil {
		log.Printf("Failed to clear feed: %v", err)
		return
	}
}