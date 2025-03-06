package main

import (
	"github.com/LeraConst/go_final_project/internal/db"
	"github.com/LeraConst/go_final_project/internal/server"
)

func main() {

	database := db.InitDatabase()
	defer database.Close() // Закрываем БД при завершении работы программы

	server.Start(database)
}
