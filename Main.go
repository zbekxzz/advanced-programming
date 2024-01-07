package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type RequestBody struct {
	Location string `json:"location"`
}

type ResponseBody struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	http.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var requestBody RequestBody
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&requestBody); err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		if requestBody.Location == "" {
			http.Error(w, "Invalid JSON message", http.StatusBadRequest)
			return
		}

		fmt.Printf("Received new location: %s\n", requestBody.Location)

		location := requestBody.Location
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		weather := rng.Intn(41) - 10
		weatherStr := strconv.Itoa(weather)

		response := ResponseBody{
			Status:  "success",
			Message: "Weather for location " + location + " - " + weatherStr + " celsius degree",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	fmt.Println("Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}
