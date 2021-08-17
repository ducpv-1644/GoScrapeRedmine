package config

import (
	"fmt"
	"github.com/joho/godotenv"
)

func LoadENV() {
	err := godotenv.Load(".env")

	if err != nil {
		fmt.Println("Error loading .env file!")
		return
	}
}
