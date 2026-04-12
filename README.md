# mgt-service

REST API server viết bằng Go + Chi, phục vụ quản lý người dùng, phân quyền và Network Element (NE) trong hệ thống viễn thông 5G.

---

## Tính năng

| Nhóm | Mô tả |
|---|---|
| **Authentication** | Đăng nhập, sinh JWT access token, lịch sử đăng nhập |
| **User Management** | Tạo / cập nhật / xóa tài khoản |
| **Permission (RBAC)** | Quản lý role, gán role cho user |
| **Network Element** | Tạo / xóa / quản lý NE (site, IP, port, namespace) |
| **User-NE Authorization** | Phân quyền user truy cập NE |
| **History / Audit** | Lưu lịch sử lệnh, export CSV hàng ngày, tự động xóa dữ liệu cũ theo tháng |
| **Admin Frontend** | Giao diện quản trị tại `/admin` (embedded, không cần build riêng) |
| **Import** | Import data hàng loạt từ file text hoặc qua frontend |
| **Metrics & pprof** | Runtime metrics tại `/metrics`, Go pprof server tại `:6060` |
| **TCP Subscriber Server** | Nhận kết nối TCP, đọc từng dòng, lưu file khi ngắt kết nối |
| **Leader Election** | Chỉ một pod thực thi task định kỳ (dùng Kubernetes Lease) |

---

## Quick Start

```bash
# Clone & chạy (cần Docker + Go >= 1.25)
make up
```

Mở browser: `http://localhost:3000/admin` — login `admin` / `admin123` (sau khi import data).

---

## Makefile

| Lệnh | Mô tả |
|---|---|
| `make up` | Start DB containers (MySQL, PostgreSQL, MongoDB) + chạy app local |
| `make down` | Stop tất cả |
| `make build` | Build binary → `bin/mgt-service` |
| `make build-docker` | Build Docker image |
| `make import FILE=data.txt` | Import data từ file text |
| `make dump` | Dump toàn bộ data trong database |
| `make metric` | Get runtime metrics (RAM, CPU, goroutines) |
| `make pprof-heap` | Mở pprof heap profile (interactive) |
| `make pprof-cpu` | Lấy CPU profile 30s |
| `make pprof-goroutine` | List goroutines |
| `make test` | Chạy tất cả tests |
| `make clean` | Xóa build artifacts |

---

## Database

### Auto-migrate

Khi server khởi động, GORM tự động tạo/cập nhật tất cả tables cho MySQL và PostgreSQL. Không cần chạy SQL thủ công.

### SQL thủ công (backup)

Nếu cần tạo tables thủ công:

```bash
mysql -u <user> -p <db_name> < deploy/local/scripts/schema.sql   # tạo tables
mysql -u <user> -p <db_name> < deploy/local/scripts/seed.sql      # sample data
```

### Multi-DB support

Hỗ trợ 3 database backend, chọn qua `DB_DRIVER`:
- `mysql` — MySQL 8.0+
- `postgres` — PostgreSQL 16+
- `mongodb` — MongoDB 7.0+

---

## Import Tool

### CLI

```bash
make import FILE=deploy/local/sample_import.txt
# hoặc
go run ./cmd/import -file <path>
```

### Frontend

Vào `http://localhost:3000/admin` → tab **Import**:
- Upload file `.txt` / `.csv`
- Hoặc paste / chỉnh sửa trực tiếp trong text editor
- Bấm **Import** để import vào database

### Định dạng file

CSV, phân section bằng `[section_name]`, dòng đầu mỗi section là header, `#` là comment:

```
[users]
username,password
admin,admin123
operator1,Pass@123

[nes]
name,site_name,ip_address,port,namespace,description
HTSMF01,HCM,10.10.1.1,22,hcm-5gc,HCM SMF Node 01
HTAMF01,HCM,10.10.2.1,22,hcm-5gc,HCM AMF Node 01

[roles]
permission,scope,ne_type,include_type,path
admin,ext-config,5GC,include,/
operator,ext-config,5GC,include,/

[user_roles]
username,permission
admin,admin
operator1,operator

[user_nes]
username,ne_name
admin,HTSMF01
operator1,HTAMF01
```

