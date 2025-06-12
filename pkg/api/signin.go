package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go1f/pkg/config"

	"github.com/golang-jwt/jwt/v5"
)

// Pass представляет структуру для парсинга входящего JSON-запроса с паролем.
// Используется в обработчике /api/signin.
type Pass struct {
	Password string `json:"password"`
}

// RespSign представляет структуру для успешного ответа с JWT-токеном.
// Возвращается при успешной аутентификации.
type RespSign struct {
	Token string `json:"token"`
}

// handleSignIn обрабатывает POST-запрос на аутентификацию (/api/signin).
//
// Принимает JSON вида {"password":"string"}.
// Сравнивает пароль с значением из переменной окружения TODO_PASSWORD.
//
// В случае успеха возвращает JWT-токен в формате:
//
//	{"token":"eyJhbGciOiJ..."}
//
// Возможные ошибки:
//   - 405: метод не POST
//   - 400: неверный формат JSON или аутентификация не настроена
//   - 401: неверный пароль или ошибка генерации токена
func handleSignIn(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		log.Println("Ошибка метода запроса при аутентификации")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var password Pass

	err := json.NewDecoder(r.Body).Decode(&password)
	if err != nil {
		sendError(w, "Неверный формат JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	secretPassword := config.App.PasswordTest
	if secretPassword == "" {
		sendError(w, "Аутентификация не настроена", http.StatusBadRequest)
		return
	}

	if password.Password != secretPassword {
		log.Printf("Введен невенрный пароль %v", password.Password)
		sendError(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	resp, err := getToken(secretPassword)
	if err != nil {
		sendError(w, "Ошибка получения токена", http.StatusUnauthorized)
		return
	}

	sendJSON(w, RespSign{resp}, http.StatusOK)
}

// getToken генерирует JWT-токен на основе пароля.
//
// Пароль хешируется с помощью SHA-256, результат используется как:
//   - Секрет для подписи токена (алгоритм HS256)
//   - Полезная нагрузка (claim "pwd_hash")
//
// Токен имеет срок жизни 8 часов (claim "exp").
//
// Возвращает:
//   - string: подписанный токен в формате JWT
//   - error: ошибка при подписании
func getToken(s string) (string, error) {

	// Создаём хэш пароля для использования в качестве секрета
	hash := sha256.Sum256([]byte(s))
	secret := hex.EncodeToString(hash[:])

	// Создаём полезную нагрузку claims с хэшем пароля
	claims := jwt.MapClaims{
		"pwd_hash": secret,
		"exp":      time.Now().Add(8 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	result, err := token.SignedString([]byte(secret))
	return result, err
}

// auth — middleware для проверки JWT-токена из куки.
//
// Если TODO_PASSWORD не задан, аутентификация пропускается.
//
// Проверяет:
//  1. Наличие куки "token"
//  2. Алгоритм подписи (должен быть HS256)
//  3. Соответствие секрета (хеш пароля из токена и env)
//  4. Срок действия токена
//
// В случае ошибки возвращает:
//   - 401: кука отсутствует/токен невалиден/пароль изменён
func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		secretPassword := config.App.PasswordTest
		if secretPassword == "" {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("token")
		if err != nil {
			sendError(w, "Требуется аутентификация", http.StatusUnauthorized)
			return
		}

		// Проверка токена
		hash := sha256.Sum256([]byte(secretPassword))
		secret := hex.EncodeToString(hash[:])

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			sendError(w, "Неверный токен", http.StatusUnauthorized)
			return
		}

		// Дополнительная проверка хэша пароля
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if claims["pwd_hash"] != secret {
				sendError(w, "Пароль изменен", http.StatusUnauthorized)
				return
			}
		}
		// вызов следующего обработчика
		next(w, r)
	}
}
