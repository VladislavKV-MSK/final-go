package main

import (
	"fmt"
	"go1f/pkg/db"
	"go1f/pkg/server"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// Проверяем существование файла .env
	if err := godotenv.Load(); err != nil {
		log.Println("ошибка загрузки .env")
	}
	// Создаем БД
	db.InitDB()
	defer db.CloseDB()

	// Запускаем сервер
	if ok := server.Run(); ok != nil {
		fmt.Println("Server is not running....")
	}
}
