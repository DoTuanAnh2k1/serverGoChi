# mgt-service

REST API server viết bằng Go + Chi, phục vụ quản lý người dùng, phân quyền và Network Element (NE) trong hệ thống viễn thông.

---

## Tính năng

| Nhóm | Mô tả |
|---|---|
| **Authentication** | Đăng nhập, sinh JWT access token, lịch sử đăng nhập |
| **User Management** | Tạo / cập nhật / xóa tài khoản |
| **Permission (RBAC)** | Quản lý role, gán role cho user |
| **Network Element** | Quản lý danh sách NE (site, IP, port, namespace) và NE monitor |
| **User-NE Authorization** | Phân quyền user truy cập NE |
| **History / Audit** | Lưu lịch sử lệnh, export CSV hàng ngày, tự động xóa dữ liệu cũ theo tháng |
| **TCP Subscriber Server** | Nhận kết nối TCP, đọc từng dòng, lưu file khi ngắt kết nối |
| **Subscriber API** | Liệt kê và xem nội dung file subscriber đã lưu |
| **Leader Election** | Chỉ một pod thực thi task định kỳ (dùng Kubernetes Lease) |

---

## Tech stack

- **Go 1.25** — ngôn ngữ chính
- **Chi** — HTTP router
- **GORM** — ORM kết nối MySQL
- **JWT (golang-jwt/jwt v5)** — xác thực token
- **bcrypt** — mã hóa mật khẩu
- **Logrus** — structured logging với formatter tùy chỉnh
- **godotenv** — đọc config từ file `.env`
- **k8s.io/client-go** — leader election qua Kubernetes Lease API
- **Docker** — multi-stage build, Alpine, non-root user

---

## Cấu trúc thư mục

```
mgt-service/
├── cmd/main/               # Entry point
├── models/
│   ├── config_models/      # Struct cấu hình (server, db, log, jwt, leader…)
│   └── db_models/          # Model sinh tự động từ GORM gen
├── internal/
│   ├── bcrypt/             # bcrypt encode/match
│   ├── config/             # Đọc và khởi tạo config từ env
│   ├── handler/            # HTTP handler, middleware, response helper
│   │   ├── middleware/     # Authenticate, CheckRole, RateLimit, CORS
│   │   └── response/       # Chuẩn hóa JSON response
│   ├── leader/             # Leader election + task định kỳ
│   ├── logger/             # App logger (Logrus) + GORM logger
│   ├── repository/mysql/   # Query layer (GORM)
│   ├── server/             # HTTP server lifecycle (graceful shutdown)
│   ├── service/            # Business logic (auth, user, ne, permission, history)
│   ├── store/              # Interface DB + singleton
│   ├── tcpserver/          # TCP server nhận dữ liệu subscriber
│   ├── testutil/           # Mock store + init logger cho test
│   └── token/              # JWT create/parse
├── deploy/k8s/             # Kubernetes manifests
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── rbac.yaml
│   ├── pvc.yaml
│   ├── deployment.yaml
│   └── service.yaml
├── api.yaml                # OpenAPI 3.0 spec
└── Dockerfile
```

---

## Cài đặt & chạy

### 1. Yêu cầu

- Go >= 1.25
- MySQL >= 5.7

### 2. Cấu hình `.env`

Copy file mẫu và chỉnh sửa:

```bash
cp .env.example .env
```

Các biến môi trường quan trọng:

```env
# HTTP server
SERVER_HOST=0.0.0.0
SERVER_PORT=3000

# MySQL
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

# CSV export (leader task)
CLI_LOG_EXPORT_DIR=/data/csv
CSV_EXPORT_HOUR=23      # giờ export mỗi ngày (0–23)

# Leader election (cần khi chạy nhiều replica)
LEADER_ELECTION_ENABLED=true
LEASE_LOCK_NAME=mgt-service-leader
LEASE_LOCK_NAMESPACE=default
POD_NAME=local-dev      # tự inject bởi k8s fieldRef khi deploy
```

### 3. Chạy local

```bash
go mod download
go run cmd/main/main.go
```

### 4. Chạy với Docker

```bash
docker build -t mgt-service .
docker run --env-file .env -p 3000:3000 -p 3675:3675 mgt-service
```

### 5. Test

```bash
go test ./...
```

---

## API

Tất cả endpoint (trừ `/health`, `/docs`) đều yêu cầu header:

```
Authorization: Bearer <jwt_token>
```

Swagger UI: `GET /docs`  
OpenAPI spec: `GET /docs/openapi.yaml`

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
| `POST` | `/aa/authorize/ne/set` | Gán NE cho user |
| `POST` | `/aa/authorize/ne/delete` | Xóa NE khỏi user |
| `GET`  | `/aa/authorize/ne/show` | Danh sách NE (5GC) |
| `GET`  | `/aa/list/ne` | NE mà user đang đăng nhập được phép truy cập |
| `GET`  | `/aa/list/ne/monitor` | NE monitor URL |

### History

| Method | Path | Mô tả |
|---|---|---|
| `POST` | `/aa/history/save` | Lưu một bản ghi lịch sử lệnh |

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
- `<index>` tự động tăng (bắt đầu từ `0`), atomic để tránh race condition khi nhiều kết nối đồng thời.

---

## Leader Election & Task định kỳ

Khi `LEADER_ELECTION_ENABLED=true`, mỗi pod tranh giành Kubernetes Lease.  
Pod nào giữ Lease sẽ chạy hai task:

| Task | Lịch | Hành động |
|---|---|---|
| **CSV Export** | Hàng ngày lúc `CSV_EXPORT_HOUR:00` | Export lịch sử lệnh ra CSV, nhóm theo NE, lưu vào `CLI_LOG_EXPORT_DIR` |
| **History Cleanup** | Ngày 1 đầu tháng, lúc `00:00` | Xóa toàn bộ bản ghi cũ hơn tháng trước (ví dụ tháng 4 → xóa trước 2026-03-01) |

---

## Deploy lên Kubernetes

```bash
# 1. Điền secret thật (base64) vào deploy/k8s/secret.yaml, rồi apply
kubectl apply -f deploy/k8s/rbac.yaml
kubectl apply -f deploy/k8s/pvc.yaml
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/secret.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
```

> Pod cần `ServiceAccount` có quyền ghi `coordination.k8s.io/leases` để leader election hoạt động — đã cấu hình sẵn trong `rbac.yaml`.

---

## Sinh DB model từ MySQL (GORM Gen)

```bash
go install gorm.io/gen/tools/gentool@latest

gentool -dsn "user:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local" \
        -outPath ./models/db_models
```
