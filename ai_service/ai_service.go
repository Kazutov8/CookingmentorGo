package ai_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AIClient управляет взаимодействием с внешней нейросетью.
type AIClient struct {
	APIURL string
}

// NewAIClient создает новый клиент для работы с AI API.
func NewAIClient(apiURL string) *AIClient {
	return &AIClient{
		APIURL: apiURL,
	}
}

// Call sends a request to the external neural network and returns the response text.
func (c *AIClient) Call(prompt string) (string, error) {
	systemPrompt := `Ты — дружелюбный и очень компетентный учитель по кулинарии и выпечке. Твоя задача — отвечать на вопросы пользователя только о рецептах, техниках приготовления еды и выпечки. Если пользователь спрашивает о чем-то другом (например, истории, политике, науке), ты должен вежливо прервать его и напомнить, что твоя специализация — это кулинария, и предложить помочь ему с рецептом или советом по готовке.`

	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
		"stream":      false,
		"temperature": 0.7,
		"max_tokens":  5000,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("ошибка маршалинга JSON: %w", err)
	}

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Post(c.APIURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("не удалось подключиться к AI API (%s): %w", c.APIURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AI API вернул ошибку %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var aiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&aiResponse); err != nil {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ошибка декодирования ответа: %w. Тело ответа было: %s", err, string(bodyBytes))
	}

	if len(aiResponse.Choices) == 0 {
		return "", fmt.Errorf("пустой ответ от нейросети")
	}
	return aiResponse.Choices[0].Message.Content, nil
}
