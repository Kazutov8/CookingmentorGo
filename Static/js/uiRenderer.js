// uiRenderer.js

const chatHistoryDiv = document.getElementById('chatHistory');
const sessionListDiv = document.getElementById('sessionList');
const chatTitleHeader = document.getElementById('chatTitleHeader');
const currentSessionIdDisplay = document.getElementById('currentSessionID');

/**
 * Обновляет заголовок чата и ID сессии в UI.
 * @param {string} title - Заголовок сессии.
 * @param {string} sessionID - ID сессии.
 */
export function updateChatHeader(title, sessionID) {
    chatTitleHeader.textContent = `Чат: "${title}"`;
    currentSessionIdDisplay.textContent = `Сессия ID: ${sessionID}`;
}

/**
 * Отрисовывает историю чата в DOM.
 * @param {string} rawHistory - Сырой лог истории от сервера.
 */
export function renderChatHistory(rawHistory) {
    // 1. Удаляем временные сообщения, если они есть (важно для обновления)
    chatHistoryDiv.querySelectorAll('[data-optimistic]').forEach(el => el.remove());

    let htmlContent = '';
    const blocks = rawHistory.split(/\n\s*\n/).filter(block => block.trim() !== '');

    blocks.forEach(block => {
        if (!block.includes("System: Welcome") && block.trim()) {
            let messageHtml = '';
            // Логика парсинга остается прежней, но теперь она чище и изолирована
            if (block.includes("You:")) {
                const userText = block.substring(5).trim();
                messageHtml += `<div class="history-message"><span class="user-bubble">${userText}</span></div>`;
            } else if (block.includes("Мастер готовки:")) {
                const aiText = block.substring(15).trim(); 
                messageHtml += `<div class="history-message"><span class="ai-bubble">Мастер готовки: ${aiText}</span></div>`;
            } else if (block.startsWith("System:") || block.startsWith("Admin:")) {
                messageHtml += `<div class="history-message system-message">${block}</div>`;
            } else {
                messageHtml += `<div class="history-message">${block}</div>`;
            }

            if (messageHtml) {
                htmlContent += messageHtml;
            }
        }
    });

    chatHistoryDiv.innerHTML = htmlContent; 
}


/**
 * Отображает временный блок сообщения пользователя ДО запроса к серверу.
 * @param {string} text - Текст сообщения.
 */
export function displayOptimisticMessage(text) {
    const msgElement = document.createElement('div');
    msgElement.classList.add('history-message');
    msgElement.setAttribute('data-optimistic', 'true'); 
    msgElement.innerHTML = `<span class="user-bubble">Вы: ${text}</span>`;
    chatHistoryDiv.appendChild(msgElement);
}

/**
 * Обновляет сайдбар, удаляя старые элементы и добавляя новые сессии.
 * @param {Array} sessions - Массив объектов сессий.
 */
export function updateSidebar(sessions) {
    sessionListDiv.innerHTML = ''; 
    if (sessions.length === 0) {
        sessionListDiv.innerHTML = '<p style="color: #888;">Нет сохраненных сессий. Начните чат!</p>';
        return null; // Возвращаем null, если нет активной сессии
    }

    let firstSessionID = null;
    const sessionElements = [];

    sessions.forEach(s => {
        const [idStr, title] = s.split(': ').map(s => s.trim());
        const sessionItem = document.createElement('div');
        sessionItem.classList.add('session-item');
        sessionItem.dataset.sessionId = idStr;
        sessionItem.textContent = title || `Сессия ${idStr}`;
        sessionListDiv.appendChild(sessionItem);
        sessionElements.push(sessionItem);

        if (!firstSessionID) {
            firstSessionID = idStr;
        }
    });
    return firstSessionID; // Возвращаем ID первой сессии для автоматической активации
}

/**
 * Устанавливает активный элемент в сайдбаре.
 * @param {string} sessionID - ID элемента, который должен стать активным.
 */
export function setActiveSessionElement(sessionID) {
    // Сначала убираем класс 'active' со всех элементов
    document.querySelectorAll('.session-item').forEach(el => el.classList.remove('active'));

    // Находим и добавляем класс 'active' к нужному элементу
    const activeElement = document.querySelector(`.session-item[data-session-id="${sessionID}"]`);
    if (activeElement) {
        activeElement.classList.add('active');
    }
}

/**
 * Прокручивает чат в самый низ.
 */
export function scrollToBottom() {
    chatHistoryDiv.scrollTop = chatHistoryDiv.scrollHeight;
}
