# Authorization Design — v2

> 3-tier permission model + two-layer group authorization.
> Tài liệu này mô tả chính xác luồng xử lý đang chạy trong code.

---

## Tổng quan kiến trúc

```
                  ┌─────────────────────────────────────┐
                  │           HTTP Request               │
                  └──────────────┬──────────────────────┘
                                 │
                  ┌──────────────▼──────────────────────┐
                  │  RouterCORS + RouterRealIP           │
                  │  pkg/handler/middleware/middleware.go │
                  └──────────────┬──────────────────────┘
                                 │
              ┌──────────────────┼──────────────────────┐
              │ Public           │ Authenticated         │
              │ /health          │                       │
              │ /aa/authenticate │  ┌────────────────┐   │
              │ /aa/export/*     │  │  Authenticate   │   │
              │ ...              │  │  middleware      │   │
              │                  │  └───────┬────────┘   │
              │                  │          │             │
              │                  │    ┌─────┼──────┐     │
              │                  │    │ GET │ Write │     │
              │                  │    │     │ ops   │     │
              │                  │    │     │       │     │
              │                  │    │     ▼       │     │
              │                  │    │ RequireAdmin│     │
              │                  │    │ middleware  │     │
              │                  │    └─────┬──────┘     │
              │                  │          │             │
              │                  │          ▼             │
              │                  │   Per-handler          │
              │                  │   super_admin check    │
              └──────────────────┴──────────────────────┘
```

---

## 1. Role — 3 tầng quản lý

Ba role cố định, không mở rộng thêm (`models/db_models/user.go:19-24`):

| Role | Level | Quyền quản lý |
|------|-------|----------------|
| `super_admin` | 3 | Full — CRUD mọi thứ |
| `admin` | 2 | CRUD users/NEs/commands/groups/policy/access-list. **Không** sửa/xoá/reset-pw tài khoản `super_admin` |
| `user` | 1 | Read-only — mọi POST/PUT/DELETE trả 403 |

### Lưu trữ

- **DB**: cột `role varchar(16) NOT NULL DEFAULT 'user'` trong bảng `user` (`db.sql:45`)
- **JWT**: claim `"role"` được mint khi login (`pkg/token/jwt.go:34`), default `"user"` nếu claim thiếu (`jwt.go:74`)
- **So sánh level**: `RoleLevel()` trả int (1/2/3) để so sánh hierarchy (`user.go:27-36`)

### Enforcement

**Tầng 1 — Middleware `RequireAdmin`** (`pkg/handler/middleware/authenticate.go:25-38`):

- Đặt sau `Authenticate` middleware trên tất cả write routes
- Check: role == `admin` hoặc `super_admin` → pass. Ngược lại → 403

**Tầng 2 — Per-handler super_admin protection** (`pkg/handler/user.go`):

Không có middleware `RequireSuperAdmin` riêng. Từng handler tự check:

| Handler | Logic |
|---------|-------|
| `HandlerCreateUser` (line 78) | Tạo user role=super_admin → caller phải là super_admin |
| `HandlerUpdateUser` (line 125) | Sửa tài khoản super_admin → caller phải là super_admin |
| `HandlerDeleteUser` (line 161) | Xoá tài khoản super_admin → caller phải là super_admin |
| `HandlerAdminResetPassword` (line 191) | Reset pw super_admin → caller phải là super_admin |

---

## 2. Authorization — hai lớp group

Mọi quyết định truy cập NE/command quy về một câu hỏi:

> **User `X` có được chạy command `Y` trên NE `Z` không?**

### Evaluator: `pkg/service/authorize.go:24-87`

```
Authorize(username, neID, commandID) → AuthorizeDecision
```

4 bước tuần tự, fail bất kỳ bước nào → deny:

| Bước | Check | Deny reason |
|------|-------|-------------|
| 1 | User tồn tại, `is_enabled=true`, `locked_at IS NULL` | `"user not found"` / `"user not active"` |
| 2 | `command.ne_id == neID` (command thuộc đúng NE đó) | `"command not registered on this NE"` |
| 3 | ∃ `ne_access_group` chứa cả user **và** NE (intersection) | `"no ne_access_group grants access to this NE"` |
| 4 | ∃ `cmd_exec_group` chứa cả user **và** command (intersection) | `"no cmd_exec_group grants this command"` |

