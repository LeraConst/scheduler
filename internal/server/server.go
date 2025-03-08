package server

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Start запускает HTTP-сервер
func Start(db *sql.DB) {

	// Определяем путь к исполняемому файлу
	appPath, err := os.Executable()
	if err != nil {
		log.Fatal("Ошибка получения пути к исполняемому файлу:", err)
	}

	// Определяем путь к директории web
	webPath := filepath.Join(filepath.Dir(appPath), "..", "web")
	// Проверяем, существует ли директория web
	_, err = os.Stat(webPath)
	if os.IsNotExist(err) {
		log.Fatalf("Директория %s не найдена", webPath)
	}

	// Определяем порт сервера
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = ":7540"
	}

	// Создаем файловый сервер, который раздает файлы из папки web/
	fs := http.FileServer(http.Dir(webPath))
	// Регистрируем файловый обработчик
	http.Handle("/", fs)

	http.HandleFunc("/api/signin", SigninHandler)
	http.HandleFunc("/api/nextdate", NextDateHandler)
	http.HandleFunc("/api/task", AuthMiddleware(func(res http.ResponseWriter, req *http.Request) {
		ManageTaskHandler(res, req, db)
	}))
	http.HandleFunc("/api/tasks", AuthMiddleware(func(res http.ResponseWriter, req *http.Request) {
		GetTaskHandler(res, req, db)
	}))
	http.HandleFunc("/api/task/done", AuthMiddleware(func(res http.ResponseWriter, req *http.Request) {
		DoneTaskHandler(res, req, db)
	}))

	// Запускаем сервер
	log.Println("Запуск сервера на порту", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
