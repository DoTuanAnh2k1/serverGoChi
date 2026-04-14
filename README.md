# cli-mgt-svc

Management service cho hệ thống CLI viễn thông 5G — viết bằng Go + Chi.  
Phục vụ quản lý người dùng, phân quyền, Network Element (NE), lịch sử lệnh và backup config NETCONF.

---

## Hệ sinh thái dịch vụ

| Service | Repo | Mô tả |
|---|---|---|
| **cli-mgt-svc** | repo này | Management API + Admin frontend |
| **cli-netconf-svc** | [cli-netconf](https://github.com/DoTuanAnh2k1/cli-netconf) | SSH server cho mode ne-config (NETCONF) |
| **SSH_SERVER** | *(repo riêng)* | SSH server cho mode ne-command |

---

## Tính năng

| Nhóm | Mô tả |
|---|---|
| **Authentication** | Đăng nhập, sinh JWT, lịch sử đăng nhập |
| **User Management** | Tạo / cập nhật / vô hiệu hóa tài khoản |
| **Permission (RBAC)** | Quản lý role, gán role cho user |
| **Network Element** | Tạo / xóa / quản lý NE (site, IP, port, namespace) |
| **NE Config** | Lưu thông tin kết nối NE cho mode ne-config (IP, port, protocol, credential) |
| **User-NE Authorization** | Phân quyền user truy cập NE |
| **Config Backup** | Lưu backup config XML từ NETCONF commit (metadata DB + file disk) |
| **History / Audit** | Lưu lịch sử lệnh, export CSV hàng ngày, tự động dọn dữ liệu cũ |
| **Admin Frontend** | Giao diện quản trị tại `/admin` (embedded, không cần build riêng) |
| **Import** | Import hàng loạt users/NEs/roles/configs từ file text |
| **Metrics & pprof** | Runtime metrics tại `/metrics`, Go pprof tại `:6060` |
| **TCP Subscriber Server** | Nhận kết nối TCP, lưu file khi client ngắt kết nối |
| **Leader Election** | Chỉ một pod thực thi task định kỳ (Kubernetes Lease) |

---

## Quick Start

```bash
# 1. Clone repo và cli-netconf-svc vào cùng thư mục cha
git clone https://github.com/DoTuanAnh2k1/serverGoChi
git clone https://github.com/DoTuanAnh2k1/cli-netconf

# 2. Tạo file .env
cp serverGoChi/deploy/local/.env.example serverGoChi/.env
# Bắt buộc đặt JWT_SECRET_KEY trong .env

# 3. Chạy toàn bộ stack
cd serverGoChi
JWT_SECRET_KEY=your-secret docker compose up -d
```

Mở browser: `http://localhost:3000/admin`

**Tài khoản mặc định:** `anhdt195` / `123` (tự động tạo khi app khởi động)

---

## Makefile

| Lệnh | Mô tả |
|---|---|
| `make up` | Start DB containers + chạy app local |
| `make down` | Stop tất cả |
| `make build` | Build binary → `bin/mgt-service` |
| `make build-docker` | Build Docker image |
| `make import FILE=data.txt` | Import data từ file text |
| `make test` | Chạy tất cả tests |
| `make metric` | Get runtime metrics |
| `make pprof-heap` | Mở pprof heap profile |
| `make pprof-cpu` | Lấy CPU profile 30s |
| `make clean` | Xóa build artifacts |

---

## Database

### Multi-DB support

Hỗ trợ 3 backend, chọn qua `DB_DRIVER`:

| Giá trị | Database |
|---|---|
| `mysql` | MySQL 8.0+ |
| `postgres` | PostgreSQL 16+ |
| `mongodb` | MongoDB 7.0+ |

### Auto-migrate

MySQL và PostgreSQL tự động tạo/cập nhật tất cả tables khi app khởi động.  
Không cần chạy SQL thủ công.

### Schema (db.sql)

File `db.sql` chứa DDL đầy đủ cho tất cả 8 bảng:

| Bảng | PK | Mô tả |
|---|---|---|
| `tbl_account` | `account_id` bigint AI | Tài khoản người dùng — bcrypt password, `is_enable` (soft-delete), `account_type` (0=SuperAdmin / 1=Admin / 2=Normal), `only_ad` (AD-only login) |
| `cli_ne` | `id` bigint AI | Network Element — kết nối lưu trong các cột `conf_*` (master/slave IP, SSH/TCP port, credential, `conf_mode`) |
| `cli_role` | `role_id` bigint AI | Định nghĩa role/permission: `permission`, `scope`, `ne_type`, `include_type`, `path` |
| `cli_role_user_mapping` | (`user_id`, `permission`) | Gán permission cho user — FK → `tbl_account` |
| `cli_user_ne_mapping` | (`user_id`, `tbl_ne_id`) | Gán NE cho user — FK → `tbl_account`, `cli_ne` |
| `cli_login_history` | `id` int AI | Lịch sử đăng nhập: `user_name`, `ip_address`, `time_login` |
| `cli_operation_history` | `id` int AI | Audit log: `account`, `cmd_name`, `ne_name`, `ne_ip`, `scope`, `result`, `created_date` |
| `cli_config_backup` | `id` bigint AI | Metadata backup NETCONF: `ne_name`, `ne_ip`, `file_path` (file XML lưu trên disk), `size`, `created_at` |

**Mapping CliNeConfig DTO ↔ cột `cli_ne`** (config API chỉ thao tác trên master):

| CliNeConfig field | Cột `cli_ne` |
|---|---|
| `ip_address` | `conf_master_ip` |
| `port` | `conf_port_master_ssh` |
| `protocol` | `conf_mode` (SSH / TELNET / NETCONF / RESTCONF) |
| `username` | `conf_username` |
| `password` | `conf_password` |

---

## Deploy

### Docker Compose (mặc định — private registry)

```bash
# Đảm bảo cli-netconf được clone cùng cấp với repo này:
#   parent/
#   ├── serverGoChi/   ← repo này
#   └── cli-netconf/   ← https://github.com/DoTuanAnh2k1/cli-netconf

JWT_SECRET_KEY=your-secret docker compose up -d
```

File `docker-compose.yml` ở root dùng private registry (`172.20.1.22`).  
Các service:

| Service | Container | Port |
|---|---|---|
| cli-mgt-svc | `cli-mgt-svc` | 3000 (API), 3675 (TCP), 6060 (pprof) |
| frontend | `mgt-frontend` | 8080 |
| cli-netconf-svc | `cli-netconf-svc` | 2222 (SSH) |
| MySQL | `mgt-mysql` | 3306 |
| PostgreSQL | `mgt-postgres` | 5432 |
| MongoDB | `mgt-mongodb` | 27017 |

### Kubernetes

```bash
kubectl apply -f deploy/k8s/rbac.yaml
kubectl apply -f deploy/k8s/pvc.yaml
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/secret.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml

# Frontend (standalone, optional)
kubectl apply -f deploy/k8s/frontend-deployment.yaml
kubectl apply -f deploy/k8s/frontend-service.yaml
```

---

## Import Tool

### Định dạng file

CSV, phân section bằng `[section_name]`, dòng đầu mỗi section là header, `#` là comment:

```
[users]
username,password
anhdt195,123
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
anhdt195,admin
operator1,operator

[user_nes]
username,ne_name
anhdt195,HTSMF01
operator1,HTAMF01

[ne_configs]
ne_name,ip_address,port,username,password,protocol,description
HTSMF01,10.10.1.1,830,netconf,netconf123,NETCONF,SMF NETCONF interface
HTSMF01,10.10.1.1,22,admin,admin123,SSH,SMF SSH fallback
HTAMF01,10.10.2.1,830,netconf,netconf123,NETCONF,AMF NETCONF interface
```

| Section | Trường | Ghi chú |
|---|---|---|
| `[users]` | username, password | Bỏ qua nếu user đã tồn tại |
| `[nes]` | name, site_name, ip_address, port, namespace, description | Luôn insert mới |
| `[roles]` | permission, scope, ne_type, include_type, path | Bỏ qua nếu role đã tồn tại |
| `[user_roles]` | username, permission | Gán permission cho user |
| `[user_nes]` | username, ne_name | Gán NE cho user |
| `[ne_configs]` | ne_name, ip_address, port, username, password, protocol, description | Luôn insert mới; protocol mặc định `SSH` |

### Frontend

`http://localhost:3000/admin` → tab **Import** → Load Sample hoặc upload file.

---

## Admin Frontend

Hỗ trợ song ngữ Tiếng Việt / English.

| Tab | Mô tả |
|---|---|
| Dashboard | Tổng quan users, permissions, NEs |
| Users | Tạo / vô hiệu hóa user |
| Permissions | Tạo / xóa role |
| Network Elements | Tạo / xóa NE |
| NE Config | Quản lý cấu hình kết nối NE (IP, port, protocol, credential) |
| NE Mapping | Gán / xóa NE cho user |
| Role Mapping | Gán / xóa role cho user |
| History | Xem lịch sử thao tác, filter theo scope và NE |
| Import | Upload file hoặc paste data để import hàng loạt |
| Guide | Hướng dẫn sử dụng |

---

## API

Tất cả endpoint (trừ `/health`, `/metrics`, `/admin`, `/docs`) yêu cầu:
```
Authorization: Basic <jwt_token>
```
> Token lấy từ `POST /aa/authenticate` — field `response_data` đã chứa sẵn prefix `Basic `, paste nguyên giá trị đó vào header.

### Authentication
| Method | Path | Mô tả |
|---|---|---|
| `POST` | `/aa/authenticate` | Đăng nhập, lấy JWT |
| `POST` | `/aa/validate-token` | Kiểm tra token |
| `POST` | `/aa/change-password` | Đổi mật khẩu (cần old_password) |
| `POST` | `/aa/authenticate/user/set` | Tạo hoặc kích hoạt lại user |
| `POST` | `/aa/authenticate/user/delete` | Vô hiệu hóa user (soft-delete) |
| `GET`  | `/aa/authenticate/user/show` | Danh sách user kèm NE & role |
| `POST` | `/aa/authenticate/user/reset-password` | Admin đặt lại mật khẩu (không cần old_password) |

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

Body của `/create` và `/update` nhận trực tiếp các trường `CliNe` (json tag snake_case).  
Kết nối được lưu trong cột `conf_*` — không dùng `ip_address`/`port` mà dùng `conf_master_ip`/`conf_port_master_ssh`.

| Method | Path | Mô tả |
|---|---|---|
| `POST` | `/aa/authorize/ne/create` | Tạo NE mới (`ne_name` bắt buộc) |
| `POST` | `/aa/authorize/ne/update` | Cập nhật NE (`id` bắt buộc) |
| `POST` | `/aa/authorize/ne/remove` | Xóa NE + toàn bộ mappings/configs (cascade) |
| `POST` | `/aa/authorize/ne/set` | Gán NE cho user |
| `POST` | `/aa/authorize/ne/delete` | Xóa NE khỏi user |
| `GET`  | `/aa/authorize/ne/show` | Danh sách NE (system_type=5GC) |
| `GET`  | `/aa/list/ne` | NE mà user đang đăng nhập được truy cập |
| `GET`  | `/aa/list/ne/monitor` | NE monitor URL |

### NE Config
| Method | Path | Mô tả |
|---|---|---|
| `POST` | `/aa/authorize/ne/config/create` | Tạo cấu hình kết nối cho NE |
| `GET`  | `/aa/authorize/ne/config/list?ne_id=X` | Danh sách config của một NE |
| `POST` | `/aa/authorize/ne/config/update` | Cập nhật config |
| `POST` | `/aa/authorize/ne/config/delete` | Xóa config theo ID |
| `GET`  | `/aa/list/ne/config` | Toàn bộ config của các NE thuộc user hiện tại |

### Config Backup (NETCONF commit)
| Method | Path | Mô tả |
|---|---|---|
| `POST` | `/aa/config-backup/save` | Lưu backup config XML |
| `GET`  | `/aa/config-backup/list?ne_name=X` | Danh sách backup (metadata) |
| `GET`  | `/aa/config-backup/{id}` | Lấy backup kèm nội dung XML |

### History
| Method | Path | Mô tả |
|---|---|---|
| `GET`  | `/aa/history/list` | Lịch sử lệnh (`?limit=N&scope=X&ne_name=Y`) |
| `POST` | `/aa/history/save` | Lưu bản ghi lịch sử |

**Scope types:**

| Scope | Nguồn | Mô tả |
|---|---|---|
| `cli-config` | cli-mgt-svc | Audit thao tác quản trị |
| `ne-command` | SSH_SERVER | Lịch sử lệnh chạy trên NE |
| `ne-config` | cli-netconf-svc | Lịch sử cấu hình NETCONF |

### Import & Subscribers
| Method | Path | Mô tả |
|---|---|---|
| `POST` | `/aa/import/` | Import hàng loạt (plain text body) |
| `GET`  | `/aa/subscribers/files` | Danh sách file subscriber |
| `GET`  | `/aa/subscribers/files/{index}` | Xem nội dung file |

---

## Cấu hình

```bash
cp deploy/local/.env.example .env
```

Các biến quan trọng:

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

# JWT (bắt buộc)
JWT_SECRET_KEY=change-me-in-production
JWT_EXPIRY_HOURS=24

# Logging
LOG_LEVEL=info
DB_LOG_LEVEL=warn

# TCP subscriber server
TCP_LISTEN_PORT=3675
TCP_DATA_DIR=/data/subscribers

# Config backup (NETCONF commit snapshots)
CLI_CONFIG_BACKUP_DIR=/data/config-backups

# CSV export
CLI_LOG_EXPORT_DIR=/data/csv
CSV_EXPORT_HOUR=23

# Profiling
PPROF_ENABLED=true
PPROF_ADDR=:6060

# Leader election (K8s only)
LEADER_ELECTION_ENABLED=false
```

---

## Cấu trúc thư mục

```
cli-mgt-svc/
├── cmd/
│   ├── main/               # Entry point
│   ├── frontend/           # Standalone frontend server
│   └── import/             # CLI import tool
├── models/
│   ├── config_models/      # Struct cấu hình
│   └── db_models/          # GORM / bson models
├── pkg/
│   ├── handler/            # HTTP handlers, frontend, router
│   │   ├── middleware/     # Auth, CheckRole, RateLimit, CORS
│   │   └── response/       # JSON response helper
│   ├── repository/         # Data access layer
│   │   ├── mysql/
│   │   ├── postgres/
│   │   └── mongodb/
│   ├── service/            # Business logic + seed
│   ├── store/              # DB interface + singleton
│   ├── testutil/           # Mock store
│   └── token/              # JWT create/parse
├── deploy/
│   ├── docker/             # Dockerfiles + docker-compose-private.yaml
│   ├── k8s/                # Kubernetes manifests
│   └── local/              # .env.example, sample_import.txt, SQL scripts
├── docker-compose.yml      # Default deploy (private registry)
├── api.yaml                # OpenAPI 3.0 spec
└── Makefile
```

---

## Tech stack

- **Go 1.25+** · **Chi** · **GORM** · **mongo-driver**
- **JWT (golang-jwt/jwt v5)** · **bcrypt** · **Logrus** · **godotenv**
- **k8s.io/client-go** (leader election) · **net/http/pprof**
- **Docker** multi-stage build · **MySQL 8** / **PostgreSQL 16** / **MongoDB 7**
