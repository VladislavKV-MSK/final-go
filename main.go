package main

import (
	"fmt"
	"go1f/pkg/db"
	"go1f/pkg/server"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем файла .env
	_ = godotenv.Load()

	// Создаем БД
	db.InitDB()
	defer db.CloseDB()

	// Запускаем сервер
	if ok := server.Run(); ok != nil {
		fmt.Println("Server is not running....")
	}
}
