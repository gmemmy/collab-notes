ENV_FILE=.env
include $(ENV_FILE)
export $(shell sed 's/=.*//' $(ENV_FILE))

.PHONY: lint format migrate dropdb connect run build

migrate: ## Run migrations
	mysql -u $(MYSQL_USER) -p$(MYSQL_PASSWORD) -h $(MYSQL_HOST) -P $(MYSQL_PORT) --protocol=TCP $(MYSQL_DATABASE) < internal/db/migrations.sql

dropdb: ## Drop and recreate the dev database
	mysql -u $(MYSQL_USER) -p$(MYSQL_PASSWORD) -h $(MYSQL_HOST) -e "DROP DATABASE IF EXISTS $(DB_NAME); CREATE DATABASE $(DB_NAME);"

connect: ## Open MySQL CLI session using .env credentials
	mysql -u $(MYSQL_USER) -p$(MYSQL_PASSWORD) -h $(MYSQL_HOST) -P $(MYSQL_PORT) --protocol=TCP $(MYSQL_DATABASE)

run: ## Run the Go app
	go run cmd/main.go

build: ## Build the Go binary
	go build -o bin/app ./cmd/main.go

lint: ## Run linting
	golangci-lint run

format: ## Format the code
	gofumpt -w .