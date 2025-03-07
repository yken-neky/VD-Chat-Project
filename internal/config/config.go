package config

import "github.com/joho/godotenv"

func LoadConfig() {
	godotenv.Load(".env")
}
