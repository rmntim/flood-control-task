package floodcontrol

import (
	"context"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type RedisFloodControl struct {
	client *redis.Client
	window time.Duration
	limit  int
}

func NewRedisFloodControl(redisAddr string, window time.Duration, limit int) *RedisFloodControl {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &RedisFloodControl{
		client: client,
		window: window,
		limit:  limit,
	}
}

func (fc *RedisFloodControl) Check(ctx context.Context, userID int64) (bool, error) {
	now := time.Now().Unix()
	earliest := now - int64(fc.window.Seconds())

	nowStr := strconv.FormatInt(now, 10)
	earliestStr := strconv.FormatInt(earliest, 10)

	reqId := uuid.New()

	countCmd := fc.client.ZCount(ctx, "flood_control:"+strconv.FormatInt(userID, 10), earliestStr, nowStr)
	count, err := countCmd.Result()
	if err != nil {
		return false, err
	}

	if count >= int64(fc.limit) {
		return false, nil
	}

	addCmd := fc.client.ZAdd(ctx, "flood_control:"+strconv.FormatInt(userID, 10), redis.Z{Score: float64(now), Member: reqId.String()})
	_, err = addCmd.Result()
	if err != nil {
		return false, err
	}

	expireCmd := fc.client.Expire(ctx, "flood_control:"+strconv.FormatInt(userID, 10), fc.window)
	_, err = expireCmd.Result()
	if err != nil {
		return false, err
	}

	return true, nil
}