**Response struct** (`authorize.go:7-15`):

```go
type AuthorizeDecision struct {
    Allowed            bool
    Reason             string
    UserExists         bool
    UserEnabled        bool
    NeReachable        bool   // bước 3 pass
    CommandOnNe        bool   // bước 2 pass
    CommandExecAllowed bool   // bước 4 pass
}
```

Intersection check dùng map O(n+m) (`authorize.go:89-103`).

### Tại sao tách hai group?

- **NE reachability** (operator có quyền chạm NE đó không?) — gắn theo site/region
- **Command eligibility** (command này được phép chạy không?) — gắn theo job function

Gộp thành 1 group → bùng nổ tổ hợp N×M hoặc phải đục lỗ khó audit.

### Schema group

**ne_access_group** (`models/db_models/groups.go:15-41`):

```
ne_access_group        (id, name UNIQUE, description)
  ├── ne_access_group_user  (group_id FK, user_id FK)   ON DELETE CASCADE
  └── ne_access_group_ne    (group_id FK, ne_id FK)     ON DELETE CASCADE
```

**cmd_exec_group** (`groups.go:43-71`):

```
cmd_exec_group         (id, name UNIQUE, description)
  ├── cmd_exec_group_user     (group_id FK, user_id FK)     ON DELETE CASCADE
  └── cmd_exec_group_command  (group_id FK, command_id FK)  ON DELETE CASCADE
```

Lookup functions tại `pkg/service/groups.go`:
- `ListNeAccessGroupsOfUser(userID)` / `ListNeAccessGroupsOfNE(neID)`
- `ListCmdExecGroupsOfUser(userID)` / `ListCmdExecGroupsOfCommand(commandID)`

---

## 3. Authentication — luồng login

**Entry**: `pkg/service/auth.go:19-71` → `Authenticate(username, password, clientIP)`

Các gate thực thi tuần tự:

### Gate 1: Access List (`pkg/service/access_list.go:46-87`)

Kiểm tra **trước** khi tìm user (chống username enumeration).

```
1. Blacklist pass: có entry match → DENY ngay
2. Whitelist pass:
   - Không có entry whitelist nào → ALLOW
   - Có entry cho match_type mà user có giá trị → user phải match ít nhất 1
   - Không có entry cho match_type đó → ALLOW (không giới hạn)
```

**match_type** (`models/db_models/auth.go:61-70`):

| `match_type` | So sánh với | Logic |
|--------------|-------------|-------|
| `username` | username | `strings.EqualFold` (case-insensitive) |
| `ip_cidr` | client IP | Exact match hoặc CIDR subnet match |
| `email_domain` | email user | Suffix match `@domain` |

Bảng `user_access_list`: `UNIQUE(list_type, match_type, pattern)`

### Gate 2: User state (`auth.go:32-41`)

- User phải tồn tại
- `is_enabled` phải `true` → nếu không: `ErrAccountDisabled`

### Gate 3: Lockout (`auth.go:40-47`)

- `locked_at != nil` **và** lockout chưa hết hạn → `ErrAccountLocked`
- Lockout hết hạn → tự clear `locked_at` + `login_failure_count`
- Config: `password_policy.lockout_minutes` (0 = tắt)

### Gate 4: Password expiry (`auth.go:48-50`)

- `password_expires_at != nil` **và** đã quá hạn → `ErrPasswordExpired`
- Config: `password_policy.max_age_days` (0 = không giới hạn)

### Gate 5: Password verify (`auth.go:52-61`)

- bcrypt compare
- Sai → `login_failure_count++`
  - Đạt threshold `max_login_failure` → set `locked_at = now`
- Config: `password_policy.max_login_failure` (0 = tắt lockout)

### Gate 6: Success (`auth.go:63-70`)

- Reset `login_failure_count = 0`, clear `locked_at`
- Stamp `last_login_at`
- Ghi `login_history`
- Mint JWT: `token.CreateToken(username, role)` → `"Basic " + HS256_JWT`

---

## 4. Password Policy

**Singleton row** `password_policy.id = 1` (`models/db_models/auth.go:14-25`):

