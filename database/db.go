package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "modernc.org/sqlite" // SQLite driver
)

// DBManager управляет подключением и операциями с базой данных.
type DBManager struct {
	DB *sql.DB
}

// NewDBManager инициализирует соединение с БД и создает необходимые таблицы.
func NewDBManager(dbPath string) (*DBManager, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	// Проверка соединения
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("не удалось пинговать базу данных: %w", err)
	}

	manager := &DBManager{DB: db}

	// Создание таблицы (если не существует)
	if err := manager.createTable(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ошибка создания таблиц: %w", err)
	}

	log.Println("Успешно подключено к базе данных и проверены таблицы.")
	return manager, nil
}

// Close закрывает соединение с базой данных.
func (m *DBManager) Close() error {
	return m.DB.Close()
}

// createTable выполняет первоначальную настройку базы данных и создает таблицу
func (m *DBManager) createTable() error {
	_, err := m.DB.Exec(`
		CREATE TABLE IF NOT EXISTS Sessions (
			id TEXT PRIMARY KEY, 
			title TEXT, 
			text TEXT
		);
	`)
	return err
}

// GetSessionTitleFromDB извлекает заголовок для конкретной сессии
func (m *DBManager) GetSessionTitleFromDB(sessionID string) (string, error) {
	row := m.DB.QueryRow("SELECT title FROM Sessions WHERE id = ?", sessionID)
	var title sql.NullString
	err := row.Scan(&title)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("сессия не найдена")
	} else if err != nil {
		return "", fmt.Errorf("ошибка чтения заголовка: %w", err)
	}
	if title.Valid {
		return title.String, nil
	}
	return "Нет заголовка", nil
}

// GetChatHistoryFromDB извлекает историю чата по ID сессии.
func (m *DBManager) GetChatHistoryFromDB(sessionID string) (string, error) {
	row := m.DB.QueryRow("SELECT text FROM Sessions WHERE id = ?", sessionID)
	var history sql.NullString
	err := row.Scan(&history)
	if err == sql.ErrNoRows {
		// Сессия существует по ID, но не заполнена, возвращаем стартовое приветствие
		return "Привет. Я тут чтобы помочь тебе научиться готовить", nil
	} else if err != nil {
		return "", fmt.Errorf("ошибка чтения истории из БД для сессии %s: %w", sessionID, err)
	}
	if !history.Valid || history.String == "" {
		// Если запись существует, но поле 'text' пустое
		return "Привет. Я тут чтобы помочь тебе научиться готовить", nil
	}
	return history.String, nil
}

// GetAllSessionTitles получает список всех активных ID и заголовков сессий
func (m *DBManager) GetAllSessionTitles() ([]string, error) {
	rows, err := m.DB.Query("SELECT id, title FROM Sessions ORDER BY CAST(id AS INTEGER) ASC") // Сортируем по числу ID
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса сессий: %w", err)
	}
	defer rows.Close()
	var titles []string
	for rows.Next() {
		var id string
		var title sql.NullString
		if err := rows.Scan(&id, &title); err != nil {
			return nil, fmt.Errorf("ошибка сканирования сессии: %w", err)
		}
		displayTitle := "Нет заголовка"
		if title.Valid && title.String != "" {
			displayTitle = title.String
		}
		// Формат: ID: Заголовок
		titles = append(titles, fmt.Sprintf("%s: %s", id, displayTitle))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по сессиям: %w", err)
	}
	return titles, nil
}

// SaveChatHistoryToDB обновляет поля Title и Text в таблице Sessions для заданной ID сессии.
func (m *DBManager) SaveChatHistoryToDB(sessionID string, newText string, newMessageContent string) error {
	// 1. Получаем существующую историю из БД
	var existingTitle string
	var existingText string
	err := m.DB.QueryRow("SELECT title, text FROM Sessions WHERE id = ?", sessionID).Scan(&existingTitle, &existingText)

	var title string

	if err != nil {
		if err == sql.ErrNoRows {
			// Сессия не существует - создаем без заголовка (будет установлен при первом "You:")
			title = "" // Временный заголовок
		} else {
			return fmt.Errorf("ошибка проверки сессии: %w", err)
		}
	} else {
		// Сессия существует - берем существующий заголовок
		title = existingTitle
	}

	// Подсчитываем количество "You:" в существующей и новой истории
	existingYouCount := strings.Count(existingText, "You:")
	newYouCount := strings.Count(newText, "You:")

	// Присваиваем заголовок только когда появляется ВТОРОЕ сообщение с "You:"
	if existingYouCount < 2 && newYouCount == 2 && title == "" {
		// Берем содержимое ПОСЛЕДНЕГО сообщения пользователя для заголовка
		// (или можно взять newMessageContent как сейчас)
		title = newMessageContent
		if len(newMessageContent) > 50 {
			runes := []rune(newMessageContent)
			if len(runes) > 50 {
				title = string(runes[:50]) + "..."
			}
		}
	}

	// 3. Сохраняем сессию
	stmt, err := m.DB.Prepare(`
		INSERT OR REPLACE INTO Sessions (id, title, text) 
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("ошибка подготовки запроса: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(sessionID, title, newText)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	return nil
}

// GetDBInstance возвращает сам *sql.DB для внешних нужд (например, в main.go)
func (m *DBManager) GetDBInstance() *sql.DB {
	return m.DB
}
