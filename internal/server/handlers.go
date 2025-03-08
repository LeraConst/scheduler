package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/LeraConst/scheduler/internal/repeat"
)

// Task - структура задачи
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// ManageTaskHandler добавляет, обновляет и удаляет задачи в БД
func ManageTaskHandler(res http.ResponseWriter, req *http.Request, db *sql.DB) {

	switch req.Method {
	// Добавление задачи в БД
	case "POST":
		log.Println("Вызван обработчик ManageTaskHandler с POST-запросом")

		var task Task
		var buf bytes.Buffer

		defer req.Body.Close()
		// Читаем тело запроса в буфер
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(res, `{"error": "Ошибка чтения тела запроса"}`, http.StatusBadRequest)
			return
		}

		// Декодируем JSON в структуру Task
		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			http.Error(res, `{"error": "Ошибка десериализации JSON"}`, http.StatusBadRequest)
			return
		}

		// Проверяем обязательное поле
		if task.Title == "" {
			http.Error(res, `{"error": "Не указан заголовок задачи"}`, http.StatusBadRequest)
			return
		}
		// Получаем сегодняшнюю дату
		now := time.Now().Format("20060102")

		// Проверяем поле date
		if task.Date == "" {
			task.Date = now
		} else {
			_, err := time.Parse("20060102", task.Date)
			if err != nil {
				http.Error(res, `{"error": "Некорректный формат даты"}`, http.StatusBadRequest)
				return
			}
		}

		// Если дата в прошлом и есть правило повторения, вычисляем следующую дату
		if task.Date < now {
			if task.Repeat == "" {
				task.Date = now
			} else {
				nextDate, err := repeat.RulesNextDate(time.Now(), task.Date, task.Repeat)
				if err != nil {
					http.Error(res, `{"error": "Некорректное правило повторения"}`, http.StatusBadRequest)
					return
				}
				task.Date = nextDate
			}
		}
		// Добавляем задачу в БД
		result, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
			sql.Named("date", task.Date),
			sql.Named("title", task.Title),
			sql.Named("comment", task.Comment),
			sql.Named("repeat", task.Repeat))

		if err != nil {
			http.Error(res, fmt.Sprintf(`{"error": "Ошибка при добавлении задачи: %v"}`, err), http.StatusInternalServerError)
			return
		}

		// Получаем ID добавленной задачи
		id, err := result.LastInsertId()
		if err != nil {
			http.Error(res, `{"error": "Ошибка получения ID задачи"}`, http.StatusInternalServerError)
			return
		}

		// Отправляем успешный ответ
		res.Header().Set("Content-Type", "application/json; charset=UTF-8") // Указываем, что ответ в формате JSON
		res.WriteHeader(http.StatusOK)                                      // Устанавливаем статус 200 OK

		response := fmt.Sprintf(`{"id": "%d"}`, id)

		if _, err := res.Write([]byte(response)); err != nil {
			http.Error(res, `{"error": "Ошибка записи в ответ"}`, http.StatusInternalServerError)
		}

		log.Println("Добавленена новая задача с id:", id)

	// Выводим задачу по id для дальнейшего редактирования
	case "GET":

		log.Println("Вызван обработчик ManageTaskHandler с GET-запросом")
		// Получаем параметры id из строки запроса
		taskId := req.URL.Query().Get("id")

		// Если передан id, ищем конкретную задачу
		if taskId == "" {
			http.Error(res, `{"error": "Не указан ID задачи"}`, http.StatusBadRequest)
			return
		}

		var task Task
		row := db.QueryRow("SELECT * FROM scheduler WHERE id = ?", taskId)
		err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		// Обработка ошибок
		if err == sql.ErrNoRows {
			http.Error(res, `{"error": "Задача не найдена"}`, http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(res, fmt.Sprintf(`{"error": "Ошибка при получении данных: %v"}`, err), http.StatusInternalServerError)
			return
		}
		// Формируем JSON-ответ
		RespondJSON(res, task)

	// Обновляем отредактированную задачу в БД
	case "PUT":

		log.Println("Вызван обработчик ManageTaskHandler с PUT-запросом")
		var task Task
		var buf bytes.Buffer

		defer req.Body.Close()

		// Читаем тело запроса в буфер
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(res, `{"error": "Ошибка чтения тела запроса"}`, http.StatusBadRequest)
			return
		}

		// Декодируем JSON в структуру Task
		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			http.Error(res, `{"error": "Ошибка десериализации JSON"}`, http.StatusBadRequest)
			return
		}

		// Проверяем обязательные поля
		if task.ID == "" {
			http.Error(res, `{"error": "Не указан ID задачи"}`, http.StatusBadRequest)
			return
		}

		if task.Title == "" {
			http.Error(res, `{"error": "Не указан заголовок задачи"}`, http.StatusBadRequest)
			return
		}
		// Получаем сегодняшнюю дату
		now := time.Now().Format("20060102")

		// Проверяем поле date
		if task.Date == "" {
			task.Date = now
		} else {
			_, err := time.Parse("20060102", task.Date)
			if err != nil {
				http.Error(res, `{"error": "Некорректный формат даты"}`, http.StatusBadRequest)
				return
			}
		}

		// Если дата в прошлом и есть правило повторения, вычисляем следующую дату
		if task.Date < now {
			if task.Repeat == "" {
				task.Date = now
			} else {
				nextDate, err := repeat.RulesNextDate(time.Now(), task.Date, task.Repeat)
				if err != nil {
					http.Error(res, `{"error": "Некорректное правило повторения"}`, http.StatusBadRequest)
					return
				}
				task.Date = nextDate
			}
		}
		// Обновляем задачу в БД
		result, err := db.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
			sql.Named("date", task.Date),
			sql.Named("title", task.Title),
			sql.Named("comment", task.Comment),
			sql.Named("repeat", task.Repeat),
			sql.Named("id", task.ID),
		)

		if err != nil {
			http.Error(res, fmt.Sprintf(`{"error": "Ошибка при обновлении задачи: %v"}`, err), http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(res, `{"error": "Ошибка проверки обновления"}`, http.StatusInternalServerError)
			return
		}

		if rowsAffected == 0 {
			http.Error(res, `{"error": "Задача не найдена"}`, http.StatusNotFound)
			return
		}

		// Отправляем успешный пустой JSON-ответ
		res.Header().Set("Content-Type", "application/json; charset=UTF-8") // Указываем, что ответ в формате JSON
		res.WriteHeader(http.StatusOK)
		if _, err := res.Write([]byte(`{}`)); err != nil {
			http.Error(res, `{"error": "Ошибка записи в ответ"}`, http.StatusInternalServerError)
		}

	// Удаляем задачу по id
	case "DELETE":

		log.Println("Вызван обработчик ManageTaskHandler с DELETE-запросом")
		// Получаем ID задачи из запроса
		taskID := req.URL.Query().Get("id")
		if taskID == "" {
			http.Error(res, `{"error": "Не указан ID задачи"}`, http.StatusBadRequest)
			return
		}

		// Удаляем задачу из БД
		result, err := db.Exec("DELETE FROM scheduler WHERE id = ?", taskID)
		if err != nil {
			http.Error(res, `{"error": "Ошибка удаления задачи"}`, http.StatusInternalServerError)
			return
		}

		// Проверяем, была ли задача удалена
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(res, `{"error": "Ошибка проверки удаления"}`, http.StatusInternalServerError)
			return
		}

		if rowsAffected == 0 {
			http.Error(res, `{"error": "Задача не найдена"}`, http.StatusNotFound)
			return
		}

		// Отправляем успешный ответ
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		if _, err := res.Write([]byte(`{}`)); err != nil {
			http.Error(res, `{"error": "Ошибка записи в ответ"}`, http.StatusInternalServerError)
		}
	}
}

