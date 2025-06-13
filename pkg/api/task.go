// Package api предоставляет функционал для работы API сервиса.
package api

import (
	"encoding/json"
	"fmt"
	"go1f/pkg/db"
	"go1f/pkg/taskdate"
	"log"
	"net/http"
	"sync"
	"time"
)

// ErrorResponse представляет структуру для возврата ошибок в API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// TasksResp представляет структуру для возврата списка задач в API.
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

var taskMutex sync.Mutex
var errTask error = fmt.Errorf("ошибка Task")

// taskHandler обрабатывает HTTP-запросы для работы с задачами.
// В зависимости от метода запроса (GET, POST, PUT, DELETE) вызывает соответствующий обработчик.
// Если метод не поддерживается, возвращает ошибку 405 Method Not Allowed.
func taskHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		handleGetTask(w, r)
	case http.MethodPost:
		handlePostTask(w, r)
	case http.MethodPut:
		handlePutTask(w, r)
	case http.MethodDelete:
		handleDeleteTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePostTask обрабатывает POST-запрос для создания новой задачи.
// Принимает JSON с данными задачи в теле запроса.
// Проверяет валидность данных, добавляет задачу в БД и возвращает ID созданной задачи.
// В случае ошибки возвращает соответствующий HTTP-статус и описание ошибки.
func handlePostTask(w http.ResponseWriter, r *http.Request) {
	var newTask db.Task

	err := json.NewDecoder(r.Body).Decode(&newTask)
	if err != nil {
		log.Println("Ошибка при разборе JSON")
		sendError(w, "Неверный формат JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	text, err := checkTask(&newTask)
	if err != nil {
		sendError(w, text, http.StatusBadRequest)
		return
	}

	taskMutex.Lock()
	defer taskMutex.Unlock()
	id, err := db.AddTask(&newTask)
	if err != nil {
		log.Println("Ошибка при добавлении задачи в БД")
		sendError(w, "Ошибка при добавлении задачи в БД", http.StatusInternalServerError)
		return
	}

	sendJSON(w, map[string]int64{"id": id}, http.StatusCreated)

}

// handleGetTask обрабатывает GET-запрос для получения задачи по ID.
// ID задачи передается в параметре запроса "id".
// Возвращает JSON с данными задачи или ошибку, если задача не найдена.
func handleGetTask(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, "id задачи не задан", http.StatusBadRequest)
		return
	}

	resp, err := db.GetTaskID(id)
	if err != nil {
		sendError(w, fmt.Sprintf("задача с id =%v не найдена", id), http.StatusBadRequest)
		return
	}

	sendJSON(w, resp, http.StatusOK)

}

// handlePutTask обрабатывает PUT-запрос для обновления существующей задачи.
// Принимает JSON с обновленными данными задачи в теле запроса.
// Проверяет валидность данных и обновляет задачу в БД.
// Возвращает пустой ответ со статусом 200 OK или описание ошибки.
func handlePutTask(w http.ResponseWriter, r *http.Request) {

	var task db.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		sendError(w, "Неверный формат JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	mess, err := checkTask(&task)
	if err != nil {
		sendError(w, mess, http.StatusBadRequest)
		return
	}

	taskMutex.Lock()
	defer taskMutex.Unlock()

	if err := db.PutTaskID(&task); err != nil {
		log.Println("Ошибка при сохранении задачи в БД")
		sendError(w, "Ошибка сохранения: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendJSON(w, struct{}{}, http.StatusOK)

}

// handleDeleteTask обрабатывает DELETE-запрос для удаления задачи по ID.
// ID задачи передается в параметре запроса "id".
// Возвращает пустой ответ со статусом 200 OK или описание ошибки.
func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, "id задачи не задан", http.StatusBadRequest)
		return
	}

	taskMutex.Lock()
	defer taskMutex.Unlock()

	err := db.DeleteTaskID(id)
	if err != nil {
		log.Println("Ошибка при удалении задачи из БД")
		sendError(w, "ошибка удаления", http.StatusInternalServerError)
		return
	}

	sendJSON(w, struct{}{}, http.StatusOK)
}

