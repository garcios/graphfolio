-include .env
export

PG_ADMIN_URL     ?= postgres://$(USER)@localhost:5432/postgres
DB_NAME          ?= graphfolio
PORTFOLIO_DB_URL ?= postgres://portfolio_svc:portfolio@localhost:5432/$(DB_NAME)?sslmode=disable
USER_DB_URL      ?= postgres://user_svc:user@localhost:5432/$(DB_NAME)?sslmode=disable

.PHONY: install generate proto run-bff run-portfolio run-web run \
	db-check db-bootstrap db-drop db-reset db-test-setup \
	migrate-up migrate-down migrate-create db-seed test

install:
	@echo "Installing frontend dependencies..."
	cd web && npm install
	@echo "Downloading backend dependencies..."
	cd bff && go mod download
	cd pkg && go mod download
	cd proto && go mod download
	cd services/portfolio-api && go mod download
	cd services/user-api && go mod download

proto:
	@echo "Generating Protocol Buffers code..."
	cd proto && protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		common/v1/decimal.proto portfolio/v1/portfolio.proto

generate: proto
	@echo "Generating GraphQL backend models..."
	cd bff && go run github.com/99designs/gqlgen generate
	@echo "Generating genql frontend client..."
	cd web && npx genql --schema ../bff/graph/schema.graphqls --output ./src/generated
	@echo "Generating Go mocks..."
	cd services/portfolio-api && go generate ./...

db-check:
	@pg_isready -h localhost -p 5432

db-bootstrap: db-check
	psql "$(PG_ADMIN_URL)" -v db=$(DB_NAME) -f scripts/db/bootstrap.sql

db-drop: db-check
	psql "$(PG_ADMIN_URL)" -v db=$(DB_NAME) -f scripts/db/teardown.sql

migrate-up:
	migrate -path services/portfolio-api/migrations -database "$(PORTFOLIO_DB_URL)&search_path=portfolio&x-migrations-table=schema_migrations" up
	migrate -path services/user-api/migrations      -database "$(USER_DB_URL)&search_path=users&x-migrations-table=schema_migrations" up

migrate-down:
	migrate -path services/portfolio-api/migrations -database "$(PORTFOLIO_DB_URL)&search_path=portfolio&x-migrations-table=schema_migrations" down 1
	migrate -path services/user-api/migrations      -database "$(USER_DB_URL)&search_path=users&x-migrations-table=schema_migrations" down 1

migrate-create:
	@test -n "$(svc)" || (echo "Usage: make migrate-create svc=portfolio-api name=add_x" && exit 1)
	@test -n "$(name)" || (echo "Usage: make migrate-create svc=portfolio-api name=add_x" && exit 1)
	migrate create -ext sql -dir services/$(svc)/migrations -seq $(name)

db-seed:
	psql "$(PORTFOLIO_DB_URL)" -f services/portfolio-api/seeds/dev_seed.sql

db-reset: db-drop db-bootstrap migrate-up db-seed

db-test-setup:
	$(MAKE) db-drop db-bootstrap DB_NAME=graphfolio_test
	migrate -path services/portfolio-api/migrations -database "$(TEST_PORTFOLIO_DB_URL)&search_path=portfolio&x-migrations-table=schema_migrations" up
	migrate -path services/user-api/migrations      -database "$(TEST_USER_DB_URL)&search_path=users&x-migrations-table=schema_migrations" up

test:
	go test ./bff/... ./pkg/... ./services/portfolio-api/... ./services/user-api/...

run-bff:
	@echo "Starting BFF server..."
	cd bff/cmd/server && go run main.go

run-portfolio:
	@echo "Starting Portfolio API..."
	cd services/portfolio-api/cmd/server && go run main.go

run-user:
	@echo "Starting User API..."
	cd services/user-api/cmd/server && go run main.go

run-web:
	@echo "Starting Web frontend..."
	cd web && npm run dev

run:
	@echo "Starting services..."
	$(MAKE) -j 3 run-portfolio run-bff run-web

