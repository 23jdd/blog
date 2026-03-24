package middle

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ReadLimiter struct {
	Limit int64      // 读取token限制
	mu    sync.Mutex // 互斥锁
	count int64      // 当前token数量
	Start int64      // 开始时间
	rate  int64      // 读取速率
}

// 令牌桶算法
func (r *ReadLimiter) Request(count int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.count += r.rate * (time.Now().Unix() - r.Start)
	r.Start = time.Now().Unix()
	if r.count >= r.Limit {
		r.count = r.Limit
	}
	if r.count < count {
		return false
	}
	r.count -= count
	return true
}
func NewReadLimiter(limit int64, rate int64) *ReadLimiter {
	return &ReadLimiter{
		Limit: limit,
		Start: time.Now().Unix(),
		rate:  rate,
	}
}

// ReadLimitMiddlerWare 读取限制中间件
func ReadLimitMiddlerWare(limit int64, rate int64) gin.HandlerFunc {
	readLimiter := NewReadLimiter(limit, rate)
	return func(ctx *gin.Context) {
		if !readLimiter.Request(1) {
			ctx.JSON(429, gin.H{
				"message": "read limit exceeded",
			}) // 429 错误
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

//
