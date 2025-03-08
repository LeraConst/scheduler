package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/LeraConst/scheduler/internal/repeat"
)

// NextDateHandler вычисляет следующую дату
func NextDateHandler(res http.ResponseWriter, req *http.Request) {
	now := req.URL.Query().Get("now")
	date := req.URL.Query().Get("date")
	repeatStr := req.URL.Query().Get("repeat")

	if now == "" || date == "" {
		http.Error(res, "отсутствуют обязательные параметры: now и date", http.StatusBadRequest)
		return
	}

	nowDate, err := time.Parse("20060102", now)
	if err != nil {
		http.Error(res, fmt.Sprintf("неверный формат даты: %v", err), http.StatusBadRequest)
	}

	nextDate, err := repeat.RulesNextDate(nowDate, date, repeatStr)
	if err != nil {
		http.Error(res, fmt.Sprintf("ошибка при расчете даты: %v", err), http.StatusInternalServerError)
	}
	if _, err := res.Write([]byte(nextDate)); err != nil {
		http.Error(res, `{"error": "Ошибка записи в ответ"}`, http.StatusInternalServerError)
	}
}
