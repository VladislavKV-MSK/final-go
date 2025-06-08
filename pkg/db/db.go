// Package db предоставляет функционал для работы с базой данных задач.
// Использует SQLite в качестве хранилища данных.
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

// Структура задачи в БД
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// ErrorResponse представляет структуру для возврата ошибок в API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// TasksResp представляет структуру для возврата списка задач в API.
type TasksResp struct {
	Tasks []*Task `json:"tasks"`
}

var dbTask *sql.DB

// InitDB инициализирует базу данных SQLite.
// Если файл БД уже существует, проверяет его целостность.
// Создает таблицу scheduler и индекс по дате, если они не существуют.
func InitDB() {

	dbPath := os.Getenv("TODO_DBFILE")
	if _, err := os.Stat(dbPath); err == nil {
		fmt.Println("Файл БД уже существует, проверяем целостность... ")
	}

	// Открываем/создаем базу данных
	var err error
	dbTask, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal("Ошибка открытия БД: ", err)
	}

	// SQL запрос для создания таблицы
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,          -- Формат YYYYMMDD (20060102)
		title TEXT NOT NULL,
		comment TEXT,
		repeat VARCHAR(128)        -- Правила повторений (макс 128 символов)
	);
	
	CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
	`

	// Выполняем SQL запрос-создание
	if _, err := dbTask.Exec(createTableSQL); err != nil {
		log.Fatal("Ошибка при инициализации БД: ", err)
	}

	log.Println("База данных успешно инициализирована")
}

// GetDB возвращает экземпляр подключения к базе данных (опционально).
// Паникует, если база данных не была инициализирована.
func GetDB() *sql.DB {
	if dbTask == nil {
		panic("База данных не инициализирована. Сначала вызывается InitDB()")
	}
	return dbTask
}

// CloseDB закрывает соединение с базой данных.
// Возвращает nil, если соединение уже закрыто.
func CloseDB() error {
	if dbTask != nil {
		return dbTask.Close()
	}
	return nil
}

// AddTask добавляет новую задачу в базу данных.
// Принимает указатель на Task, возвращает ID созданной записи и ошибку.
func AddTask(task *Task) (int64, error) {
	var id int64
	// определяем запрос
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)`
	res, err := dbTask.Exec(query,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err
}

// GetTasks возвращает список задач из базы данных, отсортированный по дате.
// Параметр limit ограничивает количество возвращаемых записей.
// Возвращает ошибку, если limit отрицательный.
func GetTasks(limit int) ([]*Task, error) {

	// Проверяем, что limit не отрицательный
	if limit < 0 {
		return nil, fmt.Errorf("limit не может быть отрицательным")
	}

	query := "SELECT * FROM scheduler ORDER BY date ASC LIMIT :limit"

	rows, err := dbTask.Query(query, sql.Named("limit", limit))
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	// Создаем слайс для хранения результатов
	var tasks []*Task

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}
	// Проверяем ошибки, которые могли возникнуть при итерации
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return tasks, nil
}

// SearchTasks выполняет поиск задач по строке или дате.
// Если строка является валидной датой (в формате DD.MM.YYYY), ищет задачи на эту дату.
// Иначе ищет задачи, содержащие строку в title или comment.
// Параметр limit ограничивает количество результатов.
func SearchTasks(s string, limit int) ([]*Task, error) {
	// Проверяем, что limit не отрицательный
	if limit < 0 {
		return nil, fmt.Errorf("limit не может быть отрицательным")
	}

	var date bool
	var query string

	t, err := time.Parse("02.01.2006", s)
	if err == nil {
		s = t.Format("20060102")
		date = true
	}

	if date {
		query = `SELECT * FROM scheduler WHERE date = :search LIMIT :limit`
	} else {
		query = `
        SELECT * 
        FROM scheduler
        WHERE title LIKE '%' || :search || '%' 
           OR comment LIKE '%' || :search || '%'
        ORDER BY date DESC
        LIMIT :limit
    `
	}

	rows, err := dbTask.Query(query,
		sql.Named("search", s),
		sql.Named("limit", limit))
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	// Создаем слайс для хранения результатов
	var tasks []*Task

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}
	// Проверяем ошибки, которые могли возникнуть при итерации
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return tasks, nil
}

// GetTaskID возвращает задачу по её ID.
// Если задача не найдена, возвращает ошибку.
func GetTaskID(id string) (Task, error) {

	var task Task
	query := `SELECT * FROM scheduler WHERE id = :id`

	row := dbTask.QueryRow(query, sql.Named("id", id))
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return task, err
	}

	return task, nil
}

// PutTaskID обновляет задачу в базе данных по её ID.
// Возвращает ошибку, если задача не найдена или произошла ошибка при обновлении.
func PutTaskID(task *Task) error {

	query := `
	UPDATE scheduler 
	SET 
		date = :date,
		title = :title,
		comment = :comment,
		repeat = :repeat
	WHERE id = :id`

	res, err := dbTask.Exec(query,
		sql.Named("id", task.ID),
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}

// DeleteTaskID удаляет задачу из базы данных по её ID.
// Возвращает ошибку, если задача не найдена или произошла ошибка при удалении.
func DeleteTaskID(id string) error {
	res, err := dbTask.Exec("DELETE FROM scheduler WHERE id = :id",
		sql.Named("id", id))
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}
