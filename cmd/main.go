package main

import (
	"github.com/LeraConst/scheduler/internal/db"
	"github.com/LeraConst/scheduler/internal/server"
)

func main() {

	database := db.InitDatabase()
	defer database.Close() // Закрываем БД при завершении работы программы

	server.Start(database)
}