File mẫu: `deploy/local/sample_import.txt`

> Import tool cũng trigger auto-migrate, nên có thể chạy trên DB trống.

---

## Admin Frontend

Hỗ trợ song ngữ Tiếng Việt / English, chuyển đổi ở sidebar.

### Chế độ embedded (chạy chung với server)

```
http://localhost:3000/admin
```

### Chế độ standalone (chạy riêng)

```bash
MGT_API_URL=http://10.10.1.100:3000/aa FRONTEND_PORT=8080 go run ./cmd/frontend
```

### Các tab

- **Dashboard** — tổng quan users, permissions, NEs
- **Users** — tạo / vô hiệu hóa user
- **Permissions** — tạo / xóa role
- **Network Elements** — tạo / xóa NE (5GC core nodes)
- **NE Mapping** — gán / xóa NE cho user (autocomplete search)
- **Role Mapping** — gán / xóa role cho user (autocomplete search)
- **History** — xem lịch sử thao tác, filter theo scope và NE
- **Import** — upload file hoặc paste data để import hàng loạt
- **Guide** — hướng dẫn sử dụng toàn bộ hệ thống

---

## Metrics & Profiling

### Runtime metrics

```bash
make metric
# hoặc
curl http://localhost:3000/metrics
```

Trả về JSON:

```json
{
  "goroutines": 8,
  "heap_alloc_mb": 1.23,
  "heap_sys_mb": 7.5,
  "sys_mem_mb": 12.3,
  "num_gc": 5,
  "num_cpu": 8,
  "gomaxprocs": 8,
  "go_version": "go1.26.0"
}
```

### Go pprof

Bật bằng `PPROF_ENABLED=true` (mặc định bật khi `make up`), listen `:6060`:

```bash
make pprof-heap         # heap profile (interactive)
make pprof-cpu          # CPU profile 30s
make pprof-goroutine    # list goroutines
```

Hoặc truy cập trực tiếp: `http://localhost:6060/debug/pprof/`

---

## Tech stack

- **Go 1.25+** — ngôn ngữ chính
- **Chi** — HTTP router
- **GORM** — ORM, hỗ trợ MySQL / PostgreSQL / MongoDB
- **JWT (golang-jwt/jwt v5)** — xác thực token
- **bcrypt** — mã hóa mật khẩu
- **Logrus** — structured logging
- **godotenv** — đọc config từ file `.env`
- **net/http/pprof** — Go profiling
- **k8s.io/client-go** — leader election qua Kubernetes Lease API
- **Docker** — multi-stage build

---

## Cấu trúc thư mục

```
mgt-service/
├── cmd/
│   ├── main/                   # Entry point (server)
│   ├── frontend/               # Standalone frontend server
│   └── import/                 # CLI import tool
├── web/
│   └── index.html              # Frontend HTML (source, dùng cho standalone)
├── models/
│   ├── config_models/          # Struct cấu hình
│   └── db_models/              # GORM model (generated)
├── pkg/
│   ├── bcrypt/                 # bcrypt encode/match
│   ├── config/                 # Config init từ env
│   ├── handler/                # HTTP handler, frontend, metrics
│   │   ├── middleware/         # Authenticate, CheckRole, RateLimit, CORS
│   │   └── response/          # JSON response helper
│   ├── leader/                 # Leader election + task định kỳ
│   ├── logger/                 # App logger + GORM logger
│   ├── repository/             # Data access layer
│   │   ├── mysql/
│   │   ├── postgres/
│   │   └── mongodb/
│   ├── server/                 # HTTP server (graceful shutdown)
│   ├── service/                # Business logic
│   ├── store/                  # DB interface + singleton
│   ├── tcpserver/              # TCP subscriber server
│   ├── testutil/               # Mock store cho test
│   └── token/                  # JWT create/parse
├── deploy/
│   ├── docker/
│   │   ├── Dockerfile                      # mgt-service (public)
│   │   ├��─ Dockerfile_private              # mgt-service (private registry)
│   │   ├── Dockerfile_frontend             # frontend (public)
│   │   ├── Dockerfile_frontend_private     # frontend (private registry)
│   │   ├── docker-compose.yml
│   │   └── docker-compose-private.yaml
│   ├── k8s/                    # Kubernetes manifests (service + frontend)
│   └── local/                  # Local dev files
│       ├── .env.example
│       ├── sample_import.txt
│       └── scripts/
│           ├── schema.sql
│           └── seed.sql
├── api.yaml                    # OpenAPI 3.0 spec
└── Makefile
```

