package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"time"
)

const k = 30
const initialRating = 1000
const noOfPointReductionMatches = 25
const noOfPointReduction = 40

func main() {
	file, err := os.Open("matches.csv")
	if err != nil {
		log.Fatal("Unable to open file:", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read all the records from the CSV file
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Unable to read file:", err)
	}

	ratings := make(map[string]float64)
	lastPlayedMatch := make(map[string]int)

	for i, record := range records {
		player1 := record[0]
		player2 := record[1]
		winner := record[2]
		player1Rating, ok := ratings[player1]
		if !ok {
			ratings[player1] = initialRating
			player1Rating = initialRating
		}
		player2Rating, ok := ratings[player2]
		if !ok {
			ratings[player2] = initialRating
			player2Rating = initialRating
		}

		player1LastMatch, exist := lastPlayedMatch[player1]
		if exist && player1LastMatch < i-noOfPointReductionMatches {
			ratings[player1] = ratings[player1] - noOfPointReduction*math.Floor(float64(i-player1LastMatch)/noOfPointReductionMatches)
			player1Rating = ratings[player1]
		}
		player2LastMatch, exist := lastPlayedMatch[player2]
		if exist && player2LastMatch < i-noOfPointReductionMatches {
			ratings[player2] = ratings[player2] - noOfPointReduction*math.Floor(float64(i-player2LastMatch)/noOfPointReductionMatches)
			player2Rating = ratings[player2]
		}

		p1 := 1 / (1 + math.Pow(10, (player2Rating-player1Rating)/400))
		p2 := 1 / (1 + math.Pow(10, (player1Rating-player2Rating)/400))

		player1Actual := 0.0
		player2Actual := 0.0
		if winner == player1 {
			player1Actual = 1
			player2Actual = 0
		} else {
			player2Actual = 1
			player1Actual = 0
		}
		ratings[player1] = player1Rating + k*(player1Actual-p1)
		ratings[player2] = player2Rating + k*(player2Actual-p2)

		// fmt.Printf("Match %d:", i+1)
		// for player, rating := range ratings {
		// 	fmt.Printf(" %s: %.2f", player, rating)
		// }
		// fmt.Println()
		lastPlayedMatch[player1] = i
		lastPlayedMatch[player2] = i
	}

	for player := range ratings {
		if lastPlayedMatch[player] < len(records)-noOfPointReductionMatches {
			ratings[player] = ratings[player] - noOfPointReduction*math.Floor(float64(len(records)-lastPlayedMatch[player])/noOfPointReductionMatches)
		}
	}

	// sort the players by rating
	players := make([]string, 0, len(ratings))

	for player := range ratings {
		players = append(players, player)
	}
	sort.Slice(players, func(i, j int) bool {
		return ratings[players[i]] > ratings[players[j]]
	})

	fmt.Println("Final ratings")
	msg := "Final ratings as of"
	currentDate := time.Now().Format("2006-01-02")
	msg = fmt.Sprintf("%s %s :", msg, currentDate)
	for _, player := range players {
		fmt.Printf(" %s: %.2f", player, ratings[player])
		msg += fmt.Sprintf("\n%s: %.2f", player, ratings[player])
	}

	url := ""
	sendMessageToGoogleChat(url, msg)

}

func sendMessageToGoogleChat(webhookURL, message string) error {
	// Create the JSON payload
	payload := map[string]string{"text": message}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Make the POST request
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 response: %v", resp.Status)
	}

	return nil
}
