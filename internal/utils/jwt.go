package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Claims struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func getJWTSecret() string {
	viper.SetDefault("jwt.secret", "blog-default-secret") // 设置默认 secret
	return viper.GetString("jwt.secret")
}

func getAccessExpireMinutes() int {
	viper.SetDefault("jwt.access_expire_minutes", 30)
	return viper.GetInt("jwt.access_expire_minutes")
}

func getRefreshExpireHours() int {
	viper.SetDefault("jwt.refresh_expire_hours", 168)
	return viper.GetInt("jwt.refresh_expire_hours")
}

func GenerateToken(userID int, username string) (string, error) {
	return GenerateAccessToken(userID, username)
}

func GenerateAccessToken(userID int, username string) (string, error) {
	now := time.Now()                                                                // 当前时间
	expirationTime := now.Add(time.Duration(getAccessExpireMinutes()) * time.Minute) // 过期时间
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(), // 生成一个唯一的 ID
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	} // 创建 claims 对象

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(getJWTSecret())) // 签名 token
}

func GenerateRefreshToken(userID int, username string) (string, *Claims, error) {
	now := time.Now()                                                             // 当前时间
	expirationTime := now.Add(time.Duration(getRefreshExpireHours()) * time.Hour) // 过期时间
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(), // 生成一个唯一的 ID
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	} // 创建 claims 对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(getJWTSecret()))
	if err != nil {
		return "", nil, err
	}
	return signed, claims, nil // 返回签名 token 和 claims 对象
}

func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(getJWTSecret()), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, errors.New("invalid token")
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeAccess { // 如果 token 类型不是 access，返回 401 错误
		return nil, errors.New("invalid token type")
	}
	return claims, nil // 返回 claims 对象
}

func ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeRefresh { // 如果 token 类型不是 refresh，返回 401 错误
		return nil, errors.New("invalid token type")
	}
	return claims, nil // 返回 claims 对象
}
