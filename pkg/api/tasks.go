// Package api предоставляет функционал для работы API сервиса.
package api

import (
	"go1f/pkg/db"
	"log"
	"net/http"

	"go1f/pkg/config"
)

// tasksHandler обрабатывает HTTP-запросы для работы с задачами.
// Поддерживает только GET-запросы.
// Параметры запроса:
//   - search: строка для поиска задач по контексту или дате (необязательный)
//
// Если параметр search не указан, возвращает список задач с ограничением по количеству,
// которое задается переменной окружения TODO_LIMIT_TASKS (по умолчанию 50).
//
// В случае ошибки возвращает соответствующий HTTP-статус и сообщение об ошибке.
func tasksHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		log.Println("Ошибка метода запроса")
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	searchQuery := r.URL.Query().Get("search")

	if searchQuery == "" {
		// просто n задач
		tasks, err := db.GetTasks(config.App.LimitTask)
		if err != nil {
			log.Println("Ошибка при получении задачи из БД")
			sendError(w, "ошибка получения задач", http.StatusInternalServerError)
			return
		}
		sendResponse(w, tasks)
	} else {
		// n задач в которых есть определенные слова или даты
		tasks, err := db.SearchTasks(searchQuery, config.App.LimitTask)
		if err != nil {
			log.Println("Ошибка с поиском контекста в задачах")
			sendError(w, "ошибка поиска задач", http.StatusInternalServerError)
			return
		}
		sendResponse(w, tasks)
	}
}

// sendResponse формирует и отправляет JSON-ответ со списком задач.
// Если tasks равен nil, возвращает пустой массив задач.
func sendResponse(w http.ResponseWriter, tasks []*db.Task) {
	if tasks == nil {
		tasks = []*db.Task{}
	}

	resp := TasksResp{
		Tasks: tasks,
	}

	sendJSON(w, resp, http.StatusOK)
}
