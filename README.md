# cli-mgt-svc (v2)

Management service cho hệ thống CLI viễn thông 5G — Go + Chi.

Toàn hệ thống trả lời một câu hỏi:

> **User X có được execute command Y trên NE Z không?**

Allowed ⇔ user *enabled* + *not locked* + *không nằm blacklist* (và nếu có whitelist thì phải match)
**AND** command Y đã đăng ký trên NE Z (exact text, không pattern)
**AND** X ∈ một `ne_access_group` chứa Z
**AND** X ∈ một `cmd_exec_group` chứa Y.

### Phân quyền 3 tầng

| Role | Quyền |
|------|-------|
| `super_admin` | Toàn quyền. Quản lý mọi tài khoản kể cả super_admin khác |
| `admin` | Cấu hình mọi thứ, nhưng **không** sửa/xóa/reset-pw tài khoản `super_admin` |
| `user` | Chỉ xem (read-only). Tất cả endpoint write trả 403 |

- JWT token chứa claim `role`.
- Mọi GET endpoint: bất kỳ authenticated user.
- Mọi POST/PUT/DELETE endpoint: yêu cầu `admin` hoặc `super_admin`.
- Seed mặc định: `admin/admin` (super_admin), `anhdt195/123` (super_admin).

---

## Mô hình dữ liệu

| Bảng | Ý nghĩa |
|---|---|
| `user` | tài khoản — có `role` (super_admin / admin / user) |
| `ne` | `namespace` (unique, ví dụ `htsmf01`) + `ne_type` (category, ví dụ `SMF`) |
| `command` | exact text, unique theo `(ne_id, service, cmd_text)`; `service` ∈ {`ne-config`, `ne-command`} |
| `ne_access_group` + pivots | M:N user ↔ NE — quyết định ai *đến được* NE nào |
| `cmd_exec_group` + pivots | M:N user ↔ command — quyết định ai *chạy được* lệnh nào |
| `password_policy` | singleton (id=1): min length, history, lockout, require rules |
| `password_history` | dùng cho quy tắc no-reuse-last-N |
| `user_access_list` | blacklist / whitelist theo `username`, `ip_cidr`, `email_domain` |
| `operation_history`, `login_history` | audit trail |
| `config_backup` | lưu XML backup từ NETCONF commit (kept as-is) |

Schema đầy đủ ở [db.sql](db.sql) và `models/db_models/`.

---

## Tính năng

| Nhóm | Ghi chú |
|---|---|
| Authentication | `POST /aa/authenticate` → JWT (`sub` = username, `role` = super_admin/admin/user) |
| Access-list gate | Blacklist chặn; whitelist theo từng `match_type`: nếu có entry nào của `match_type` X thì user bắt buộc match ít nhất 1 |
| Lockout | Tự khoá sau `max_login_failure`, tự mở sau `lockout_minutes` (cả 2 = 0 để tắt) |
| Password policy | Singleton global. `PUT /aa/password-policy` ghi đè toàn bộ |
| Password history | Chặn reuse N password gần nhất khi đổi password |
| User CRUD | `/aa/users` + `POST /aa/users/{id}/reset-password` |
| NE / Command | `/aa/nes`, `/aa/commands` (filter `?ne_id=&service=`) |
| Group membership | `/aa/{ne-access-groups,cmd-exec-groups}/{id}/{users,nes,commands}` |
| Authorize check | `POST /aa/authorize/check {username, ne_id, command_id}` → `{allowed, reason, trace}` |
| History | `/aa/history?limit=&scope=&ne_namespace=&account=`; `POST /aa/history/save` không cần JWT (audit sink) |
| Config backup | `POST /aa/config-backup/save`, `GET /aa/config-backup/{list,{id}}` |
| Import/Export | CSV import/export cho Users, NEs, Commands. Export: `GET /aa/export/{users,nes,commands}`. Import: `POST /aa/import/{users,nes,commands}` (admin+, multipart upload) |
| Admin frontend | Embed tại `/admin`; 11 tab; paging 20/page; EN+VI; role-aware UI (user=read-only, admin ẩn nút sửa/xóa super_admin); nút Export/Import CSV trên mỗi tab |
| Permission | 3 tầng: `super_admin` > `admin` > `user`. GET = any auth, write = admin+. Xem bảng phân quyền ở trên |
| Multi-DB | MySQL, PostgreSQL (GORM auto-migrate) + MongoDB (index plan + counter collection cho sequential IDs) |
| Test | Service-layer: evaluator (4 case), authenticate (happy path / wrong pw failure-count / lockout / disabled / blacklist), policy apply chain |

---

## Quick Start

```bash
# Cần Docker + Go >= 1.25
make up
```

