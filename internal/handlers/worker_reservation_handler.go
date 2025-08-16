package handlers

import (
	"context"
	"database/sql"
	"e_meeting/internal/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func WorkerReservation(rdb *redis.Client, dbConn *sql.DB) {
	ctx := context.Background()
	fmt.Println("Worker started, waiting for jobs...")

	for {
		// ambil job dari antrean, list reservation:queue hilang 1 persatu
		job, err := rdb.BLPop(ctx, 0*time.Second, "reservation:queue").Result()
		if err != nil {
			fmt.Println("Error reading queue:", err)
			continue
		}

		// mengambil data request dari redis yang telah dipush dengan marshal
		var request models.ReservationRequest
		if err := json.Unmarshal([]byte(job[1]), &request); err != nil { // job[1] adalah data request
			fmt.Println("Invalid job data:", err)
			continue
		}

		if err := ProcessReservation(dbConn, request); err != nil {
			fmt.Println("Failed to process reservation:", err)
		} else {
			fmt.Printf("Reservation processed for user %d\n", request.UserID)
		}
	}
}
