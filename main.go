package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:56268", // kalau pakai container mappingnya ke host
	})

	lockKey := "meeting:lock"
	lockValue := "1"

	// coba set lock selama 5 detik
	ok, err := rdb.SetNX(ctx, lockKey, lockValue, 5*time.Second).Result()
	if err != nil {
		panic(err)
	}

	if !ok {
		fmt.Println("Request sedang diproses, coba lagi nanti.")
		return
	}

	fmt.Println("Lock didapat, proses request...")

	// simulasi proses
	time.Sleep(3 * time.Second)

	// hapus lock
	rdb.Del(ctx, lockKey)
	fmt.Println("Lock dilepas, request selesai.")
}
