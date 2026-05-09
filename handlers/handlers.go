package handlers

import (
	"fmt"
	"gin-app/ai_service"
	"gin-app/database"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AppHandlers struct {
	dbManager *database.DBManager
	aiClient  *ai_service.AIClient
}

var handlersInstance *AppHandlers

// InitHandlers инициализирует все маршруты
func InitHandlers(r *gin.Engine, dbManager *database.DBManager) {
	const aiAPIURL = "http://localhost:1234/v1/chat/completions"
	aiClient := ai_service.NewAIClient(aiAPIURL)

	handlersInstance = &AppHandlers{
		dbManager: dbManager,
		aiClient:  aiClient,
	}

	setupStaticFiles(r)
	setupAPISessions(r)
}

func setupStaticFiles(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.File("Static/index.html")
	})
	r.Static("/styles", "./Static/styles")
	r.Static("/js", "./Static/js")
}

func setupAPISessions(r *gin.Engine) {
	// API для получения ВСЕХ сессий
	r.GET("/api/sessions", func(c *gin.Context) {
		sessions, err := handlersInstance.dbManager.GetAllSessionTitles()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список сессий"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"sessions": sessions})
	})

	// API для получения истории ЧАТА по ID сессии
	r.GET("/api/history/:sessionID", func(c *gin.Context) {
		sessionID := c.Param("sessionID")
		title, errTitle := handlersInstance.dbManager.GetSessionTitleFromDB(sessionID)
		if errTitle != nil {
			fmt.Printf("Warning: Could not fetch title for %s: %v\n", sessionID, errTitle)
		}

		history, err := handlersInstance.dbManager.GetChatHistoryFromDB(sessionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить историю чата"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"history": history,
			"title":   title,
		})
	})

	// API для отправки нового сообщения
	r.POST("/api/send/:sessionID", func(c *gin.Context) {
		handleSendMessage(c)
	})
}

func handleSendMessage(c *gin.Context) {
	sessionID := c.Param("sessionID")
	var message struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	// 1. Получаем текущую историю
	currentHistory, err := handlersInstance.dbManager.GetChatHistoryFromDB(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Критическая ошибка при чтении истории"})
		return
	}

	// 2. Формируем контекст для нейросети
	userMessageFormatted := fmt.Sprintf("You: %s", message.Message)

	var prompt string
	if currentHistory == "" || strings.Contains(currentHistory, "Привет. Тут мы учимся готовить еду") {
		// Это первое сообщение пользователя или почти новая сессия
		prompt = userMessageFormatted
	} else {
		prompt = fmt.Sprintf("%s\n\n%s", currentHistory, userMessageFormatted)
	}

	// 3. Вызов нейросети уже с историей и текущим сообщением
	aiResponseText, err := handlersInstance.aiClient.Call(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ошибка связи с AI: %v", err)})
		return
	}

	// 4. Форматируем ответ нейросети
	aiResponseFormatted := "Мастер готовки: " + aiResponseText

	// 5. Формируем новую историю
	var newHistory string
	if currentHistory == "" || strings.Contains(currentHistory, "Привет. Тут мы учимся готовить еду") {
		newHistory = userMessageFormatted + "\n\n" + aiResponseFormatted
	} else {
		newHistory = fmt.Sprintf("%s\n\n%s\n\n%s", currentHistory, userMessageFormatted, aiResponseFormatted)
	}

	// 6. Сохраняем в БД
	err = handlersInstance.dbManager.SaveChatHistoryToDB(sessionID, newHistory, message.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить сообщение в БД", "details": err.Error()})
		return
	}

	// 7. Отправляем ответ клиенту
	title, _ := handlersInstance.dbManager.GetSessionTitleFromDB(sessionID)
	c.JSON(http.StatusOK, gin.H{"history": newHistory, "title": title})
}
