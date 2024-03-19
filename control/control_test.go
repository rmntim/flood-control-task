package floodcontrol

import (
	"context"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"strconv"
	"testing"
	"time"
)

func TestRedisFloodControl_Check(t *testing.T) {
	client, mock := redismock.NewClientMock()

	fc := &RedisFloodControl{
		client: client,
		window: 10 * time.Second,
		limit:  5,
	}

	userID := int64(123)

	now := time.Now().Unix()
	earliest := now - int64(fc.window.Seconds())

	nowStr := strconv.FormatInt(now, 10)
	earliestStr := strconv.FormatInt(earliest, 10)

	// Test case 1: Within limit
	mock.ExpectZCount("flood_control:123", earliestStr, nowStr).SetVal(int64(4))
	mock.Regexp().ExpectZAdd("flood_control:123", redis.Z{Score: float64(now), Member: "[a-z0-9\\-]+"}).SetVal(1)
	mock.ExpectExpire("flood_control:123", 10*time.Second).SetVal(true)

	passed, err := fc.Check(context.Background(), userID)
	if err != nil {
		t.Errorf("Error occurred: %v", err)
	}
	if !passed {
		t.Error("Expected check to pass but it failed")
	}

	// Test case 2: Exceeds limit
	mock.ExpectZCount("flood_control:123", earliestStr, nowStr).SetVal(int64(5))

	passed, err = fc.Check(context.Background(), userID)
	if err != nil {
		t.Errorf("Error occurred: %v", err)
	}
	if passed {
		t.Error("Expected check to fail but it passed")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
