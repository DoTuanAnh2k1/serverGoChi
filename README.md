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
# Legacy entities
show user|ne|group [<field> <value> | <name|id>]
set user name <u> password <p> [email <e>] [full_name <f>] [account_type 1|2] [...]
set ne ne_name <n> namespace <ns> conf_master_ip <ip> conf_port_master_tcp <port> command_url <url> [ne_profile <p>] [...]
set group name <n> [description <d>]
update <entity> <name|id> <field> <value> [<field> <value> ...]
delete <entity> <name|id>
map user <u> ne <ne|id>           map user <u> group <g|id>         map group <g|id> ne <ne|id>
unmap ...                         (cùng shape)

# RBAC (docs/rbac-design.md)
show|set|update|delete ne-profile|command-def|command-group ...
map command-group <cg> command <cmd_def_id>
unmap command-group <cg> command <cmd_def_id>
allow  <group> <grant_type> <grant_value> [ne_scope <scope>] [service <svc>]
deny   <group> <grant_type> <grant_value> [ne_scope <scope>] [service <svc>]
revoke <group> <perm_id>

help [command [entity]]           # hoặc append '--help' / '-h' vào bất kỳ lệnh
exit
```

Alias: `name` → `account_name` (user) hoặc `ne_name` (ne). `port` → `conf_port_master_tcp`. `ip` → `conf_master_ip`. `type` → `account_type`.

**Help contextual**: thêm `--help` (hoặc `-h`) vào bất kỳ vị trí nào trên dòng lệnh để in help cho lệnh đó. Ví dụ: `set user --help`, `show ne --help`, `map --help`. Help topic được suy ra từ verb + entity (nếu có), fallback về verb-only nếu không có entry chuyên biệt.

**Show filters**: ngoài dạng legacy `show <entity> <name|id>`, `show` nhận thêm `<field> <value>`:

| Entity | Filter fields | Ghi chú |
|--------|---------------|--------|
| user | `name` \| `id` \| `email` \| `role` | `role` in bảng (SuperAdmin/Admin/Normal hoặc 0/1/2); các field khác in detail khi match duy nhất. |
| ne   | `name` \| `id` \| `site` \| `namespace` | `name` in bảng nếu trùng `ne_name` qua nhiều namespace; `site`/`namespace` luôn in bảng. |
| group | `name` \| `id` | In detail (users + ne_ids). |

Ví dụ:
```
show user role Admin              # list tất cả user role Admin
show user email alice@example.com # tra cứu theo email
show ne name HTSMF01              # liệt kê tất cả NE trùng tên (nếu có)
show ne site HN                   # tất cả NE ở site HN
show ne namespace tenant-a        # tất cả NE trong namespace tenant-a
show group name dev               # detail group
```

**Delete confirmation**: mọi lệnh `delete user|ne|group <target>` đều prompt `Delete <kind> "<target>"? [y/N]:`. Chỉ `y` / `yes` (không phân biệt hoa thường) mới thực hiện; input khác, dòng trống, Ctrl+C / Ctrl+D, hoặc lỗi đọc đều abort và in `aborted`.

**Re-enable disabled user**: khi gọi `set user name <u> password <p>` mà `<u>` đã tồn tại nhưng đang disable, server sẽ:
- Merge các trường non-empty từ request (email, full_name, phone, address, description, account_type) vào record cũ. Trường không truyền ⇒ giữ nguyên giá trị cũ. Password luôn được refresh (CLI luôn gửi).
- Validate lại required-fields cho admin trên kết quả merged (full_name + phone). Group mappings cũ được giữ; nếu truyền thêm `group_ids` thì add thêm (không ghi đè).
- Bật `is_enable=true` và `UpdateUser` — trả 201 Created.

**Email uniqueness**: `EnsureEmailUnique` bỏ qua tài khoản `is_enable=false`. Nghĩa là email thuộc về user đã disable coi như "free" — user mới hoặc flow re-enable có thể dùng lại email đó. Chỉ email của user đang active mới reject 400 "email already in use".

### RBAC — user X, NE Y, command Z

Kiến trúc đầy đủ ở [docs/rbac-design.md](docs/rbac-design.md). Mgt-svc lưu 5 bảng mới + 1 cột thêm:

| Bảng | Mục đích |
|------|----------|
| `cli_ne_profile` | Phân loại NE theo tập lệnh (SMF/AMF/UPF/generic-router/...) |
| `cli_command_def` | Registry lệnh, gắn với `(service, ne_profile)`. Pattern hỗ trợ exact + prefix + `*` cuối |
| `cli_command_group` | Bundle nhiều command-def cùng profile |
| `cli_command_group_mapping` | M:N giữa group và def |
| `cli_group_cmd_permission` | Rule allow/deny cho `cli_group` tại 1 `ne_scope` |
| `cli_ne.ne_profile_id` | Cột mới — FK xuống `cli_ne_profile` |

**Evaluation algorithm** (service/rbac.go): kết hợp AWS IAM (explicit deny > explicit allow > implicit deny) với Vault-style scope specificity (`ne:X` > `profile:Y` > `*`). Tại cấp scope cụ thể nhất có match, deny thắng allow. Scope broader chỉ thắng khi không có rule nào ở scope cụ thể hơn match.

**Phân quyền SSH login**:
- SuperAdmin / Admin → vào được cả 3 mode (`cli-config` / `ne-config` / `ne-command`).
- Normal user → vào được `ne-config` / `ne-command`, KHÔNG thấy `cli-config` trong menu. Việc lệnh gì được chạy trên NE nào do downstream service (`ne-config`, `ne-command`) tự verify bằng cách gọi mgt-svc lấy whitelist (`/aa/authorize/rbac/effective`) rồi match local, hoặc per-command (`/aa/authorize/rbac/check-command`).

**Endpoint whitelist cho downstream service**:
```
GET  /aa/authorize/rbac/effective              # full list NE + rules để cache
POST /aa/authorize/rbac/check-command          # realtime { service, command, ne_id } → { allowed, reason }
```

**CRUD qua CLI** (admin only):
```
show ne-profile                                   set ne-profile name SMF description "..."
show command-def [service X] [ne_profile Y]       set command-def service ne-command ne_profile SMF pattern "get subscriber" category monitoring
show command-group [service X] [ne_profile Y]     set command-group name smf-subscriber-ops ne_profile SMF
map command-group smf-subscriber-ops command 10   unmap command-group smf-subscriber-ops command 10
allow team-smf-l1 command-group smf-subscriber-ops ne_scope profile:SMF
deny  team-smf-l2 pattern "delete *"              ne_scope ne:SMF-01
show group team-smf-l1                             # list all permissions via GET /aa/group/{id}/cmd-permissions
revoke team-smf-l1 <perm_id>                       # remove a specific rule
update ne HTSMF01 ne_profile SMF                   # assign NE profile to an existing NE
```

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
| Method | Path | Auth | Mô tả |
|---|---|---|---|
| `GET`  | `/aa/history/list` | JWT | Lịch sử lệnh (`?limit=N&scope=X&ne_name=Y`) |
| `POST` | `/aa/history/save` | **không yêu cầu JWT** | Lưu bản ghi lịch sử. Caller truyền `account` trong body (nếu có JWT trong context thì ưu tiên username từ JWT; thiếu cả hai fallback `unknown`). |

### RBAC (docs/rbac-design.md)
| Method | Path | Mô tả |
|---|---|---|
| `GET`    | `/aa/ne-profile/list`                       | Liệt kê profile |
| `POST`   | `/aa/ne-profile/create`                     | Tạo profile |
| `POST`   | `/aa/ne-profile/update`                     | Sửa profile |
| `DELETE` | `/aa/ne-profile/{id}`                       | Xoá profile |
| `POST`   | `/aa/ne/{ne_id}/profile`                    | Gán profile cho NE |
| `GET`    | `/aa/command-def/list[?service=&ne_profile=&category=]` | Danh sách command-def |
| `POST`   | `/aa/command-def/create`                    | Tạo command-def |
| `POST`   | `/aa/command-def/update`                    | Sửa command-def |
| `POST`   | `/aa/command-def/import`                    | Bulk import (array) |
| `DELETE` | `/aa/command-def/{id}`                      | Xoá command-def |
| `GET`    | `/aa/command-group/list[?service=&ne_profile=]` | Danh sách command-group |
| `POST`   | `/aa/command-group/create`                  | Tạo command-group |
| `POST`   | `/aa/command-group/update`                  | Sửa command-group |
| `DELETE` | `/aa/command-group/{id}`                    | Xoá command-group |
| `GET`    | `/aa/command-group/{id}/commands`           | Liệt kê lệnh trong group |
| `POST`   | `/aa/command-group/{id}/commands`           | Thêm lệnh vào group |
| `DELETE` | `/aa/command-group/{id}/commands/{cmd_id}`  | Gỡ lệnh khỏi group |
| `GET`    | `/aa/group/{id}/cmd-permissions`            | Liệt kê permission của group |
| `POST`   | `/aa/group/{id}/cmd-permissions`            | Thêm allow/deny rule |
| `DELETE` | `/aa/group/{id}/cmd-permissions/{perm_id}`  | Xoá rule |
| `GET`    | `/aa/authorize/rbac/effective`              | Full NE + rule của caller (để downstream cache) |
| `POST`   | `/aa/authorize/rbac/check-command`          | `{ service, command, ne_id }` → `{ allowed, reason }` |

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
