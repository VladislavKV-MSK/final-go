# Task Scheduler (Final Go Project)

![Go](https://img.shields.io/badge/Go-1.24-blue)
![SQLite](https://img.shields.io/badge/SQLite-3-lightgrey)
![GitHub Actions](https://img.shields.io/badge/GitHub_Actions-passing-brightgreen)

Простое и эффективное приложение для управления задачами, написанное на Go с использованием SQLite в качестве базы данных.
В рамках проекта выплнены все задания со звездочкой:

- использование переменных окружения;
- обработка правил повторения по неделям и месяцам;
- поиск задачи по контексту или дате;
- аутентификация пользователя.

## 📖 Документация

### Godoc
Все публичные функции и методы имеют подробные комментарии в стиле godoc. Чтобы сгенерировать документацию:


```bash
# Локальный просмотр
GO111MODULE=on godoc -http=:6060
```
Затем откройте http://localhost:6060/pkg/github.com/VladislavKV-MSK/final-go/

## 📌 Особенности

- 📅 Планирование задач с указанием даты и времени
- 🔄 Поддержка повторяющихся задач
- 🔍 Поиск задач по дате или ключевым словам
- 📱 Минималистичный веб-интерфейс
- 🐳 Готовая Docker-конфигурация
- ✅ Автоматические тесты GitHub Actions

## 🚀 Быстрый старт

### Требования
- Go 1.24+
- SQLite 3
- Docker v4.41.2
  
### Установка
```bash
git clone https://github.com/VladislavKV-MSK/final-go.git
cd final-go
go mod download
```

### Конфигурация
Создайте .env файл:
```
TODO_PORT=7540
TODO_DBFILE=scheduler.db 
LIMIT_TASKS=50
TODO_PASSWORD=your_password
```
### Запуск
```bash
go run main.go
```
Приложение будет доступно на http://localhost:7540/login.html

### 🐳 Docker
Сборка осуществляется командой 
```bash
docker build -t todo-app .
```
А запуск одним из двух способов
```bash
docker run -p 7540:7540 --env-file .env -it todo-app
docker run -d   -p 7540:7540   -v app-data:/data   -e TODO_PORT=7540   -e TODO_DBFILE=/data/scheduler.db   -e LIMIT_TASKS=50   -e TODO_PASSWORD=your_password -it todo-app
```

### 📂 Структура проекта
```text
final-go/
├── .github/           # GitHub Actions workflows
├── pkg/
    ├── api/           # Основная логика приложения
│   ├── db/            # Работа с базой данных
│   └── server/        # HTTP обработчики
├── tests/             # Тесты
├── web/               # Веб-интерфейс
│   ├── static/        # Статические файлы
│   └── templates/     # HTML шаблоны
├── go.mod             # Зависимости
├── main.go            # Точка входа
└── scheduler.db       # База данных SQLite
```
## 🛠️ API Endpoints

| Метод  | Путь           | Описание                      |
|--------|----------------|-------------------------------|
| GET    | `/tasks`       | Получить список задач         |
| POST   | `/tasks`       | Добавить новую задачу         |
| PUT    | `/tasks/{id}`  | Обновить существующую задачу  |
| DELETE | `/tasks/{id}`  | Удалить задачу                |


### 🤖 Тестирование
Запуск тестов:
Перед тестами нужно обновить переменную в tests/settings.go
```
var Token = `новый 8-часовой токен`
```
Чтобы его получить сделайте запрос :
```
curl -X POST http://localhost:7540/api/signin \
     -H "Content-Type: application/json" \
     -d '{"password":"ваш_пароль_из_TODO_PASSWORD"}'
```

```bash
go test -v ./...
```



