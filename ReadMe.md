# To-Do App (Go + Postgres + Docker)

Простое и эффективное приложение для управления списком задач с использованием слоистой архитектуры (Clean Architecture).

## 🚀 Стек технологий
- **Language:** Go 1.21
- **Database:** PostgreSQL
- **Migration:** Goose
- **Logs:** slog (Structured logging)
- **Tests:** Testify (Unit + Mocks)
- **Frontend:** Vanilla JavaScript, HTML5, CSS3 (раздается встроенным HTTP-сервером Go)
- **Infrastructure:** Docker, Docker Compose, Makefile

## Особенности
- [x] Полная изоляция слоев (Clean Architecture principles)
- [x] Автоматическое ожидание доступности БД при старте (Retry logic)
- [x] Безопасная остановка сервера (Graceful Shutdown)
- [x] Управление проектом через `Makefile`
- [x] Unit-тестирование хендлеров через моки

## ⚙️ Как запустить приложение

### Предварительные требования
Убедитесь, что у вас установлены:
*   **Docker**
*   **Docker Compose**

## Быстрый старт

1. Склонируйте репозиторий.
2. Создайте `.env` файл на основе `.env.example`.
3. Запустите проект:
   ```bash
   make dc-up
4. Приложение будет доступно по адресу: http://localhost:8080

## 📡 API документация
Все ответы API имеют унифицированную структуру:{"data": <payload>, "error": "<message>"}

*   **Метод	Путь	Описание**
*   **GET**	/tasks	Получить список всех задач
*   **POST**	/tasks	Создать новую задачу
*   **PATCH**	/tasks/{id}	Изменить статус выполнения (done)
*   **DELETE**	/tasks/{id}	Удалить задачу из базы
*   **GET**	/health	Проверка работоспособности API



## 📄 Лицензия
Этот проект создан в целях исследования программного продукта и доступен для свободного использования.