---

## Cấu hình

Copy file mẫu và chỉnh sửa:

```bash
cp deploy/local/.env.example .env
```

Các biến môi trường quan trọng:

```env
# HTTP server
SERVER_HOST=0.0.0.0
SERVER_PORT=3000

# Database (mysql | postgres | mongodb)
DB_DRIVER=mysql
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=secret
MYSQL_DB_NAME=cli_db

# JWT
JWT_SECRET_KEY=change-me-in-production
JWT_EXPIRY_HOURS=24

# Logging
LOG_LEVEL=info          # debug | info | warn | error
DB_LOG_LEVEL=warn

# TCP subscriber server
TCP_LISTEN_PORT=3675
TCP_DATA_DIR=/data/subscribers

# Profiling
PPROF_ENABLED=true      # bật pprof server
PPROF_ADDR=:6060        # pprof listen address

# CSV export (leader task)
CLI_LOG_EXPORT_DIR=/data/csv
CSV_EXPORT_HOUR=23      # giờ export mỗi ngày (0–23)

# Leader election (cần khi chạy nhiều replica trong K8s)
LEADER_ELECTION_ENABLED=false
LEASE_LOCK_NAME=mgt-service-leader
LEASE_LOCK_NAMESPACE=default
POD_NAME=local-dev
```

---

## API

Tất cả endpoint (trừ `/health`, `/metrics`, `/admin`, `/docs`) yêu cầu header:

```
Authorization: Basic <jwt_token>
```

| Endpoint | Mô tả |
|---|---|
| `GET /health` | Health check |
| `GET /metrics` | Runtime metrics |
| `GET /admin` | Admin frontend |
| `GET /docs` | Swagger UI |

### Authentication

| Method | Path | Mô tả |
|---|---|---|
| `POST` | `/aa/authenticate` | Đăng nhập, lấy JWT |
| `POST` | `/aa/validate-token` | Kiểm tra token hợp lệ |
| `POST` | `/aa/change-password` | Đổi mật khẩu |

### User Management

| Method | Path | Mô tả |
|---|---|---|
| `POST` | `/aa/authenticate/user/set` | Tạo hoặc kích hoạt lại user |
| `POST` | `/aa/authenticate/user/delete` | Vô hiệu hóa user |
| `GET`  | `/aa/authenticate/user/show` | Danh sách user kèm NE & role |

### Permission (RBAC)

| Method | Path | Mô tả |
|---|---|---|
| `POST` | `/aa/authorize/permission/set` | Tạo role |
| `POST` | `/aa/authorize/permission/delete` | Xóa role |
| `GET`  | `/aa/authorize/permission/show` | Danh sách role |
| `POST` | `/aa/authorize/user/set` | Gán quyền cho user |
| `POST` | `/aa/authorize/user/delete` | Xóa quyền của user |
| `GET`  | `/aa/authorize/user/show` | Quyền của tất cả user |

### Network Element