// handleDoneTask обрабатывает POST-запрос для завершения задачи.
// Для одноразовых задач - удаляет их, для повторяющихся - вычисляет следующую дату выполнения.
// ID задачи передается в параметре запроса "id".
// Возвращает пустой ответ со статусом 200 OK или описание ошибки.
func handleDoneTask(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, "id задачи не задан", http.StatusBadRequest)
		return
	}
	task, err := db.GetTaskID(id)
	if err != nil {
		sendError(w, fmt.Sprintf("задача с id =%v не найдена", id), http.StatusBadRequest)
		return
	}

	if task.Repeat == "" {
		// Удаляем одноразовую задачу
		err = db.DeleteTaskID(id)
		if err != nil {
			log.Println("Ошибка при удалении задачи из БД")
			sendError(w, "ошибка удаления", http.StatusInternalServerError)
			return
		}
	} else {
		// Персчитываем дату для задачи
		newDate, err := taskdate.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			log.Println("Ошибка при пересчете даты задачи из БД")
			sendError(w, "ошибка при расчете новой даты", http.StatusInternalServerError)
			return
		}
		// Обновляем задачу в БД
		task.Date = newDate
		db.PutTaskID(&task)
	}

	sendJSON(w, struct{}{}, http.StatusOK)
}

// nextDayHandler обрабатывает запрос для вычисления следующей даты выполнения задачи.
// Принимает параметры:
//   - now (опционально) - текущая дата в формате YYYYMMDD
//   - date - исходная дата задачи
//   - repeat - правило повторения
//
// Возвращает новую дату в формате YYYYMMDD или описание ошибки.
func nextDayHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	var now time.Time
	var err error
	nowParam := r.FormValue("now")

	// Если параметр не пустой парсим его
	if nowParam != "" {
		now, err = time.Parse(taskdate.DateFormat, nowParam)
		if err != nil {
			log.Println("Ошибка с получением текущей даты")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Если параметр пустой используем текущую дату
		now = time.Now()
	}

	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	date, err = taskdate.NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := w.Write([]byte(date)); err != nil {
		log.Printf("Ошибка при записи ответа по дате: %v \n", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// sendJSON отправляет ответ в формате JSON с указанным HTTP-статусом.
// Принимает:
//   - w - ResponseWriter для записи ответа
//   - resp - данные для сериализации в JSON
//   - status - HTTP-статус ответа
//
// В случае ошибки сериализации отправляет ошибку 500 Internal Server Error.
func sendJSON(w http.ResponseWriter, resp any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println("Ошибка при формировании JSON")
		sendError(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
	}
}

// sendError отправляет ошибку в формате JSON с указанным HTTP-статусом.
// Принимает:
//   - w - ResponseWriter для записи ответа
//   - message - текст сообщения об ошибке
//   - statusCode - HTTP-статус ошибки
func sendError(w http.ResponseWriter, message string, statusCode int) {
	response := ErrorResponse{
		Error: message,
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// checkTask проверяет валидность данных задачи.
// Проверяет:
//   - наличие заголовка (Title)
//   - корректность формата даты
//   - актуальность даты (при необходимости вычисляет следующую дату по правилу повторения)
//
// Возвращает текст ошибки и nil, если проверка прошла успешно,
// или текст ошибки и errTask, если найдены ошибки.
// Может модифицировать дату задачи для приведения к корректному значению.
func checkTask(t *db.Task) (string, error) {

	// Проверка на пустоту заголовка
	if t.Title == "" {
		return "Поле Title не должно быть пустым", errTask
	}

	now := time.Now()
	today := now.Format(taskdate.DateFormat)

	// Обработка пустой даты
	if t.Date == "" {
		t.Date = today
		return "", nil
	}

	// Парсинг даты
	_, err := time.Parse(taskdate.DateFormat, t.Date)
	if err != nil {
		return "Поле Date указано неверно", errTask
	}

	// Если дата в будущем или сегодняшнаяя - оставляем без изменений
	if t.Date >= today {
		return "", nil
	}

	if t.Repeat == "" {
		// Без правила - ставим сегодня
		t.Date = today
	} else {
		// С правилом - вычисляем следующую доступную дату
		next, err := taskdate.NextDate(now, t.Date, t.Repeat)
		if err != nil {
			return "Неверное правило повторения: " + err.Error(), errTask
		}
		t.Date = next
	}
	return "", nil
}
