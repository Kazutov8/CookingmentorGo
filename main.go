package main

import (
	"fmt"
	"gin-app/database"
	"gin-app/handlers"

	"github.com/gin-gonic/gin"
)

var dbManager *database.DBManager

func main() {
	// Инициализация базы данных
	dbPath := "./chat_sessions.db"
	var err error
	dbManager, err = database.NewDBManager(dbPath)
	if err != nil {
		fmt.Println("КРИТИЧЕСКАЯ ОШИБКА ПРИ ИНИЦИАЛИЗАЦИИ БД:", err)
		return
	}
	defer dbManager.Close()

	r := gin.Default()

	// Инициализация обработчиков с зависимостями
	handlers.InitHandlers(r, dbManager)

	fmt.Println("Сервер запущен на http://localhost:8080")
	r.Run(":8080")
}
