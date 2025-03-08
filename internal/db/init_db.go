package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// InitDatabase инициализирует БД
func InitDatabase() *sql.DB {
	dbFile := os.Getenv("TODO_DBFILE") // Получаем путь к БД из переменной окружения

	// Если переменная окружения не задана, используем путь по умолчанию
	if dbFile == "" {
		appPath, err := os.Executable() // возвращаем путь к исполняемому файлу
		if err != nil {
			log.Fatal("Ошибка получения пути к исполняемому файлу:", err)
		}
		// возвращаем директорию из переданного пути appPath и склеиваем переданные строки в корректный путь
		dbFile = filepath.Join(filepath.Dir(appPath), "..", "internal", "db", "scheduler.db")
		log.Println("Используем путь к БД по умолчанию:", dbFile)
	} else {
		log.Println("Используем путь к БД из переменной окружения:", dbFile)
	}

	// проверяем существование файла базы данных по его пути dbFile
	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err) // True, если файла нет

	// Открываем БД
	log.Println("Открываем БД по пути:", dbFile)
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal("Ошибка при открытии БД:", err)
	}

	// Проверяем подключение к БД
	err = db.Ping()
	if err != nil {
		log.Fatal("Ошибка при подключении к БД:", err)
	}

	if install {
		CreateTable(db)
	} else {
		log.Println("БД уже существует, пропускаем создание.")
	}

	return db
}

// CreateTable создает БД
func CreateTable(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE scheduler (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date CHAR(8) NOT NULL DEFAULT "",
			title VARCHAR(256) NOT NULL DEFAULT "",
			comment TEXT NOT NULL DEFAULT "",
			repeat VARCHAR(128)
		);
		CREATE INDEX task_date ON scheduler(date);
		`)
	if err != nil {
		log.Fatal("Ошибка при создании таблицы:", err)
	}
	log.Println("БД и таблица scheduler успешно созданы.")
}
