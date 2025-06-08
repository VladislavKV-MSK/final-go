// Package server предоставляет функционал для запуска HTTP-сервера приложения.
// Сервер использует порт, указанный в переменной окружения TODO_PORT.
package server

import (
	"fmt"
	"go1f/pkg/api"
	"net/http"
	"os"
)

// Run запускает HTTP-сервер приложения.
// Инициализирует API и начинает прослушивание указанного порта.
// Возвращает ошибку в случае проблем с запуском сервера.
//
// Порт для прослушивания берется из переменной окружения TODO_PORT.
func Run() error {

	port := os.Getenv("TODO_PORT")

	api.Init()

	return http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
