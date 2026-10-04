.PHONY: install generate run-bff run-web run

install:
	@echo "Installing frontend dependencies..."
	cd web && npm install
	@echo "Downloading backend dependencies..."
	cd bff && go mod download

generate:
	@echo "Generating GraphQL backend models..."
	cd bff && go run github.com/99designs/gqlgen generate
	@echo "Generating genql frontend client..."
	cd web && npx genql --schema ../bff/graph/schema.graphqls --output ./src/generated

run-bff:
	@echo "Starting BFF server..."
	cd bff/cmd/server && go run main.go

run-portfolio:
	@echo "Starting Portfolio API..."
	cd services/portfolio-api/cmd/server && go run main.go

run-web:
	@echo "Starting Web frontend..."
	cd web && npm run dev

run:
	@echo "Starting services..."
	$(MAKE) -j 3 run-portfolio run-bff run-web
