package redis

import "time"

type ArticleInfo struct {
	ID        int
	Timestamp time.Time
}