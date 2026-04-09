# serverGoChi — Management Service

REST API server viết bằng Go + Chi, phục vụ quản lý người dùng, phân quyền và Network Element (NE) trong hệ thống viễn thông.

---

## Tính năng

| Nhóm | Mô tả |
|---|---|
| **Authentication** | Đăng nhập, sinh JWT access token |
| **User Management** | Tạo / cập nhật / xóa tài khoản người dùng |
| **Permission (RBAC)** | Quản lý role, gán quyền cho role, gán role cho user |
| **Network Element** | Quản lý danh sách NE (site, IP, port, namespace) và command-mode NE |
| **User-NE Authorization** | Phân quyền user được phép truy cập NE nào |
| **History / Audit** | Lưu & truy vấn lịch sử lệnh, lịch sử đăng nhập |

---

## Tech stack

- **Go 1.19+** — ngôn ngữ chính
- **Chi** — HTTP router
- **GORM** — ORM kết nối MySQL
- **JWT** — xác thực token
- **bcrypt** — mã hóa mật khẩu
- **godotenv** — đọc cấu hình từ file `.env`
- **Docker** — container hóa (multi-stage build, Alpine)

---

## Cấu trúc thư mục

```
serverGoChi/
├── cmd/main/           # Entry point
├── config/             # Đọc và khởi tạo config toàn cục
├── models/
│   ├── config_models/  # Struct cấu hình (server, db, log, jwt…)
│   └── db_models/      # Model sinh tự động từ GORM gen
├── src/
│   ├── db/mysql/       # Query layer (GORM)
│   ├── router/         # HTTP handler, middleware, response helper
│   ├── server/         # HTTP server lifecycle (start/stop graceful)
│   ├── service/        # Business logic (authenticate, user, authorize, history…)
│   ├── store/          # Khởi tạo kết nối DB
│   ├── logger/         # Wrapper logging (app + GORM)
│   └── utils/          # JWT helper, bcrypt helper
├── api.yaml            # OpenAPI 3.0 spec
└── Dockerfile
```

---

## Cài đặt & chạy

### 1. Yêu cầu

- Go >= 1.19
- MySQL >= 5.7
- (Tùy chọn) Docker

### 2. Cấu hình `.env`

Tạo file `.env` tại thư mục gốc:

```env
SERVER_HOST=0.0.0.0
SERVER_PORT=3000

DB_DRIVER=mysql
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=secret
MYSQL_DB_NAME=mgt_db

LOG_LEVEL=info
DB_LOG_LEVEL=warn
```

### 3. Chạy local

```bash
go mod download
go run cmd/main/main.go
```

### 4. Chạy với Docker

```bash
docker build -t mgt-service .
docker run --env-file .env -p 3000:3000 mgt-service
```

---

## API

Base URL: `{host}/mgt-svc/v1`

Xem đầy đủ tại [api.yaml](api.yaml) (OpenAPI 3.0).

| Method | Path | Mô tả |
|---|---|---|
| POST | `/user/authen` | Đăng nhập, lấy JWT |
| POST | `/user` | Tạo / cập nhật user |
| DELETE | `/user` | Xóa user |
| GET | `/user/permission` | Xem quyền của user |
| POST | `/user/permission` | Gán quyền cho user |
| DELETE | `/user/permission` | Xóa quyền của user |
| GET | `/users/permission` | Danh sách quyền tất cả user |
| POST | `/permission` | Tạo / cập nhật permission |
| DELETE | `/permission` | Xóa permission |
| POST | `/user/network-element` | Gán NE cho user |
| DELETE | `/user/network-element` | Xóa NE khỏi user |
| GET | `/user/network-elements` | Danh sách NE được phép của user |
| POST | `/user/network-elements/delete` | Xóa nhiều NE của user |
| GET | `/users/network-element` | Danh sách NE toàn bộ user |
| POST | `/network-element` | Tạo / cập nhật NE |
| DELETE | `/network-element` | Xóa NE |
| POST | `/network-element/command-mode` | Tạo / cập nhật command-mode NE |
| DELETE | `/network-element/command-mode` | Xóa command-mode NE |
| GET | `/network-elements` | Danh sách NE |
| GET | `/network-elements/command-mode` | Danh sách command-mode NE |
| GET | `/history` | Truy vấn lịch sử (theo NE, khoảng thời gian) |
| POST | `/history` | Lưu lịch sử |

---

## Sinh DB model từ MySQL (GORM Gen)

```bash
go install gorm.io/gen/tools/gentool@latest

gentool -dsn "user:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local"
```

Output được đặt vào `models/db_models/`.

---

## Những điểm cần cải thiện

1. **`godotenv.Load()` fatal khi không có `.env`** — nên fallback sang biến môi trường hệ thống thay vì `log.Fatal`.
2. **Token config rỗng** (`TokenConfig struct{}`) — cần bổ sung `SecretKey`, `ExpiryDuration` để cấu hình JWT linh hoạt.
3. **Lỗi `ListenAndServe` bị bỏ qua** trong `server.Start()` — nên log hoặc propagate error.
4. **`os.Exit(1)` khi shutdown bình thường** — nên dùng `os.Exit(0)`.
5. **Health check endpoint `/health`** được khai báo trong Dockerfile nhưng chưa có route tương ứng.
6. **Thiếu input validation** — các endpoint chưa validate body/query trước khi xử lý.
7. **Thiếu rate limiting** — endpoint `/user/authen` nên có rate limit để chống brute-force.
8. **Không có `.env.example`** — nên thêm file mẫu để onboarding dễ hơn.
9. **CORS chưa cấu hình** — `RouterConfig` có field `Origins/Methods/Headers` nhưng chưa thấy middleware CORS được dùng.
10. **Swagger UI** — có `api.yaml` nhưng chưa serve trực tiếp qua endpoint (ví dụ `/docs`).
