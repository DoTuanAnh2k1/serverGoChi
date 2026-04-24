# MGT Service Makefile

APP_NAME     := mgt-service
COMPOSE      := docker compose -f deploy/docker/docker-compose.yml
COMPOSE_PRV  := docker compose -f deploy/docker/docker-compose-private.yaml
COMPOSE_CMD  := docker compose -f deploy/docker/docker-compose-command-private.yaml
MAIN         := ./cmd/main/main.go
BUILD_DIR    := bin

# Ports
SERVER_PORT  := $(or $(SERVER_PORT),3000)
TCP_PORT     := $(or $(TCP_LISTEN_PORT),3675)
PPROF_PORT   := $(or $(PPROF_PORT),6060)
SWAGGER_PORT := $(or $(SWAGGER_PORT),8080)

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
	PPROF_ENABLED=true PPROF_ADDR=:$(PPROF_PORT) \
	SWAGGER_PORT=$(SWAGGER_PORT)
endef

.PHONY: help up up-docker up-local up-cmd down down-cmd build build-docker dump metric \
        pprof-heap pprof-cpu pprof-goroutine test clean logs logs-mgt logs-cli-command exec-mysql ps

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

## ── Environment ────────────────────────────────────────────────────

up: up-docker ## Start full environment in Docker (DB + app + pprof)

up-docker: ## Start all services in Docker containers
	JWT_SECRET_KEY=$${JWT_SECRET_KEY:-dev-secret-key} \
	PPROF_ENABLED=true PPROF_PORT=$(PPROF_PORT) \
	SWAGGER_PORT=$(SWAGGER_PORT) \
	$(COMPOSE) up -d --build
	@echo ""
	@echo "  App:     http://localhost:$(SERVER_PORT)"
	@echo "  Admin:   http://localhost:$(SERVER_PORT)/admin"
	@echo "  Metrics: http://localhost:$(SERVER_PORT)/metrics"
	@echo "  Swagger: http://localhost:$(SWAGGER_PORT)"
	@echo "  pprof:   http://localhost:$(PPROF_PORT)/debug/pprof/"
	@echo "  TCP:     localhost:$(TCP_PORT)"
	@echo "  MySQL:   localhost:$(DB_PORT)"
	@echo "  CLI SSH: ssh <user>@localhost -p $${NE_SSH_PORT:-2222}"
	@echo ""

up-cmd: ## Start minimal stack: MySQL + cli-mgt-svc + cli-command-svc (private registry)
	JWT_SECRET_KEY=$${JWT_SECRET_KEY:-dev-secret-key} \
	SWAGGER_PORT=$(SWAGGER_PORT) \
	$(COMPOSE_CMD) up -d --build
	@echo ""
	@echo "  App:     http://localhost:$(SERVER_PORT)"
	@echo "  Admin:   http://localhost:$(SERVER_PORT)/admin"
	@echo "  Swagger: http://localhost:$(SWAGGER_PORT)"
	@echo "  TCP:     localhost:$(TCP_PORT)"
	@echo "  MySQL:   localhost:$(DB_PORT)"
	@echo "  CLI SSH: ssh <user>@localhost -p $${NE_SSH_PORT:-2222}"
	@echo ""

down-cmd: ## Stop minimal stack (MySQL + cli-mgt-svc + cli-command-svc)
	-JWT_SECRET_KEY=x $(COMPOSE_CMD) down 2>/dev/null
	@echo "Done."

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

logs-mgt: ## Tail cli-mgt-svc container logs
	docker logs -f cli-mgt-svc

logs-cli-command: ## Tail cli-command-svc container logs
	docker logs -f cli-command-svc

exec-mysql: ## Open MySQL shell in mgt-mysql container
	docker exec -it mgt-mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME)

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

dump: ## Dump v2 core tables from MySQL
	@echo "=== users ==="
	@docker exec mgt-mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT id, username, email, is_enabled, locked_at FROM user;" 2>/dev/null
	@echo ""
	@echo "=== ne ==="
	@docker exec mgt-mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT id, namespace, ne_type, site_name, master_ip, master_port FROM ne;" 2>/dev/null
	@echo ""
	@echo "=== command ==="
	@docker exec mgt-mysql mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) \
		-e "SELECT id, ne_id, service, LEFT(cmd_text, 80) AS cmd FROM command;" 2>/dev/null

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
