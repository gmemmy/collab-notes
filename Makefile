ENV_FILE=.env
include $(ENV_FILE)
export $(shell sed 's/=.*//' $(ENV_FILE))

.PHONY: lint format migrate dropdb connect run build test

migrate: ## Run migrations
	mysql -u $(MYSQL_USER) -p$(MYSQL_PASSWORD) -h $(MYSQL_HOST) -P $(MYSQL_PORT) --protocol=TCP $(MYSQL_DATABASE) < internal/db/migrations.sql

dropdb: ## Drop and recreate the dev database
	@MYSQL_PWD=$(MYSQL_PASSWORD) mysql -u $(MYSQL_USER) -h $(MYSQL_HOST) -P $(MYSQL_PORT) --protocol=TCP -e "DROP DATABASE IF EXISTS \`$(DB_NAME)\`; CREATE DATABASE \`$(DB_NAME)\`;"

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

test: ## Run tests
	go test -v ./...
