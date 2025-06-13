// Package server предоставляет функционал для запуска HTTP-сервера приложения.
// Сервер использует порт, указанный в переменной окружения TODO_PORT.
package server

import (
	"fmt"
	"go1f/pkg/api"
	"go1f/pkg/config"
	"net/http"
)

// Run запускает HTTP-сервер приложения.
// Инициализирует API и начинает прослушивание указанного порта.
// Возвращает ошибку в случае проблем с запуском сервера.
//
// Порт для прослушивания берется из переменной окружения TODO_PORT.
func Run() error {

	port := config.App.PortServ

	api.Init()

	return http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
