package main

import (
	"context"
	floodcontrol "task/control"
	"time"
)

func main() {
	redisFlood := floodcontrol.NewRedisFloodControl("localhost:6379", 10*time.Second, 10)
	ctx := context.Background()
	id := int64(1)

	for i := 0; i < 15; i++ {
		passed, err := redisFlood.Check(ctx, id)
		if err != nil {
			panic(err)
		}
		if passed {
			println("passed")
		} else {
			println("failed")
		}
	}
}

// FloodControl интерфейс, который нужно реализовать.
// Рекомендуем создать директорию-пакет, в которой будет находиться реализация.
type FloodControl interface {
	// Check возвращает false если достигнут лимит максимально разрешенного
	// кол-ва запросов согласно заданным правилам флуд контроля.
	Check(ctx context.Context, userID int64) (bool, error)
}
