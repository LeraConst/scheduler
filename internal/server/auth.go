package server

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Секретный ключ для подписи
var JWTSecret = []byte("my_secret_key")

// Claims - структура для данных в JWT-токене
type Claims struct {
	PasswordHash string `json:"password_hash"`
	jwt.RegisteredClaims
}

// SigninHandler обрабатывает запрос на вход, выдает JWT
func SigninHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, `{"error": "Ошибка чтения тела запроса"}`, http.StatusBadRequest)
		return
	}

	var request struct {
		Password string `json:"password"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(res, `{"error": "Неверный формат запроса"}`, http.StatusBadRequest)
		return
	}

	// Проверяем сохраненный пароль
	storedPassword := os.Getenv("TODO_PASSWORD")
	if storedPassword == "" {
		http.Error(res, `{"error": "Аутентификация отключена"}`, http.StatusUnauthorized)
		return
	}

	if request.Password != storedPassword {
		http.Error(res, `{"error": "Неверный пароль"}`, http.StatusUnauthorized)
		return
	}

	claims := Claims{
		PasswordHash: storedPassword,
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := jwtToken.SignedString(JWTSecret)
	if err != nil {
		http.Error(res, `{"error": "Ошибка генерации токена"}`, http.StatusInternalServerError)
		return
	}

	// Устанавливаем токен в куки
	http.SetCookie(res, &http.Cookie{
		Name:  "token",
		Value: signedToken,
	})

	response := map[string]string{"token": signedToken}

	// Преобразуем в JSON
	resultToken, err := json.Marshal(response)
	if err != nil {
		http.Error(res, `{"error": "Ошибка кодирования JSON"}`, http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок и записываем JSON
	res.Header().Set("Content-Type", "application/json")
	if _, err := res.Write(resultToken); err != nil {
		http.Error(res, `{"error": "Ошибка записи в ответ"}`, http.StatusInternalServerError)
	}

}

// AuthMiddleware проверяет JWT-токен в заголовке Authorization или в cookie "token"
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var tokenString string

		// Проверяем заголовок Authorization
		authHeader := req.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// Если заголовка нет, проверяем cookie "token"
			cookie, err := req.Cookie("token")
			if err == nil {
				tokenString = cookie.Value
			}
		}

		// Если токена нет, возвращаем ошибку
		if tokenString == "" {
			http.Error(res, `{"error": "Неавторизованный доступ"}`, http.StatusUnauthorized)
			return
		}

		// Проверяем токен
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return JWTSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(res, `{"error": "Недействительный токен"}`, http.StatusUnauthorized)
			return
		}

		// Токен действителен, передаем управление следующему обработчику
		next.ServeHTTP(res, req)
	}
}
