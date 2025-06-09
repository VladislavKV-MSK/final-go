// Package api предоставляет функционал для работы API сервиса.
package api

import "net/http"

// Init инициализирует маршруты HTTP-сервера.
//
// Регистрирует следующие обработчики:
//   - GET /api/nextdate - обработчик для получения следующей даты
//   - /api/task - обработчик для работы с отдельной задачей (CRUD операции)
//   - /api/tasks - обработчик для получения списка задач
//   - /api/task/done - обработчик для отметки задачи как выполненной
//   - / - обработчик для обслуживания статических файлов из директории "web"
func Init() {
	http.HandleFunc("/api/nextdate", nextDayHandler)
	http.HandleFunc("/api/task", auth(taskHandler))
	http.HandleFunc("/api/tasks", auth(tasksHandler))
	http.HandleFunc("/api/task/done", auth(handleDoneTask))
	http.HandleFunc("/api/signin", handleSignIn)

	http.Handle("/", http.FileServer(http.Dir("web"))) //последним идет обработчик для статичных файлов, чтобы не перекрывать остальные
}
