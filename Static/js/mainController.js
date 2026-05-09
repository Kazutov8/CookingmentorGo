// mainController.js

import { loadSessions, loadHistory, sendMessageToServer, generateNewSessionID } from './apiService.js';
import * as UI from './uiRenderer.js'; // Импортируем все функции рендеринга

const messageForm = document.getElementById('messageForm');
const messageInput = document.getElementById('messageInput');
const sendButton = document.getElementById('sendButton');
const newSessionButton = document.getElementById('newSessionButton');


let activeSessionID = null;

/**
 * Инициализирует приложение: загружает сессии и активирует первую.
 */
async function initializeApp() {
    // 1. Назначаем обработчики событий
    messageForm.addEventListener('submit', handleSendMessage);
    newSessionButton.addEventListener('click', createNewSession);

    // 2. Устанавливаем начальное состояние textarea
    messageInput.style.height = 'auto'; 

    try {
        // 3. Загружаем сессии и получаем ID первой активной сессии
        const sessions = await loadSessions();
        const firstSessionID = UI.updateSidebar(sessions); // Обновляем сайдбар
        
        if (firstSessionID) {
            activeSessionID = firstSessionID; 
            await loadAndDisplayHistory(firstSessionID); // Автоматически загружаем историю первой сессии
        }

    } catch (error) {
        console.error("Failed to initialize application:", error);
        // Здесь можно показать пользователю общее сообщение об ошибке
    }
}


/**
 * Обрабатывает клик по элементу в сайдбаре.
 * @param {Event} e - Событие клика.
 */
async function handleSessionClick(e) {
    const target = e.target.closest('.session-item');
    if (!target) return;

    const clickedSessionID = target.dataset.sessionId;
    activeSessionID = clickedSessionID; 
    await loadAndDisplayHistory(clickedSessionID);
}


/**
 * Загружает и отображает историю чата по заданному ID сессии. (Универсальная функция)
 * @param {string} sessionID - ID сессии для загрузки.
 */
async function loadAndDisplayHistory(sessionID) {
    if (!sessionID) return;

    // 1. Обновляем UI: делаем элемент активным и меняем заголовки
    UI.setActiveSessionElement(sessionID);
    try {
        const { history, title } = await loadHistory(sessionID);
        UI.updateChatHeader(title, sessionID);
        // 2. Рендерим историю
        UI.renderChatHistory(history);
        UI.scrollToBottom();

    } catch (error) {
        console.error("Failed to load history:", error);
        UI.updateChatHeader("Общий Чат", sessionID);
        UI.renderChatHistory(`<p style="color: red;">Не удалось получить историю чата для сессии ${sessionID}. Попробуйте позже.</p>`);
    }
}


/**
 * Обрабатывает отправку сообщения пользователем.
 * @param {Event} e - Событие формы.
 */
async function handleSendMessage(e) {
    e.preventDefault();
    const messageText = messageInput.value.trim();
    if (!messageText || !activeSessionID) return;

    // 1. Оптимистичное отображение (показываем сообщение пользователя сразу)
    UI.displayOptimisticMessage(messageText);

    // Сброс поля ввода и подготовка UI к ожиданию ответа
    messageInput.value = ''; 
    messageInput.style.height = 'auto'; 

    sendButton.disabled = true;
    sendButton.textContent = 'Думает...';

    try {
        // 2. Отправляем запрос на сервер и ждем лог истории (включая ответ AI)
        const rawHistory = await sendMessageToServer(messageText, activeSessionID);

        // 3. Обновление истории: заменяем оптимистичное сообщение на полную историю
        UI.renderChatHistory(rawHistory);

    } catch (error) {
        console.error("Sending message failed:", error);
        alert(`Не удалось отправить сообщение или получить ответ от AI: ${error.message}`);
        // 4. Если ошибка произошла, удаляем оптимистично добавленное сообщение пользователя из UI
        const lastMsg = document.querySelector('[data-optimistic="true"]');
        if (lastMsg) lastMsg.remove();

    } finally {
        // Восстанавливаем состояние кнопки и скролл
        sendButton.disabled = false;
        sendButton.textContent = 'Отправить';
        UI.scrollToBottom(); 
    }
}


/**
 * Создает новую сессию (тестовая функция).
 */
async function createNewSession() {
    if (!confirm("Вы уверены, что хотите создать новую пустую тестовую сессию?")) return;

    const newID = generateNewSessionID();
    activeSessionID = newID;

    try {
        // Имитация отправки системного сообщения для записи в БД и получения лога
        await sendMessageToServer("Мы начинаем новый урок по приготовлению", newID); 

        // Обновляем UI
        UI.updateChatHeader(`Новая Сессия ${newID}`, newID);
        UI.renderChatHistory("Добро пожаловать! Я ваш учитель по приготовлению еды."); // Начальное сообщение
        UI.setActiveSessionElement(newID);

        // Добавляем новую запись в сайдбар (имитируем API)
        const newElement = document.createElement('div');
        newElement.classList.add('session-item', 'active');
        newElement.dataset.sessionId = newID; 
        newElement.textContent = `Новая сессия ${newID}`;
        document.getElementById('sessionList').prepend(newElement);

    } catch (e) {
        alert("Ошибка при создании новой тестовой сессии.");
    }
}


// --- ИНИЦИАЛИЗАЦИЯ ПРИ ЗАГРУЗКЕ СТРАНИЦЫ ---
window.onload = () => {
    // Добавляем обработчик клика на сайдбар после загрузки DOM
    document.getElementById('sessionList').addEventListener('click', handleSessionClick);
    
    initializeApp(); // Запускаем главный цикл приложения
};
