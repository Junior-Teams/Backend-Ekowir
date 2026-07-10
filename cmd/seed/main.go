package main

import (
	"log"

	"github.com/ALZEE23/ApiGo/database"
	"github.com/ALZEE23/ApiGo/seeds"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on process environment")
	}

	database.ConnectDb()
	seeds.RunSeeders()
}
