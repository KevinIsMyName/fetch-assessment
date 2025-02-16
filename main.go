package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type ReceiptResponse struct {
	ID string `json:"id"`
}

type PointsResponse struct {
	Points int `json:"points"`
}

var receipts = make(map[string]Receipt)

func generateUuid() string {
	newUuid := uuid.New().String()
	for _, exists := receipts[newUuid]; exists; _, exists = receipts[newUuid] {
		newUuid = uuid.New().String()
	}
	return newUuid
}

func isAlphanumeric(char byte, alphabet []byte, numbers []byte) bool {
	for _, letter := range alphabet {
		if char == letter {
			return true
		}
	}
	for _, number := range numbers {
		if char == number {
			return true
		}
	}
	return false
}

func main() {
	http.HandleFunc("/receipts/process", processReceiptHandler)
	http.HandleFunc("/receipts/", getPointsHandler)

	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func processReceiptHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var receipt Receipt
	if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
		http.Error(w, "Invalid receipt format", http.StatusBadRequest)
		return
	}

	receiptId := generateUuid()
	receipts[receiptId] = receipt

	resp := ReceiptResponse{ID: receiptId}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getPointsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/receipts/"), "/points")
	if id == "" {
		http.Error(w, "Receipt ID is required", http.StatusBadRequest)
		return
	}

	fmt.Println(receipts)
	fmt.Println(id)
	receipt, exists := receipts[id]
	if !exists {
		http.Error(w, "No receipt found for that ID", http.StatusNotFound)
		return
	}

	points := calculatePoints(receipt)

	resp := PointsResponse{Points: points}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func calculatePoints(receipt Receipt) int {
	points := 0

	// * One point for every alphanumeric character in the retailer name.
	alphabet := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}
	numbers := []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	for i := 0; i < len(receipt.Retailer); i++ {
		char := receipt.Retailer[i]
		if isAlphanumeric(char, alphabet, numbers) {
			points++
		}
	}

	// * 50 points if the total is a round dollar amount with no cents.
	total, _ := strconv.ParseFloat(receipt.Total, 64)
	if math.Mod(total, 1) == 0 {
		points += 50
	}

	// * 25 points if the total is a multiple of `0.25`.
	if math.Mod(total, 0.25) == 0 {
		points += 25
	}

	// * 5 points for every two items on the receipt.
	points += len(receipt.Items) / 2 * 5

	// * If the trimmed length of the item description is a multiple of 3, multiply the price by `0.2` and round up to the nearest integer. The result is the number of points earned.
	for _, item := range receipt.Items {
		descLength := len(strings.TrimSpace(item.ShortDescription))
		if descLength%3 == 0 {
			price, _ := strconv.ParseFloat(item.Price, 64)
			points += int(math.Ceil(price * 0.2))
		}
	}

	// * 6 points if the day in the purchase date is odd.
	purchaseDate, _ := time.Parse("2006-01-02", receipt.PurchaseDate) // date is trivial, only need format
	if purchaseDate.Day()%2 != 0 {
		points += 6
	}

	// * 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	purchaseTime, _ := time.Parse("15:04", receipt.PurchaseTime) // time is trivial, only need format
	if purchaseTime.Hour() >= 14 && purchaseTime.Hour() < 16 {
		points += 10
	}

	return points
}
