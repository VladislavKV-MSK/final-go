package main

import (
	"fmt"
	"go1f/pkg/config"
	"go1f/pkg/db"
	"go1f/pkg/server"
)

func main() {

	// Загружаем настройки сервера
	config.ConfigServer()

	// Создаем БД
	db.InitDB()
	defer db.CloseDB()

	// Запускаем сервер
	if err := server.Run(); err != nil {
		fmt.Println("Server is not running....")
	}
}
