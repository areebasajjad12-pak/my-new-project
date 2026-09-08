package main

import (
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	fmt.Println("Setup ready")
}