| Field | Default | Ý nghĩa |
|-------|---------|----------|
| `min_length` | 8 | Độ dài tối thiểu |
| `max_age_days` | 0 | Thời hạn password (0 = vô hạn) |
| `require_uppercase` | false | Bắt buộc chữ hoa |
| `require_lowercase` | false | Bắt buộc chữ thường |
| `require_digit` | false | Bắt buộc số |
| `require_special` | false | Bắt buộc ký tự đặc biệt |
| `history_count` | 0 | Số password cũ không được dùng lại (0 = tắt) |
| `max_login_failure` | 0 | Ngưỡng lockout (0 = tắt) |
| `lockout_minutes` | 0 | Thời gian khoá (0 = tắt) |

**Validate**: `pkg/service/password_policy.go:44-75` — dùng khi tạo user, đổi mật khẩu, admin reset.

**Password history**: bảng `password_history` (`auth.go:29-38`), check reuse tại `pkg/service/user.go:137-144`.

---

## 5. Route protection

Định nghĩa tại `pkg/handler/router.go:24-132`.

### Public (không auth)

| Route | Chức năng |
|-------|-----------|
| `GET /health` | Health check |
| `GET /metrics` | Runtime metrics |
| `GET /admin` | Frontend SPA |
| `POST /aa/authenticate` | Login (có rate limit) |
| `POST /aa/validate-token` | Validate JWT |
| `POST /aa/history/save` | Lưu history |
| `GET /aa/export/{users,nes,commands}` | CSV export (auth qua `_token` query param) |

### Authenticated (middleware `Authenticate`)

Tất cả GET endpoints:
`/aa/users`, `/aa/nes`, `/aa/commands`, `/aa/ne-access-groups`, `/aa/cmd-exec-groups`,
`/aa/password-policy`, `/aa/access-list`, `/aa/authorize/check`, `/aa/history`,
`/aa/config-backup/*`, `/aa/change-password`

### Admin+ (middleware `Authenticate` + `RequireAdmin`)

Tất cả write endpoints:
`/aa/import/*`, user CRUD, NE CRUD, command CRUD, group management,
password policy update, access list management, config backup save.

---

## 6. Import / Export

### Export (`pkg/handler/import_export.go:40-105`)

| Endpoint | Auth | Output |
|----------|------|--------|
| `GET /aa/export/users` | JWT header hoặc `_token` query param | CSV: username, email, full_name, phone, role, is_enabled |
| `GET /aa/export/nes` | JWT header hoặc `_token` query param | CSV: namespace, ne_type, site_name, description, master_ip, master_port, ssh_username, command_url, conf_mode |
| `GET /aa/export/commands` | JWT header hoặc `_token` query param | CSV: ne_namespace, service, cmd_text, description |

> SSH password **không bao giờ** xuất ra CSV.

### Import (`import_export.go:109-266`)

| Endpoint | Auth | Behaviour |
|----------|------|-----------|
| `POST /aa/import/users` | admin+ | Multipart upload. Default password `"Changeme1!"` nếu thiếu. Skip duplicate. |
| `POST /aa/import/nes` | admin+ | Skip duplicate (case-insensitive namespace). |
| `POST /aa/import/commands` | admin+ | Resolve `ne_namespace` → `ne_id`. Skip duplicate. |

Response: `{ "created": N, "skipped": N, "errors": [...] }`

---

## 7. Schema highlights

- Tất cả FK dùng `ON DELETE CASCADE` — xoá group → xoá pivot; xoá NE → xoá commands + pivots
- `UNIQUE(ne_id, service, cmd_text)` trên `command` — chặn duplicate
- `UNIQUE(list_type, match_type, pattern)` trên `user_access_list` — chặn duplicate rule
- JWT format: `"Basic " + HS256_JWT`, secret key qua env `JWT_SECRET_KEY`

---

## 8. Test coverage

| File | Covers |
|------|--------|
| `pkg/service/authorize_test.go` | Deny no groups, allow both layers, deny locked, deny wrong NE |
| `pkg/service/auth_test.go` | Wrong password count, lockout threshold, disabled account, username blacklist |

Cả hai suite chạy với in-memory `testutil.MockStore` (`pkg/testutil/mock_store.go`).
