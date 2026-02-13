.PHONY: dev build generate migrate css setup test lint clean

dev:
	air

build: generate css
	go build -o bin/server ./cmd/server

generate:
	sqlc generate -f sqlc/sqlc.yaml
	templ generate

migrate:
	goose -dir internal/database/migrations sqlite3 $(DATABASE_URL) up

migrate-down:
	goose -dir internal/database/migrations sqlite3 $(DATABASE_URL) down

css:
	npm run css:build

css-watch:
	npm run css:watch

setup:
	go mod tidy
	npm install
	mkdir -p data tmp

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin tmp static/css/output.css