| Method | Path | Mô tả |
|---|---|---|
| `POST` | `/aa/authorize/ne/create` | Tạo NE mới |
| `POST` | `/aa/authorize/ne/remove` | Xóa NE theo ID |
| `POST` | `/aa/authorize/ne/set` | Gán NE cho user |
| `POST` | `/aa/authorize/ne/delete` | Xóa NE khỏi user |
| `GET`  | `/aa/authorize/ne/show` | Danh sách NE (5GC) |
| `GET`  | `/aa/list/ne` | NE mà user đang đăng nhập được phép truy cập |
| `GET`  | `/aa/list/ne/monitor` | NE monitor URL |

### Import

| Method | Path | Mô tả |
|---|---|---|
| `POST` | `/aa/import/` | Import data (plain text body, format import.txt) |

### History

| Method | Path | Mô tả |
|---|---|---|
| `GET`  | `/aa/history/list` | Danh sách lịch sử lệnh (`?limit=N&scope=X&ne_name=Y`) |
| `POST` | `/aa/history/save` | Lưu một bản ghi lịch sử lệnh |

**Scope types:**

| Scope | Nguồn | Mô tả |
|---|---|---|
| `cli-config` | mgt-service (AA server) | Audit thao tác quản trị (tạo/xóa user, NE, role...) |
| `ne-command` | SSH CLI app (ne-command) | Lịch sử lệnh chạy trên NE |
| `ne-config` | SSH CLI app (ne-config) | Lịch sử cấu hình NE |

Filter ví dụ: `GET /aa/history/list?limit=100&scope=ne-command&ne_name=HTSMF01`

### Subscriber files

| Method | Path | Mô tả |
|---|---|---|
| `GET` | `/aa/subscribers/files` | Danh sách file subscriber đã lưu |
| `GET` | `/aa/subscribers/files/{index}` | Xem nội dung file theo index |

---

## TCP Subscriber Server

Server lắng nghe tại `TCP_LISTEN_PORT` (mặc định `3675`).

- Mỗi kết nối TCP được xử lý trong một goroutine riêng.
- Server đọc dữ liệu từng dòng (`\n`).
- Khi client ngắt kết nối, toàn bộ dữ liệu được ghi vào `TCP_DATA_DIR/list_subscribers_results.<index>`.
- `<index>` tự động tăng, atomic để tránh race condition.

---

## Leader Election & Task định kỳ

Khi `LEADER_ELECTION_ENABLED=true`, mỗi pod tranh giành Kubernetes Lease.  
Pod nào giữ Lease sẽ chạy hai task:

| Task | Lịch | Hành động |
|---|---|---|
| **CSV Export** | Hàng ngày lúc `CSV_EXPORT_HOUR:00` | Export lịch sử lệnh ra CSV, nhóm theo NE |
| **History Cleanup** | Ngày 1 đầu tháng | Xóa bản ghi cũ hơn tháng trước |

---

## Deploy

### Docker Compose

```bash
# Public registry
docker compose -f deploy/docker/docker-compose.yml up -d

# Private registry (cụm nội bộ)
docker compose -f deploy/docker/docker-compose-private.yaml up -d
```

### Kubernetes

```bash
# mgt-service
kubectl apply -f deploy/k8s/rbac.yaml
kubectl apply -f deploy/k8s/pvc.yaml
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/secret.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml

# frontend (standalone, optional)
kubectl apply -f deploy/k8s/frontend-deployment.yaml
kubectl apply -f deploy/k8s/frontend-service.yaml
```

> Pod mgt-service cần `ServiceAccount` có quyền ghi `coordination.k8s.io/leases` — đã cấu hình sẵn trong `rbac.yaml`.  
> Frontend trỏ `MGT_API_URL=http://mgt-service/aa` qua ClusterIP.

---

## Sinh DB model từ MySQL (GORM Gen)

```bash
go install gorm.io/gen/tools/gentool@latest

gentool -dsn "user:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local" \
        -outPath ./models/db_models
```
