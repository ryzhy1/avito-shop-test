package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type Storage struct {
	db         *redis.Client
	refreshTTL time.Duration
}

func InitRedis(connStr, redisPassword, redisDbNumber string, refreshTTL time.Duration) (*Storage, error) {
	dbNumber, err := strconv.Atoi(redisDbNumber)
	if err != nil {
		return nil, err
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     connStr,
		Username: "",
		Password: redisPassword,
		DB:       dbNumber,
	})
	return &Storage{db: redisClient, refreshTTL: refreshTTL}, nil
}

var ctx = context.Background()

func (s *Storage) StoreRefreshToken(userID, refreshToken string) error {
	err := s.db.Set(ctx, refreshToken, userID, s.refreshTTL).Err()
	if err != nil {
		return err
	}

	return nil
}
