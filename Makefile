# MGT Service Makefile

APP_NAME     := mgt-service
COMPOSE      := docker compose -f deploy/docker/docker-compose.yml
COMPOSE_PRV  := docker compose -f deploy/docker/docker-compose-private.yaml
MAIN         := ./cmd/main/main.go
IMPORT_CMD   := ./cmd/import
BUILD_DIR    := bin

# Ports
SERVER_PORT  := $(or $(SERVER_PORT),3000)
TCP_PORT     := $(or $(TCP_LISTEN_PORT),3675)
PPROF_PORT   := $(or $(PPROF_PORT),6060)

# DB connection
DB_DRIVER    := $(or $(DB_DRIVER),mysql)
DB_HOST      := $(or $(MYSQL_HOST),localhost)
DB_PORT      := $(or $(MYSQL_PORT),3306)
DB_USER      := $(or $(MYSQL_USER),mgtuser)
DB_PASS      := $(or $(MYSQL_PASSWORD),mgtpassword)
DB_NAME      := $(or $(MYSQL_DB_NAME),cli_db)

# Common env for local go run
define APP_ENV
	DB_DRIVER=$(DB_DRIVER) \
	MYSQL_HOST=$(DB_HOST) MYSQL_PORT=$(DB_PORT) \
	MYSQL_USER=$(DB_USER) MYSQL_PASSWORD=$(DB_PASS) MYSQL_DB_NAME=$(DB_NAME) \
	SERVER_HOST=0.0.0.0 SERVER_PORT=$(SERVER_PORT) \
	JWT_SECRET_KEY=$${JWT_SECRET_KEY:-dev-secret-key} \
	JWT_EXPIRY_HOURS=24 LOG_LEVEL=info DB_LOG_LEVEL=warn \
	CORS_ORIGINS="*" LEADER_ELECTION_ENABLED=false \
	TCP_LISTEN_PORT=$(TCP_PORT) TCP_DATA_DIR=/tmp/mgt-subscribers \
	CLI_LOG_EXPORT_DIR=/tmp/mgt-csv CSV_EXPORT_HOUR=23 \
	PPROF_ENABLED=true PPROF_ADDR=:$(PPROF_PORT)
endef

.PHONY: help up up-docker up-local down build build-docker import dump metric \
        pprof-heap pprof-cpu pprof-goroutine test clean logs ps

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

## ── Environment ────────────────────────────────────────────────────

up: up-docker ## Start full environment in Docker (DB + app + pprof)

up-docker: ## Start all services in Docker containers
	JWT_SECRET_KEY=$${JWT_SECRET_KEY:-dev-secret-key} \
	PPROF_ENABLED=true PPROF_PORT=$(PPROF_PORT) \
	$(COMPOSE) up -d --build
	@echo ""
	@echo "  App:     http://localhost:$(SERVER_PORT)"
	@echo "  Admin:   http://localhost:$(SERVER_PORT)/admin"
	@echo "  Metrics: http://localhost:$(SERVER_PORT)/metrics"
	@echo "  pprof:   http://localhost:$(PPROF_PORT)/debug/pprof/"
	@echo "  TCP:     localhost:$(TCP_PORT)"
	@echo "  MySQL:   localhost:$(DB_PORT)"
	@echo ""

up-local: ## Start DB in Docker + app locally (hot reload with go run)
	$(COMPOSE) up -d mysql mongodb postgres
	@echo "Waiting for databases..."
	@sleep 10
	@echo "Starting app locally..."
	$(APP_ENV) go run $(MAIN)

down: ## Stop all services and kill local processes
	-JWT_SECRET_KEY=x $(COMPOSE) down 2>/dev/null
	@-pkill -f "go run $(MAIN)" 2>/dev/null || true
	@-lsof -ti:$(SERVER_PORT) | xargs kill -9 2>/dev/null || true
	@-lsof -ti:$(TCP_PORT) | xargs kill -9 2>/dev/null || true
	@-lsof -ti:$(PPROF_PORT) | xargs kill -9 2>/dev/null || true
	@echo "Done."

logs: ## Tail app container logs
	docker logs -f mgt-service

ps: ## Show running containers
	@docker ps --filter "name=mgt-" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

