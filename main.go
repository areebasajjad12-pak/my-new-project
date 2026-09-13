package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

var db *sql.DB
var rdb *redis.Client
var ctx = context.Background()

func main() {
	_ = godotenv.Load()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open DB connection: ", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	fmt.Println("Connected to database successfully")

	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatal("Failed to connect to Redis: ", err)
	}
	fmt.Println("Connected to Redis successfully")

	http.HandleFunc("/users", createUserHandler)

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Write to Postgres (source of truth)
	_, err := db.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", u.Name, u.Email)
	if err != nil {
		http.Error(w, "Failed to insert user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Mirror to Redis (cache)
	userJSON, err := json.Marshal(u)
	if err != nil {
		log.Println("Warning: failed to marshal user for cache:", err)
	} else {
		cacheKey := "user:" + u.Email
		if err := rdb.Set(ctx, cacheKey, userJSON, 0).Err(); err != nil {
			log.Println("Warning: failed to cache user in Redis:", err)
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "user created"})
}
