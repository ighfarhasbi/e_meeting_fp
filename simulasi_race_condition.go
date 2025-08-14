package main

import (
	"bytes"
	"e_meeting/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	apiURL := "http://localhost:8080/reservations"
	// JWT tolen yang valid
	token := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTUyMzYzODAsImlhdCI6MTc1NTE0OTk4MCwiaWQiOjI1LCJyb2xlIjoiYWRtaW4iLCJzdGF0dXMiOiJhY3RpdmUiLCJ0eXBlIjoiYWNjZXNzIiwidXNlcm5hbWUiOiJhZG1pbmRvY2tlciJ9.eZIpSHVJUIFZrXelZmb1nk2whNYtGP8XRNrjzcCvins"

	// siapkan payload
	reqBody := models.ReservationRequest{
		UserID:      25,
		Name:        "User Test",
		PhoneNumber: "081234567890",
		Company:     "Test Company",
		Notes:       "Simulasi race condition",
		Rooms: []models.RoomReservation{
			{
				ID:           1,
				SnackID:      2,
				StartTime:    "2025-08-14 09:00:00.000 +0700",
				EndTime:      "2025-08-14 11:00:00.000 +0700",
				Participants: 5,
			},
		},
	}

	// konversi ke JSON, untuk redis
	jsonData, _ := json.Marshal(reqBody)

	// buat WaitGroup untuk 10 request paralel
	var wg sync.WaitGroup
	wg.Add(10)

	start := time.Now()

	for i := 1; i <= 10; i++ {
		go func(user int) {
			defer wg.Done()

			// jika terjadi request error
			req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Printf("User %d: Request error: %v\n", user, err)
				return
			}

			// menambahkan header
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)

			// jika aplikasi belum berjalan maka akan error
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Printf("User %d: HTTP error: %v\n", user, err)
				return
			}
			defer resp.Body.Close()

			// menampilkan status code yang diambil dari response saat hit API
			fmt.Printf("User %d: Status Code %d\n", user, resp.StatusCode)
		}(i)
	}

	wg.Wait()

	// menampilkan waktu total semua request
	fmt.Println("All requests done in:", time.Since(start))
}
