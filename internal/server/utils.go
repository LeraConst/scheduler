package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RespondJSON выполняет сериализацию в JSON
func RespondJSON(res http.ResponseWriter, data interface{}) {
	out, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		http.Error(res, fmt.Sprintf(`{"error": "Ошибка сериализации в JSON: %v"}`, err), http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	if _, err := res.Write(out); err != nil {
		http.Error(res, `{"error": "Ошибка записи в ответ"}`, http.StatusInternalServerError)
	}
}
