# To-Do App (Go + Postgres + Docker)

Простое и эффективное приложение для управления списком задач с использованием слоистой архитектуры (Clean Architecture).

## 🚀 Стек технологий
- **Language:** Go 1.21
- **Database:** PostgreSQL (Migrations: Goose)
- **API:** REST (HTTP) + gRPC Client
- **Architecture:** Clean Architecture (Handlers -> Service -> Repository)
- **Logs:** slog (Structured logging)
- **Testing:** Testify (Unit + Mocks)
- **Frontend:** Vanilla JavaScript, HTML5, CSS3 (раздается встроенным HTTP-сервером Go)
- **Infrastructure:** Docker, Docker Compose, Makefile
- **Configuration:** Cleanenv (.env)

## Архитектурные особенности
- [x] Полная изоляция слоев (Clean Architecture principles)
- [x] Автоматическое ожидание доступности БД при старте (Retry logic)
- [x] Безопасная остановка сервера (Graceful Shutdown)
- [x] Управление проектом через `Makefile`
- [x] Unit-тестирование хендлеров через моки

## Межсервисное взаимодействие
Сервис интегрирован с `Notifier Service` через **gRPC**. При создании задачи ToDo-сервис отправляет асинхронное уведомление (через Goroutines), не блокируя основной поток выполнения.

## Быстрый запуск
Приложение полностью контейнеризировано.

### Предварительные требования
   ```bash
   # Запуск всей инфраструктуры (включая БД и Notifier)
   docker-compose up --build
   ```
Приложение будет доступно по адресу: http://localhost:8080

## API документация
Все ответы API имеют унифицированную структуру:{"data": <payload>, "error": "<message>"}

*   **Метод	Путь	Описание**
*   **GET**	/tasks	Получить список всех задач
*   **POST**	/tasks	Создать новую задачу
*   **PATCH**	/tasks/{id}	Изменить статус выполнения (done)
*   **DELETE**	/tasks/{id}	Удалить задачу из базы
*   **GET**	/health	Проверка работоспособности API

## Миграции
Для управления схемой БД используется инструмент **Goose**.
Миграции применяются автоматически при старте или вручную:
   ```bash
   goose -dir migrations postgres "user=... dbname=..." up
   ```

## Тестирование
Бизнес-логика (слой Service) покрыта Unit-тестами с использованием моков:
   ```bash
   go test ./internal/service/...
   ```

## 📄 Лицензия
Этот проект создан в целях исследования программного продукта и доступен для свободного использования.
