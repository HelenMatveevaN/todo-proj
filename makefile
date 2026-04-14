include .env
export

.PHONY: dc-up
dc-up: ## Запустить всё в Docker
	docker-compose up -d --build --remove-orphans

.PHONY: dc-down
dc-down: ## Остановить всё в Docker
	docker-compose down

clean: 	## Быстрая очистка всего
	docker-compose down --remove-orphans
	docker system prune -f

.PHONY: up
up: ## Применить миграции (с хоста на Docker-базу)
	goose -dir migrations postgres $(DB_URL) up

.PHONY: status
status: ## Проверить статус
	goose -dir migrations postgres $(DB_URL) status

.PHONY: create
create: ## Создать новую миграцию (usage: make create name=add_users)
	goose -dir migrations create $(name) sql

.PHONY: redo
redo: ## Перезапустить последнюю миграцию (down + up)
	goose -dir $migrations postgres $(DB_URL) redo

.PHONY: test
test: ## Запустить все тесты
	go test -v -cover ./...
	
.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/notifier.proto