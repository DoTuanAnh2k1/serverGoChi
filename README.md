# cli-mgt-svc

Management service cho hệ thống CLI viễn thông 5G — viết bằng Go + Chi.  
Phục vụ quản lý người dùng, phân quyền, Network Element (NE), lịch sử lệnh và backup config NETCONF.

---

## Hệ sinh thái dịch vụ

| Service | Repo | Mô tả |
|---|---|---|
| **cli-mgt-svc** | repo này | Management API + Admin frontend (embedded) |
| **cli-ssh-svc** | repo này (`cmd/ssh`) | SSH CLI bastion — 3 mode: `cli-config` (quản trị user/NE/group qua REST), `ne-config` & `ne-command` (SSH proxy) |
| **cli-netconf-svc** | [cli-netconf](https://github.com/DoTuanAnh2k1/cli-netconf) | SSH server cho mode ne-config (NETCONF) |
| **SSH_SERVER** | *(repo riêng)* | SSH server cho mode ne-command |

---

## Tính năng

| Nhóm | Mô tả |
|---|---|
| **Authentication** | Đăng nhập, sinh JWT, lịch sử đăng nhập. Role lấy từ JWT `aud` claim — không cần gọi lại listing để xác thực role |
| **User Management** | Tạo / cập nhật / vô hiệu hóa tài khoản |
| **Permission** | Phân quyền dựa trên `account_type` (SuperAdmin/Admin/Normal). SuperAdmin ẩn khỏi mọi listing (`/aa/admin/user/list`, `/aa/authenticate/user/show`) nhưng vẫn đăng nhập & thao tác được bình thường |
| **Network Element** | Tạo / sửa / xóa NE (site, IP, port, namespace, conf_mode) |
| **User-NE Authorization** | Phân quyền user truy cập NE — trực tiếp hoặc qua Group |
| **Groups** | Gom user ↔ group ↔ NE; `/aa/list/ne` trả về hợp (union) của direct + via-group |
| **Config Backup** | Lưu backup config XML từ NETCONF commit |
| **History / Audit** | Lưu lịch sử lệnh, filter theo scope/NE, export CSV, tự dọn dữ liệu cũ |
| **Admin Frontend** | Giao diện quản trị tại `/admin` (embedded, song ngữ EN/VI) |
| **Import** | Import hàng loạt users/NEs/mappings từ file text hoặc frontend |
| **Metrics & pprof** | Runtime metrics tại `/metrics`, Go pprof tại `:6060` |
| **Multi-DB** | MySQL / MariaDB / PostgreSQL / MongoDB |

---

## Quick Start

```bash
# Cần Docker + Go >= 1.25
make up
```

Mở browser: `http://localhost:3000/admin`  
**Tài khoản mặc định:** `anhdt195` / `123` (SuperAdmin, tự tạo khi khởi động)

### Dev stack (với cli-netconf + ConfD)

```bash
# Clone cli-netconf cùng cấp
git clone https://github.com/DoTuanAnh2k1/cli-netconf ../cli-netconf

# Start mysql + mgt-svc + cli-netconf
docker compose -f docker-compose-dev.yaml up -d --build

# SSH vào cli-netconf
ssh anhdt195@localhost -p 2222
```

---

## Makefile

| Lệnh | Mô tả |
|---|---|
| `make up` | Start DB containers + chạy app local |
| `make up-docker` | Start toàn bộ trong Docker |
| `make down` | Stop tất cả |
| `make build` | Build binary → `bin/mgt-service` |
| `make build-docker` | Build Docker image |
| `make import FILE=data.txt` | Import data từ file text |
| `make dump` | Dump toàn bộ data trong database |
| `make metric` | Get runtime metrics |
| `make pprof-heap` | Heap profile |
| `make pprof-cpu` | CPU profile 30s |
| `make test` | Chạy tất cả tests |
| `make logs` | Tail app container logs |
| `make ps` | Show running containers |
| `make clean` | Xóa build artifacts + stop containers |

---

## Database

### Multi-DB support

| `DB_DRIVER` | Database |
|---|---|
| `mysql` | MySQL 8.0+ |
| `mariadb` | MariaDB 10.5+ |
| `postgres` | PostgreSQL 16+ |
| `mongodb` | MongoDB 7.0+ |

### Auto-migrate

MySQL / MariaDB / PostgreSQL tự động tạo/cập nhật tables khi app khởi động.

### Schema (db.sql)

| Bảng | Mô tả |
|---|---|
| `tbl_account` | Tài khoản: bcrypt password, `account_type` (0=SuperAdmin, 1=Admin, 2=Normal) |
| `cli_ne` | Network Element: `ne_name`, `namespace`, `conf_master_ip`, `conf_port_master_tcp`, `conf_mode`, ... |
| `cli_user_ne_mapping` | Gán NE cho user (FK → tbl_account, cli_ne) |
| `cli_operation_history` | Audit log: account, cmd_name, ne_name, scope, result |
| `cli_login_history` | Lịch sử đăng nhập |
| `cli_config_backup` | Metadata backup NETCONF (XML lưu trên disk) |

---

## Import Tool

```bash
make import FILE=deploy/local/sample_import.txt
```

Hoặc qua frontend: `http://localhost:3000/admin` → tab **Import**

### Định dạng file

```
[users]
username,password,email
admin,admin123,admin@vht.com
operator1,Pass@123,op1@vht.com

[nes]
ne_name,site_name,namespace,command_url,conf_mode,conf_master_ip,conf_port_master_ssh,conf_username,conf_password,description
HTSMF01,HCM,hcm-5gc,http://10.10.1.1:8080,SSH,10.10.1.1,22,admin,admin,HCM SMF Node 01

[user_nes]
username,ne_name
admin,HTSMF01
```

---

## Admin Frontend

Embedded tại `http://localhost:3000/admin` — song ngữ Tiếng Việt / English.

| Tab | Mô tả |
|---|---|
| Dashboard | Tổng quan users, NEs |
| Users | Tạo (có password + confirm password) / sửa thông tin / admin reset password / đổi mật khẩu của chính mình / vô hiệu hóa |
| Network Elements | Tạo / sửa (inline edit) / xóa NE |
| NE Mapping | Gán quyền cho user — trực tiếp bằng NE hoặc qua Group (radio Target: NE / Group). Bảng tách cột **Direct** (badge xanh dương) và **Via group** (badge xanh lá); Remove chỉ áp dụng cho direct |
| Groups | Tạo group → gán tập NE + tập user. User thuộc group thấy toàn bộ NE của group (hiển thị ở tab NE Mapping với badge xanh lá) |
| History | Lịch sử thao tác, filter theo scope (cli-config/ne-command/ne-config) và NE |
| Import | Upload file hoặc paste data để import hàng loạt |
| Guide | Hướng dẫn sử dụng (song ngữ) |

---

## SSH CLI (`cmd/ssh`)

Một bastion SSH chạy song song với mgt-svc, cho phép Admin/SuperAdmin thao tác
user/NE/group qua dòng lệnh và SSH proxy sang các NE CLI phía sau.

```bash
# Local (cần mgt-svc đã chạy)
MGT_SVC_BASE=http://localhost:3000 make run-ssh

# Hoặc trong docker-compose.local.yml: service cli-ssh-svc, expose :2223
docker compose -f docker-compose.local.yml up -d --build cli-ssh-svc
ssh anhdt195@localhost -p 2223
```

### Env

| Var | Mặc định | Mô tả |
|---|---|---|
| `SSH_CLI_LISTEN_ADDR` | `:2223` | Địa chỉ bind SSH |
| `SSH_CLI_HOST_KEY_PATH` | `/data/ssh_cli_host_key` | Ed25519 host key (tự sinh nếu chưa có) |
| `MGT_SVC_BASE` | *required* | Base URL của cli-mgt-svc, vd `http://cli-mgt-svc:3000` |
| `NE_CONFIG_SSH_ADDR` | — | Địa chỉ upstream cho mode `ne-config` (vd `cli-netconf-svc:22`) |
| `NE_COMMAND_SSH_ADDR` | — | Địa chỉ upstream cho mode `ne-command` |
| `LOG_LEVEL` | `info` | trace/debug/info/warn/error |

### Flow

1. SSH bằng username/password → forward `/aa/authenticate` để lấy JWT; reject nếu `account_type=2` (Normal).
2. Menu 3 mode (Tab cycle / autocomplete): `cli-config / ne-config / ne-command`. Tab lần đầu hiện gợi ý ngay **dưới prompt** (prompt được giữ nguyên); các lần Tab tiếp theo rotate qua từng candidate; phím bất kỳ khác Tab sẽ xoá dòng gợi ý.
3. **cli-config**: REPL với tab completion (verb → entity → field → enum value); mọi lệnh đều gọi HTTP sang mgt-svc kèm JWT.
4. **ne-config / ne-command**: mở SSH outbound với cùng username/password sang địa chỉ cấu hình, pipe stdin/stdout/stderr, forward window-change.

### Command grammar (cli-config)

Pairs là `field value` (space-separated, không `=`). Quote cho value có khoảng trắng.

```
show user|ne|group [<name|id>]
set user name <u> password <p> [email <e>] [full_name <f>] [account_type 1|2] [...]
set ne ne_name <n> namespace <ns> conf_master_ip <ip> conf_port_master_tcp <port> command_url <url> [...]
set group name <n> [description <d>]
update <entity> <name|id> <field> <value> [<field> <value> ...]
delete <entity> <name|id>
map user <u> ne <ne|id>           map user <u> group <g|id>         map group <g|id> ne <ne|id>
unmap ...                         (cùng shape)
help [command]
exit
```

Alias: `name` → `account_name` (user) hoặc `ne_name` (ne). `port` → `conf_port_master_tcp`. `ip` → `conf_master_ip`. `type` → `account_type`.

---

## Metrics & Profiling

```bash
make metric                    # GET /metrics — goroutines, heap, GC, CPU
make pprof-heap                # heap profile (interactive)
make pprof-cpu                 # CPU profile 30s
```

pprof server bật khi `PPROF_ENABLED=true` (default khi `make up`), listen `:6060`.

---

## API

Header: `Authorization: Basic <jwt_token>` (token từ `/aa/authenticate` đã chứa prefix `Basic`)

### Authentication
| Method | Path | Mô tả |
|---|---|---|
| `POST` | `/aa/authenticate` | Đăng nhập, lấy JWT |
| `POST` | `/aa/validate-token` | Kiểm tra token |
| `POST` | `/aa/change-password/` | Đổi mật khẩu của chính mình (yêu cầu old_password) |
| `POST` | `/aa/authenticate/user/set` | Tạo / kích hoạt lại user |
| `POST` | `/aa/authenticate/user/delete` | Vô hiệu hóa user |
| `GET`  | `/aa/authenticate/user/show` | Danh sách user kèm NE & role |
| `POST` | `/aa/authenticate/user/reset-password` | Admin reset mật khẩu |

### Network Element
| Method | Path | Mô tả |
|---|---|---|
| `POST` | `/aa/authorize/ne/create` | Tạo NE (ne_name bắt buộc) |
| `POST` | `/aa/authorize/ne/update` | Cập nhật NE (id bắt buộc) |
| `POST` | `/aa/authorize/ne/remove` | Xóa NE + cascade mappings |
| `POST` | `/aa/authorize/ne/set` | Gán NE cho user |
| `POST` | `/aa/authorize/ne/delete` | Xóa NE khỏi user |
| `GET`  | `/aa/authorize/ne/show` | Danh sách NE (5GC) |
| `GET`  | `/aa/list/ne` | NE user được truy cập (hợp của assign trực tiếp + qua group) |
| `GET`  | `/aa/list/ne/monitor` | NE monitor URL |

### Admin (Frontend API)
| Method | Path | Mô tả |
|---|---|---|
| `GET`  | `/aa/admin/user/list` | Danh sách user đầy đủ (không password) |
| `POST` | `/aa/admin/user/update` | Cập nhật user metadata |
| `GET`  | `/aa/admin/ne/list` | Danh sách NE đầy đủ |
| `POST` | `/aa/admin/ne/create` | Tạo NE (ne_name*, namespace*, conf_master_ip*, conf_port_master_tcp*, command_url*) |
| `POST` | `/aa/admin/ne/update` | Cập nhật NE |

### History
| Method | Path | Mô tả |
|---|---|---|
| `GET`  | `/aa/history/list` | Lịch sử lệnh (`?limit=N&scope=X&ne_name=Y`) |
| `POST` | `/aa/history/save` | Lưu bản ghi lịch sử |

### Import & Others
| Method | Path | Mô tả |
|---|---|---|
| `POST` | `/aa/import/` | Import hàng loạt (plain text body) |
| `POST` | `/aa/config-backup/save` | Lưu backup config XML |
| `GET`  | `/aa/config-backup/list` | Danh sách backup |
| `GET`  | `/aa/config-backup/{id}` | Lấy backup kèm XML |

---

## Deploy

### Docker Compose

```bash
# Dev (mysql + mgt-svc + cli-netconf)
docker compose -f docker-compose-dev.yaml up -d --build

# Production — public registry
docker compose -f deploy/docker/docker-compose.yml up -d

# Production — private registry (172.20.1.22)
docker compose -f deploy/docker/docker-compose-private.yaml up -d
```

### Published images

| Image | Dockerfile | Vai trò |
|---|---|---|
| `hsdfat/cli-mgt:<tag>` | `deploy/docker/Dockerfile` | Management API + Admin frontend |
| `hsdfat/cli-gate:<tag>` | `deploy/docker/Dockerfile_ssh` | SSH CLI bastion (port 2223) |

Tag là số 5 chữ số tăng dần mỗi lần build (không reuse). Build ví dụ:

```bash
docker build -f deploy/docker/Dockerfile_ssh -t hsdfat/cli-gate:21040 .
docker push hsdfat/cli-gate:21040
```

### Kubernetes

```bash
kubectl apply -f deploy/k8s/rbac.yaml
kubectl apply -f deploy/k8s/pvc.yaml
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/secret.yaml
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
```

---

## Cấu trúc thư mục

```
cli-mgt-svc/
├── cmd/
│   ├── main/                   # Entry point
│   └── import/                 # CLI import tool
├── models/
│   ├── config_models/          # Struct cấu hình
│   └── db_models/              # GORM models
├── pkg/
│   ├── handler/                # HTTP handlers, frontend (embedded), router
│   │   ├── middleware/         # Auth, CheckRole, RateLimit, CORS
│   │   └── response/          # JSON response helper
│   ├── repository/             # Data access layer
│   │   ├── mysql/
│   │   ├── postgres/
│   │   └── mongodb/
│   ├── service/                # Business logic + seed
│   ├── store/                  # DB interface + singleton
│   ├── testutil/               # Mock store
│   └── token/                  # JWT create/parse
├── deploy/
│   ├── docker/                 # Dockerfiles + docker-compose (public/private)
│   ├── k8s/                    # Kubernetes manifests
│   └── local/                  # .env.example, sample_import.txt, SQL scripts
├── docker-compose-dev.yaml     # Dev stack (mysql + mgt + cli-netconf)
├── db.sql                      # DDL schema
├── api.yaml                    # OpenAPI 3.0 spec
└── Makefile
```

---

## Tests

```bash
make test                                      # chạy toàn bộ
go test -cover ./...                           # kèm coverage
go test -coverprofile=cover.out ./pkg/sshcli && go tool cover -func=cover.out | tail -5
```

Coverage hiện tại (snapshot):

| Package | Coverage |
|---|---|
| `pkg/handler/response` | 97.4% |
| `pkg/bcrypt` | 83.3% |
| `pkg/tcpserver` | 76.3% |
| `pkg/token` | 74.1% |
| `pkg/sshcli` | 74.0% |
| `pkg/handler/middleware` | 65.4% |
| `pkg/leader` | 45.0% |
| `pkg/service` | 40.5% |
| `pkg/handler` | 33.6% |

`pkg/repository/*` (mongo/mysql/postgres) và `pkg/{config,logger,store,server}` hiện chưa có unit test.

---

## Tech stack

- **Go 1.25+** · **Chi** · **GORM** · **mongo-driver**
- **JWT (golang-jwt/jwt v5)** · **bcrypt** · **Logrus** · **godotenv**
- **k8s.io/client-go** (leader election) · **net/http/pprof**
- **Docker** multi-stage build · **MySQL 8** / **MariaDB** / **PostgreSQL 16** / **MongoDB 7**
