package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
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
	// Create context with timeout for external API call
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

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

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
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
		http.Error(w, "bad upstream response", http.StatusBadGateway)
		return
	}

	out, _ := json.MarshalIndent(data, "", "  ")

	w.Write(out)
}

func main() {
	// Make .env optional; ignore missing file
	_ = godotenv.Load()

	secret := os.Getenv("JWT_SECRET")
	issuer := os.Getenv("JWT_ISSUER")
	if secret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// Configure connection pool for production
	db.SetMaxOpenConns(25)                 // max concurrent connections
	db.SetMaxIdleConns(5)                  // max idle connections in pool
	db.SetConnMaxLifetime(5 * time.Minute) // max time a connection can be reused
	db.SetConnMaxIdleTime(1 * time.Minute) // max time a connection can be idle

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
	authHandler := handlers.NewAuthHandler(db, secret, issuer)
	http.HandleFunc("/api/register", authHandler.Register)
	http.HandleFunc("/api/login", authHandler.Login)

	// Requires login (use auth middleware)
	http.Handle("/api/profile", authHandler.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteJSON(w, http.StatusOK, map[string]string{"you": "are in!"})
	})))

	srv := &http.Server{
		Addr:              port,
		Handler:           nil, // use default mux
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Server started on port %s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
