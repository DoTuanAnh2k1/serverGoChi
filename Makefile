# MGT Service Makefile

APP_NAME    := mgt-service
COMPOSE     := docker compose -f deploy/docker/docker-compose.yml
MAIN        := ./cmd/main/main.go
IMPORT_CMD  := ./cmd/import
BUILD_DIR   := bin
SERVER_PORT := $(or $(SERVER_PORT),3000)

# DB connection (for dump)
DB_DRIVER   := $(or $(DB_DRIVER),mysql)
DB_HOST     := $(or $(MYSQL_HOST),localhost)
DB_PORT     := $(or $(MYSQL_PORT),3306)
DB_USER     := $(or $(MYSQL_USER),mgtuser)
DB_PASS     := $(or $(MYSQL_PASSWORD),mgtpassword)
DB_NAME     := $(or $(MYSQL_DB_NAME),cli_db)

.PHONY: help up down build build-docker import dump metric test clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

## ── Local environment ──────────────────────────────────────────────

up: ## Start local environment (DB containers + app)
	$(COMPOSE) up -d mysql mongodb postgres
	@echo "Waiting for databases..."
	@sleep 10
	@echo "Starting app..."
	@DB_DRIVER=$(DB_DRIVER) \
	MYSQL_HOST=$(DB_HOST) MYSQL_PORT=$(DB_PORT) \
	MYSQL_USER=$(DB_USER) MYSQL_PASSWORD=$(DB_PASS) MYSQL_DB_NAME=$(DB_NAME) \
	SERVER_HOST=0.0.0.0 SERVER_PORT=$(SERVER_PORT) \
	JWT_SECRET_KEY=$${JWT_SECRET_KEY:-dev-secret-key} \
	JWT_EXPIRY_HOURS=24 LOG_LEVEL=info DB_LOG_LEVEL=warn \
	CORS_ORIGINS="*" LEADER_ELECTION_ENABLED=false \
	TCP_LISTEN_PORT=3675 TCP_DATA_DIR=/tmp/mgt-subscribers \
	CLI_LOG_EXPORT_DIR=/tmp/mgt-csv CSV_EXPORT_HOUR=23 \
	PPROF_ENABLED=true PPROF_ADDR=:6060 \
	go run $(MAIN)

down: ## Stop local environment
	$(COMPOSE) down
	@-pkill -f "go run $(MAIN)" 2>/dev/null || true
	@-lsof -ti:$(SERVER_PORT) | xargs kill -9 2>/dev/null || true
	@-lsof -ti:3675 | xargs kill -9 2>/dev/null || true
	@echo "Done."

## ── Build ──────────────────────────────────────────────────────────

build: ## Build app binary
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)
	@echo "Built: $(BUILD_DIR)/$(APP_NAME)"

build-docker: ## Build Docker image
	docker build -f deploy/docker/Dockerfile -t $(APP_NAME):latest .

## ── Data ───────────────────────────────────────────────────────────

import: ## Import data from file: make import FILE=data.txt
	@if [ -z "$(FILE)" ]; then echo "Usage: make import FILE=<path>"; exit 1; fi
	DB_DRIVER=$(DB_DRIVER) \
	MYSQL_HOST=$(DB_HOST) MYSQL_PORT=$(DB_PORT) \
	MYSQL_USER=$(DB_USER) MYSQL_PASSWORD=$(DB_PASS) MYSQL_DB_NAME=$(DB_NAME) \
	LOG_LEVEL=error DB_LOG_LEVEL=error \
	go run $(IMPORT_CMD) -file $(FILE)

dump: ## Dump all data from database
	@echo "=== tbl_account ==="
	@docker exec mgt-mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT account_id, account_name, account_type, is_enable, created_by FROM tbl_account;" 2>/dev/null || \
		echo "(container mgt-mysql not running, trying direct connection)" && \
		mysql -h$(DB_HOST) -P$(DB_PORT) -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT account_id, account_name, account_type, is_enable, created_by FROM tbl_account;" 2>/dev/null
	@echo ""
	@echo "=== cli_ne ==="
	@docker exec mgt-mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT id, name, site_name, ip_address, port, system_type FROM cli_ne;" 2>/dev/null || \
		mysql -h$(DB_HOST) -P$(DB_PORT) -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT id, name, site_name, ip_address, port, system_type FROM cli_ne;" 2>/dev/null
	@echo ""
	@echo "=== cli_role ==="
	@docker exec mgt-mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT * FROM cli_role;" 2>/dev/null || \
		mysql -h$(DB_HOST) -P$(DB_PORT) -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT * FROM cli_role;" 2>/dev/null
	@echo ""
	@echo "=== cli_role_user_mapping ==="
	@docker exec mgt-mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT u.account_name, m.permission FROM cli_role_user_mapping m JOIN tbl_account u ON u.account_id=m.user_id;" 2>/dev/null || \
		mysql -h$(DB_HOST) -P$(DB_PORT) -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT u.account_name, m.permission FROM cli_role_user_mapping m JOIN tbl_account u ON u.account_id=m.user_id;" 2>/dev/null
	@echo ""
	@echo "=== cli_user_ne_mapping ==="
	@docker exec mgt-mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT u.account_name, n.name AS ne_name FROM cli_user_ne_mapping m JOIN tbl_account u ON u.account_id=m.user_id JOIN cli_ne n ON n.id=m.tbl_ne_id ORDER BY u.account_name, n.name;" 2>/dev/null || \
		mysql -h$(DB_HOST) -P$(DB_PORT) -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT u.account_name, n.name AS ne_name FROM cli_user_ne_mapping m JOIN tbl_account u ON u.account_id=m.user_id JOIN cli_ne n ON n.id=m.tbl_ne_id ORDER BY u.account_name, n.name;" 2>/dev/null

## ── Metrics & profiling ────────────────────────────────────────────

metric: ## Get app runtime metrics
	@curl -s http://localhost:$(SERVER_PORT)/metrics 2>/dev/null | python3 -m json.tool || echo "App not running on port $(SERVER_PORT)"

pprof-heap: ## Get heap profile
	go tool pprof http://localhost:6060/debug/pprof/heap

pprof-cpu: ## Get 30s CPU profile
	go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

pprof-goroutine: ## List goroutines
	@curl -s http://localhost:6060/debug/pprof/goroutine?debug=1 2>/dev/null | head -50 || echo "pprof not running (set PPROF_ENABLED=true)"

## ── Test ───────────────────────────────────────────────────────────

test: ## Run all tests
	go test ./...

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)
