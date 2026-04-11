.PHONY: run test swag up down tidy

# Запуск проекта локально
run:
	go run .

# Запуск всех тестов с подробным выводом
test:
	go test -v ./...

# Обновление документации Swagger
swag:
	swag init

# Поднять базу и сервер в Docker (в фоновом режиме)
up:
	docker-compose up --build -d

# Остановить и удалить контейнеры Docker
down:
	docker-compose down

# Причесать зависимости (удалить лишнее, скачать нужное)
tidy:
	go mod tidy