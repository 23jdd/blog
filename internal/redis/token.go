package redis

import (
	"context"
	"fmt"
	"time"

	R "github.com/redis/go-redis/v9"
)

const (
	loginFailLimit = 5
	loginFailTTL   = 15 * time.Minute
)

// SaveRefreshToken 保存 refresh token 的 jti
func SaveRefreshToken(jti string, userID int, ttl time.Duration) error {
	key := fmt.Sprintf("refresh:%s", jti)
	return Client.Set(context.Background(), key, userID, ttl).Err()
}

// VerifyRefreshToken 校验 refresh token jti 是否有效
func VerifyRefreshToken(jti string) (bool, error) {
	key := fmt.Sprintf("refresh:%s", jti)
	count, err := Client.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteRefreshToken 删除 refresh token jti
func DeleteRefreshToken(jti string) error {
	key := fmt.Sprintf("refresh:%s", jti)
	return Client.Del(context.Background(), key).Err()
}

// BlacklistAccessToken 将 access token jti 加入黑名单
func BlacklistAccessToken(jti string, ttl time.Duration) error {
	key := fmt.Sprintf("blacklist:access:%s", jti)
	return Client.Set(context.Background(), key, "1", ttl).Err()
}

// IsAccessTokenBlacklisted 检查 access token jti 是否在黑名单
func IsAccessTokenBlacklisted(jti string) (bool, error) {
	key := fmt.Sprintf("blacklist:access:%s", jti)
	count, err := Client.Exists(context.Background(), key).Result() // 检查 token 是否在黑名单中
	if err != nil {
		return false, err
	}
	return count > 0, nil // 如果 token 在黑名单中，返回 true
}

// IncreaseLoginFailCount 增加登录失败计数
func IncreaseLoginFailCount(username string) (int64, error) {
	key := fmt.Sprintf("login_fail:%s", username)
	count, err := Client.Incr(context.Background(), key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		_ = Client.Expire(context.Background(), key, loginFailTTL).Err()
	}
	return count, nil
}

// ClearLoginFailCount 清理登录失败计数
func ClearLoginFailCount(username string) error {
	key := fmt.Sprintf("login_fail:%s", username)
	return Client.Del(context.Background(), key).Err()
}

// IsLoginBlocked 判断用户是否因为连续失败被限制登录
func IsLoginBlocked(username string) (bool, int64, error) {
	key := fmt.Sprintf("login_fail:%s", username)
	count, err := Client.Get(context.Background(), key).Int64()
	if err != nil {
		if err == R.Nil {
			return false, 0, nil
		}
		return false, 0, err
	}
	return count >= loginFailLimit, count, nil
}
