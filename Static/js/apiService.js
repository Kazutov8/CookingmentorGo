// apiService.js

/**
 * Загружает все сессии из API.
 * @returns {Promise<Array>} Массив объектов сессий.
 */
export async function loadSessions() {
    try {
        const response = await fetch('/api/sessions');
        if (!response.ok) throw new Error('Ошибка загрузки списка сессий.');
        const data = await response.json();
        return data.sessions;
    } catch (error) {
        console.error("API Error: loadSessions", error);
        throw error; // Перебрасываем ошибку, чтобы контроллер мог ее обработать
    }
}

/**
 * Загружает историю чата по заданному ID сессии. 
 * @param {string} sessionID - ID активной сессии.
 * @returns {Promise<{history: string, title: string}>} История и заголовок.
 */
export async function loadHistory(sessionID) {
    try {
        const response = await fetch(`/api/history/${sessionID}`);
        if (!response.ok) throw new Error('Ошибка загрузки истории.');

        const data = await response.json();
        return { history: data.history, title: data.title || `Сессия ${sessionID}` };
    } catch (error) {
        console.error("API Error: loadHistory", error);
        throw error;
    }
}

/**
 * Отправляет сообщение на сервер и ждет полного лога истории с ответом AI. 
 * @param {string} messageText - Текст сообщения пользователя.
 * @param {string} sessionID - ID активной сессии.
 * @returns {Promise<string>} Полный сырой лог истории (включая ответ AI).
 */
export async function sendMessageToServer(messageText, sessionID) {
    try {
        const response = await fetch(`/api/send/${sessionID}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ message: messageText })
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Ошибка сервера при отправке сообщения.');
        }

        const data = await response.json();
        return data.history; // Возвращаем сырой лог истории для рендеринга
    } catch (error) {
        console.error("API Error: sendMessageToServer", error);
        throw error;
    }
}

/**
 * Создает новую сессию и возвращает ее ID. 
 * В реальном приложении это должно быть отдельное API-вычисление.
 */
export function generateNewSessionID() {
    return Math.floor(Math.random() * 10000);
}