Frontend: `http://localhost:3000/admin`
API docs (Swagger UI): `http://localhost:3000/docs`
Tài khoản mặc định (seed khi DB trống — đổi ngay):
- `admin / admin` (super_admin)
- `anhdt195 / 123` (super_admin)

### Local dev

```bash
cp .env.example .env    # chỉnh DB driver + secret
make build
./build/mgt
```

---

## HTTP Surface (rút gọn)

```text
POST /aa/authenticate           { username, password } → { token }
POST /aa/validate-token         { token }              → { username }
POST /aa/change-password        (auth) { old_password, new_password }

GET/POST/PUT/DELETE /aa/users[/id]
POST                /aa/users/{id}/reset-password
GET                 /aa/users/{id}/executable-commands  → [Command]
GET                 /aa/users/{id}/accessible-nes       → [NE]
GET/POST/PUT/DELETE /aa/nes[/id]
GET                 /aa/nes/{id}/authorized-users       → [User]
GET/POST/PUT/DELETE /aa/commands[/id]  (GET ?ne_id=&service=)
GET                 /aa/commands/{id}/authorized-users   → [User]

GET/POST/PUT/DELETE /aa/ne-access-groups[/id]
GET/POST/DELETE     /aa/ne-access-groups/{id}/users
GET/POST/DELETE     /aa/ne-access-groups/{id}/nes

GET/POST/PUT/DELETE /aa/cmd-exec-groups[/id]
GET/POST/DELETE     /aa/cmd-exec-groups/{id}/users
GET/POST/DELETE     /aa/cmd-exec-groups/{id}/commands

GET/PUT             /aa/password-policy
GET/POST/DELETE     /aa/access-list
POST                /aa/authorize/check

GET                 /aa/export/{users,nes,commands}     (CSV download)
POST                /aa/import/{users,nes,commands}     (CSV upload, admin+)

GET                 /aa/history
POST                /aa/history/save                   (public — proxy audit push)
POST/GET            /aa/config-backup/{save,list,{id}}
```

Auth header: `Authorization: Basic <jwt>` (prefix mang tính legacy, giữ nguyên).

---

## Những gì đã bị bỏ so với v1

| Đã bỏ | Thay bằng |
|---|---|
| `tbl_account`, `account_type` (0/1/2) | `user.role` (super_admin / admin / user) + membership groups |
| NE profile (SMF/AMF/...) như first-class | `ne.ne_type` — chỉ là string label |
| `command_def` với pattern + risk_level + category | `command` exact text |
| `command_group` + `group_cmd_permission` + `ne_scope` (allow/deny) + AWS-IAM evaluator | `cmd_exec_group` (M:N user↔command), flat allow |
| Per-group password policy | Singleton global policy |
| Mgt-permission (per-group resource×action) | 3-tier role: `user` read-only, `admin` write, `super_admin` full; authorize check cho NE/command |
| SSH CLI bastion (`cmd/ssh`, `pkg/sshcli`) | Hiện chưa có trong v2 — HTTP API + frontend là interface đầy đủ. CLI sẽ được viết lại sau |
| Bulk-import tool (`cmd/import`) | CSV import/export qua API (`/aa/import/*`, `/aa/export/*`) + nút trên frontend |

---

## Development

```bash
go build ./...
go test ./...               # include evaluator + auth tests
go vet ./...
```

Test cần chạy DB thật: không có trong bộ này (v1 tests dùng `testutil.MockStore` — in-memory `DatabaseStore`).

Mock store có hook field (`DeleteHistoryBeforeFn`, `GetDailyOperationHistoryFn`) để inject spy trong `pkg/leader` — xem `pkg/testutil/mock_store.go`.

---

## Client Libraries

| Ngôn ngữ | File | Typed models | API coverage |
|-----------|------|-------------|-------------|
| **Java** | [`examples/java/MgtServiceClient.java`](examples/java/MgtServiceClient.java) | 9 classes (User, NE, Command, Group, ...) | 100% — all endpoints |
| **C** | [`examples/c/mgt_client.{h,c}`](examples/c/) | 6 structs | 100% — 41 functions |

Cả 2 client đều zero-dependency (Java: JDK 11+, C: libcurl) và hỗ trợ đầy đủ role field.

---

## Docker

Image: `hsdfat/cli-mgt` (tag tăng dần 5 chữ số, không reuse).
`Dockerfile` là multi-stage build — `go build` trong stage 1, binary nhẹ ở stage 2.

```bash
docker build -t hsdfat/cli-mgt:NNNNN .
docker push hsdfat/cli-mgt:NNNNN
```

Xem [deployments/k8s/](deployments/k8s/) cho manifest tham chiếu.
