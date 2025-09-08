package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"web-api/handlers"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Healthy!!")
}

type OutputContent struct {
	Text string `json:"text"`
}

type OutputItem struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Content []OutputContent `json:"content"`
}

type Response struct {
	Output []OutputItem `json:"output"`
}

func checkinHandler(w http.ResponseWriter, r *http.Request) {
	url := "https://api.openai.com/v1/responses"

	payload := map[string]any{
		"prompt": map[string]any{
			"id": "pmpt_68b2181fe7348190a88a27ac360085a2045165832abddf4a",
			"variables": map[string]string{
				"yesterday_intent":    "I wanted to learn about AI for the render SRE projectI",
				"yesterday_completed": "I read a chapter in the O'Reilly AI book",
				"yesterday_friction":  "None",
				"today_intent":        "Getting this AI manager working",
				"today_questions":     "Nope",
				"big_picture":         "I'd like to be more autonomous ",
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}

	apiKey := "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"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to call OpenAI", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	// get body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		// Handle error
	}

	// Get the first ID
	jsonText := response.Output[0].Content[0].Text

	var data map[string]any
	if err := json.Unmarshal([]byte(jsonText), &data); err != nil {
		log.Fatal("bad json:", err)
	}

	out, _ := json.MarshalIndent(data, "", "  ")

	w.Write(out)
}

func main() {
	// Make .env optional; ignore missing file
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	port := ":8080"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get("http://example.com/")

		if err != nil {
			json.NewEncoder(w).Encode("there was a problem")
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			json.NewEncoder(w).Encode("Body was bad")
			return
		}

		// json.NewEncoder(w).Encode(string(body))
		w.Write(body)
	})

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/llm", checkinHandler)

	// Auth endpoints
	authHandler := handlers.NewAuthHander(db)
	http.HandleFunc("/api/register", authHandler.Register)
	http.HandleFunc("/api/login", authHandler.Login)

	// Requires login
	http.HandleFunc("/api/profile", func(w http.ResponseWriter, r *http.Request) {
		item := r.Header["Authorization"]

		newStr, _ := strings.CutPrefix(item[0], "Bearer ")
		log.Println(newStr)
		claims, _ := handlers.ValidateToken(newStr, handlers.Secret)
		log.Println(claims)
		if claims.Email != "bruceape@gmail.com" {
			log.Println("You're NOT allowed")
			return
		}

		log.Println("You're allowed")
		// If authentication token is legit
		// And we're authorizated to do this
		// We may continue
	})

	log.Printf("Server started on port %s", port)
	http.ListenAndServe(port, nil)
}