// GetTaskHandler возвращает задачи из БД
func GetTaskHandler(res http.ResponseWriter, req *http.Request, db *sql.DB) {
	log.Println("Вызван обработчик GetTaskHandler с GET-запросом")
	// Проверяем метод запроса
	if req.Method != "GET" {
		http.Error(res, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}
	// Получаем параметры search из строки запроса
	search := req.URL.Query().Get("search")
	var rows *sql.Rows

	// Проверяем, является ли "search" датой в формате 02.01.2006
	parseDate, err := time.Parse("02.01.2006", search)
	if err == nil {
		// Если это дата, преобразуем её в формат YYYYMMDD (так хранятся даты в БД)
		searchDate := parseDate.Format("20060102")
		rows, err = db.Query("SELECT * FROM scheduler WHERE date = ? ORDER BY date ASC LIMIT ?", searchDate, 50)
	} else if search != "" {
		// Если это не дата, ищем в title и comment (регистр не учитывается)
		searchPattern := "%" + search + "%"
		rows, err = db.Query("SELECT * FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date ASC LIMIT ?", searchPattern, searchPattern, 50)
	} else {
		// Если параметр поиска отсутствует, возвращаем все задачи
		rows, err = db.Query("SELECT * FROM scheduler ORDER BY date ASC LIMIT ?", 50)
	}

	// Обрабатываем ошибки запроса
	if err != nil {
		http.Error(res, fmt.Sprintf(`{"error": "Ошибка при получении задач: %v"}`, err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task

	// Читаем строки из базы данных и добавляем их в список tasks
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			http.Error(res, `{"error": "Ошибка при обработке данных"}`, http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}
	// Проверяем на ошибки после завершения чтения строк
	if err := rows.Err(); err != nil {
		http.Error(res, fmt.Sprintf(`{"error": "Ошибка чтения данных: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Если tasks == null
	if len(tasks) == 0 {
		response := map[string]interface{}{
			"tasks": []Task{},
		}
		RespondJSON(res, response)
		return
	}

	// Формируем JSON с нужным форматом
	response := map[string]interface{}{
		"tasks": tasks,
	}
	// Преобразуем в формат JSON
	RespondJSON(res, response)
}

// DoneTaskHandler отмечает задачу выполненной (удаляет или переносит на следующую дату)
func DoneTaskHandler(res http.ResponseWriter, req *http.Request, db *sql.DB) {
	log.Println("Вызван обработчик DoneTaskHandler с POST-запросом")
	// Проверяем метод запроса
	if req.Method != http.MethodPost {
		http.Error(res, `{"error": "Метод не поддерживается"}`, http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID задачи из запроса
	taskID := req.URL.Query().Get("id")
	if taskID == "" {
		http.Error(res, `{"error": "Не указан ID задачи"}`, http.StatusBadRequest)
		return
	}

	// Получаем информацию о задаче
	var task Task
	row := db.QueryRow("SELECT id, date, repeat FROM scheduler WHERE id = ?", taskID)
	err := row.Scan(&task.ID, &task.Date, &task.Repeat)

	if err == sql.ErrNoRows {
		http.Error(res, `{"error": "Задача не найдена"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(res, `{"error": "Ошибка при получении данных"}`, http.StatusInternalServerError)
		return
	}

	// Если у задачи нет правила повторения, удаляем её
	if task.Repeat == "" {
		_, err = db.Exec("DELETE FROM scheduler WHERE id = ?", taskID)
		if err != nil {
			http.Error(res, `{"error": "Ошибка удаления задачи"}`, http.StatusInternalServerError)
			return
		}
	} else {
		// Вычисляем новую дату для повторяющейся задачи
		nextDate, err := repeat.RulesNextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			http.Error(res, `{"error": "Ошибка вычисления следующей даты"}`, http.StatusBadRequest)
			return
		}

		// Обновляем дату задачи в БД
		_, err = db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, taskID)
		if err != nil {
			http.Error(res, `{"error": "Ошибка обновления даты задачи"}`, http.StatusInternalServerError)
			return
		}
	}

	// Отправляем успешный ответ
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	if _, err := res.Write([]byte(`{}`)); err != nil {
		http.Error(res, `{"error": "Ошибка записи в ответ"}`, http.StatusInternalServerError)
	}

}
