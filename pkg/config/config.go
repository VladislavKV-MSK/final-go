/*
Package config предоставляет инструменты для загрузки конфигурации приложения через переменные окружения.

Конфигурация читается из .env файла или системных переменных окружения с приоритетом:
1. Переменные окружения ОС
2. Значения из .env файла
3. Встроенные значения по умолчанию

Основные настройки:
- Ограничение количества задач
- Порт веб-сервера
- Путь к файлу базы данных
- Тестовый пароль для доступа
*/
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Переменные из env импортируемые в другие пакеты
type Config struct {
	LimitTask    int
	PathToDB     string
	PortServ     string
	PasswordTest string
}

var App Config

// Значения по умолчанию для ключевых параметров приложения.
const (
	DefaultLimitTasks   = 50                   // Значение по умолчанию кол-ва отображаемых задач
	DefaultPort         = `7540`               // Значение по умолчнию порта
	DefaultPathDb       = `/data/scheduler.db` // Значение по умолчнию пути к БД
	DefaultTestPassword = `1234`               // Значение по умолчнию тестового пароля
)

// ConfigServer инициализирует систему конфигурации.
// Загружает переменные окружения из .env файла в корне проекта.
// Должен вызываться при старте приложения перед использованием других функций.
func ConfigServer() {
	// Загружаем файл .env
	_ = godotenv.Load()
	App = Config{
		LimitTask:    getLimitTasks(),
		PathToDB:     getPathDB(),
		PortServ:     getPort(),
		PasswordTest: getPassword()}

}

// getLimitTasks возвращает максимальное количество задач для отображения.
// Читает значение из переменной окружения TODO_LIMIT_TASKS.
// При ошибке парсинга или отсутствии или отрицательном значении возвращает DefaultLimitTasks = 50.
func getLimitTasks() int {
	if limitStr := os.Getenv("TODO_LIMIT_TASKS"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			log.Printf("Будет выведено до %v задач \n", limit)
			return limit
		}
	}
	log.Printf("Будет выведено до %v задач \n", DefaultLimitTasks)
	return DefaultLimitTasks
}

// getPort возвращает TCP-порт для HTTP сервера.
// Читает значение из переменной окружения TODO_PORT.
// Если значение не задано, возвращает DefaultPort = 7540.
func getPort() string {
	if port := os.Getenv("TODO_PORT"); port != "" {
		log.Printf("Сервер запущен на порту %v \n", port)
		return port
	}
	log.Printf("Сервер запущен на порту  %v (по умолчанию) \n", DefaultPort)
	return DefaultPort
}

// getPathDB возвращает путь к файлу базы данных SQLite.
// Читает значение из переменной окружения TODO_DBFILE.
// При отсутствии значения возвращает DefaultPathDb = "/data/scheduler.db".
func getPathDB() string {
	if pathDB := os.Getenv("TODO_DBFILE"); pathDB != "" {
		log.Printf("База данных будет открыта по пути: %v \n", pathDB)
		return pathDB
	}
	log.Printf("База данных будет октрыта по пути: %v \n", DefaultPathDb)
	return DefaultPathDb
}

// getPassword возвращает тестовый пароль для авторизации.
// Читает значение из переменной окружения TODO_PASSWORD.
// При отсутствии значения пароль не требуется.
// В случае ошибки выставляется пароль 1234.
func getPassword() string {
	if password := os.Getenv("TODO_PASSWORD"); password != "" {
		log.Printf("Пароль для входа %v \n", password)
		return password
	}
	log.Printf("Пароль для входа (по умолчанию) %v \n", DefaultTestPassword)
	return DefaultTestPassword
}
