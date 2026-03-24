package redis

import (
	"context"
	"time"
)

type Session struct {
	SessionID string
	Map       map[string]string
}

func NewSession(sessionID string) *Session {
	return &Session{
		SessionID: sessionID,
		Map:       make(map[string]string),
	}
}
func GetSession(sessionID string) (*Session, error) {
	mapData, err := Client.HGetAll(context.Background(), sessionID).Result()
	if err != nil {
		return nil, err
	}
	return &Session{
		SessionID: sessionID,
		Map:       mapData,
	}, nil
}
func (s *Session) Save(Expire time.Duration) error {
	err := Client.HSet(context.Background(), s.SessionID, s.Map).Err()
	if err == nil {
		return Client.Expire(context.Background(), s.SessionID, Expire).Err()
	}
	return err
}
func (s *Session) Get(key string) (string, error) {
	return Client.HGet(context.Background(), s.SessionID, key).Result()
}
func (s *Session) Set(key, value string) error {
	return Client.HSet(context.Background(), s.SessionID, key, value).Err()
}
func (s *Session) Delete() error {
	return Client.Del(context.Background(), s.SessionID).Err()
}
func (s *Session) GetAll() (map[string]string, error) {
	Data, err := Client.HGetAll(context.Background(), s.SessionID).Result()
	if err != nil {
		return nil, err
	}
	return Data, nil
}
