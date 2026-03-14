package redis

import (
	"context"
	"time"
)

// GetNextID 获取下一个ID
func GetNextID(ctx context.Context, key string) (int64, error) {
	id, err := Client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return Decode(id), nil
}

func Decode(id int64) int64 {
	return time.Now().UnixNano()&int64(0xffffffff) | id
}