## ── Build ──────────────────────────────────────────────────────────

build: ## Build app binary -> bin/mgt-service
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)
	@echo "Built: $(BUILD_DIR)/$(APP_NAME)"

build-docker: ## Build Docker image (public registry)
	docker build -f deploy/docker/Dockerfile -t $(APP_NAME):latest .

build-docker-private: ## Build Docker image (private registry)
	docker build -f deploy/docker/Dockerfile_private -t $(APP_NAME):latest .

## ── Data ───────────────────────────────────────────────────────────

import: ## Import data from file: make import FILE=data.txt
	@if [ -z "$(FILE)" ]; then echo "Usage: make import FILE=<path>"; exit 1; fi
	$(APP_ENV) LOG_LEVEL=error DB_LOG_LEVEL=error \
	go run $(IMPORT_CMD) -file $(FILE)

dump: ## Dump all data from database
	@echo "=== Users ==="
	@docker exec mgt-mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT account_id, account_name, account_type, is_enable, created_by FROM tbl_account;" 2>/dev/null || \
		mysql -h$(DB_HOST) -P$(DB_PORT) -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT account_id, account_name, account_type, is_enable, created_by FROM tbl_account;" 2>/dev/null
	@echo ""
	@echo "=== Network Elements ==="
	@docker exec mgt-mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT id, name, site_name, ip_address, port, system_type FROM cli_ne;" 2>/dev/null || \
		mysql -h$(DB_HOST) -P$(DB_PORT) -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT id, name, site_name, ip_address, port, system_type FROM cli_ne;" 2>/dev/null
	@echo ""
	@echo "=== Roles ==="
	@docker exec mgt-mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT role_id, permission, scope, ne_type FROM cli_role;" 2>/dev/null || \
		mysql -h$(DB_HOST) -P$(DB_PORT) -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT role_id, permission, scope, ne_type FROM cli_role;" 2>/dev/null
	@echo ""
	@echo "=== User -> Roles ==="
	@docker exec mgt-mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT u.account_name, m.permission FROM cli_role_user_mapping m JOIN tbl_account u ON u.account_id=m.user_id ORDER BY u.account_name;" 2>/dev/null || \
		mysql -h$(DB_HOST) -P$(DB_PORT) -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT u.account_name, m.permission FROM cli_role_user_mapping m JOIN tbl_account u ON u.account_id=m.user_id ORDER BY u.account_name;" 2>/dev/null
	@echo ""
	@echo "=== User -> NEs ==="
	@docker exec mgt-mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT u.account_name, n.name AS ne_name, n.site_name FROM cli_user_ne_mapping m JOIN tbl_account u ON u.account_id=m.user_id JOIN cli_ne n ON n.id=m.tbl_ne_id ORDER BY u.account_name, n.name;" 2>/dev/null || \
		mysql -h$(DB_HOST) -P$(DB_PORT) -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT u.account_name, n.name AS ne_name, n.site_name FROM cli_user_ne_mapping m JOIN tbl_account u ON u.account_id=m.user_id JOIN cli_ne n ON n.id=m.tbl_ne_id ORDER BY u.account_name, n.name;" 2>/dev/null

## ── Metrics & profiling ────────────────────────────────────────────

metric: ## Get app runtime metrics
	@curl -s http://localhost:$(SERVER_PORT)/metrics 2>/dev/null | python3 -m json.tool || echo "App not running on port $(SERVER_PORT)"

pprof-heap: ## Interactive heap profile
	go tool pprof http://localhost:$(PPROF_PORT)/debug/pprof/heap

pprof-cpu: ## 30s CPU profile
	go tool pprof http://localhost:$(PPROF_PORT)/debug/pprof/profile?seconds=30

pprof-goroutine: ## List goroutines
	@curl -s http://localhost:$(PPROF_PORT)/debug/pprof/goroutine?debug=1 2>/dev/null | head -50 || echo "pprof not running on port $(PPROF_PORT)"

## ── Test ───────────────────────────────────────────────────────────

test: ## Run all tests
	go test ./...

clean: ## Remove build artifacts and stop containers
	rm -rf $(BUILD_DIR)
	-JWT_SECRET_KEY=x $(COMPOSE) down -v 2>/dev/null
