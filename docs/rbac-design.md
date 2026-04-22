# RBAC Design Reference — mgt-svc

## Mục lục

1. [Hệ thống hiện tại và vấn đề](#1-hệ-thống-hiện-tại-và-vấn-đề)
2. [Tham khảo các hệ thống lớn](#2-tham-khảo-các-hệ-thống-lớn)
3. [So sánh các mô hình](#3-so-sánh-các-mô-hình)
4. [Thiết kế đề xuất cho mgt-svc](#4-thiết-kế-đề-xuất-cho-mgt-svc)
5. [Database Schema](#5-database-schema)
6. [Flow chi tiết](#6-flow-chi-tiết)
7. [API Contract](#7-api-contract)
8. [Migration Plan](#8-migration-plan)

---

## 1. Hệ thống hiện tại và vấn đề

### Kiến trúc hiện tại

```
cli-ne-config ──┐
                ├──▶ mgt-svc ──▶ DB
cli-ne-command ─┘
```

- **mgt-svc** (project này): quản lý user, group, NE, cấp JWT token
- **cli-ne-config**: NETCONF proxy — gọi mgt-svc để authen, lấy danh sách NE được phép
- **cli-ne-command**: SSH proxy — gọi mgt-svc để authen, lấy danh sách NE được phép

### Vấn đề

| Vấn đề | Chi tiết |
|--------|----------|
| Role quá thô | Chỉ có 3 mức cứng: SuperAdmin (0), Admin (1), Normal (2) |
| Không quản lý quyền lệnh | cli-ne-config/cli-ne-command không biết user được chạy lệnh gì |
| Không có password policy | Có field `ForceChangePass`, `LastChangePass` nhưng không có logic hết hạn |
| Group chỉ map NE | Group hiện tại chỉ dùng để gom NE, không gắn quyền |
| JWT quá đơn giản | Chỉ chứa `aud: "admin"/"user"`, không mang thông tin group hay permission |

---

## 2. Tham khảo các hệ thống lớn

### 2.1. Cisco IOS — Privilege Levels + Parser Views

**Mô hình**: 16 cấp (0-15), cumulative (cấp cao kế thừa cấp thấp).

```
Level 0:  logout, enable, help
Level 1:  show, ping, traceroute (user EXEC)
Level 2-14: custom — admin gán lệnh vào từng level
Level 15: full access (privileged EXEC)
```

**Gán lệnh vào level**:
```
privilege exec level 7 show running-config
privilege exec level 10 configure terminal
```

User ở level 7 chạy được mọi lệnh từ level 0-7. Không có deny — chỉ có additive.

**Parser Views** (RBAC nâng cao):
- Tạo "view" = tập lệnh được phép
- `commands exec include show interface *` — cho phép
- `commands exec exclude show running-config` — cấm
- **Superview** = gom nhiều view lại, user được union tất cả

**AAA/TACACS+**: Mỗi lệnh user gõ được gửi đến TACACS+ server để xét duyệt realtime.
Server có rule: `cmd = show { permit .* }`, `cmd = reload { deny .* }`.
**First-match wins**.

**Password**: Cơ bản — `min-length`, `login block-for`. Hết hạn/history phải dùng AAA server bên ngoài.

### 2.2. Juniper JUNOS — Login Classes + Permission Flags

**Mô hình**: Mỗi user thuộc 1 login class. Class định nghĩa:

1. **Permission flags** (bit flags, additive):

| Flag | Quyền |
|------|-------|
| `view` | show commands |
| `configure` | vào config mode |
| `interface` / `interface-control` | xem / sửa interface |
| `routing` / `routing-control` | xem / sửa routing |
| `security` / `security-control` | xem / sửa security |
| `system` / `system-control` | xem / sửa system |
| `admin` / `admin-control` | quản lý user / restart |
| `shell` | access local shell |
| `all` | tất cả |

2. **allow-commands / deny-commands** (regex):

```
class NOC-L2 {
    permissions [ view interface network configure ];
    allow-commands "(show .*)|(ping .*)|(configure terminal)";
    deny-commands "(request system reboot)|(file delete .*)";
}
```

**Evaluation**: deny-commands **luôn thắng** allow-commands. Lệnh phải:
1. Được phép bởi permission flags, VÀ
2. Match allow-commands (nếu có), VÀ
3. KHÔNG match deny-commands

**Password**: Cơ bản — `min-length`, `min-changes`. Hết hạn phải dùng RADIUS/TACACS+.

### 2.3. Huawei VRP — 4 Command Levels

**Mô hình**: 4 cấp cố định, đơn giản nhất.

| Level | Tên | Ví dụ lệnh |
|-------|-----|-------------|
| 0 | Visit | `ping`, `tracert`, `telnet` |
| 1 | Monitor | `display` (= show), `debugging` |
| 2 | Configure | `interface`, routing config, service config |
| 3 | Manage | `file ops`, `user mgmt`, `reset saved-config`, `reboot` |

Mỗi lệnh có level mặc định, admin có thể đổi:
```
command-privilege level 1 view system display current-configuration
```

User level 2 chạy được level 0-2. **Không có deny** — purely hierarchical.

**Password policy** (tốt nhất trong nhóm network vendor):
- `user-password min-len 8`
- `user-password complexity-check` (upper + lower + digit + special)
- `user-password expire 90` (hết hạn sau 90 ngày)
- `user-password history-record 12` (không trùng 12 pass gần nhất)
- `aaa local-user lockout period 30 count 5` (khoá 30 phút sau 5 lần sai)

### 2.4. Nokia SR OS — Profile Entries (First-Match)

**Mô hình**: Mỗi user có 1 profile. Profile = danh sách entry đánh số, xét từ trên xuống.

```
profile "noc-operator"
    default-action deny-all
    entry 10
        match "show"
        action permit
    entry 20
        match "tools dump"
        action permit
    entry 30
        match "configure router"
        action deny
    entry 40
        match "admin save"
        action permit
```

**Evaluation**: **First match wins** — entry nào match trước thì dùng action của nó. Không match entry nào → `default-action`.

**Ưu điểm**: Prefix matching đơn giản, dễ debug vì biết chính xác entry nào match.
**Nhược điểm**: Thứ tự entry quan trọng, dễ sai nếu không cẩn thận.

**Password**: Đầy đủ — `min-length`, `complexity`, `aging`, `history-size`, `lockout`.

### 2.5. Linux sudo — Cmnd_Alias + Last-Match

**Mô hình**: Flat rules trong `/etc/sudoers`.

```
# Định nghĩa nhóm lệnh
Cmnd_Alias NETWORKING = /sbin/ip, /usr/sbin/tcpdump, /usr/bin/ss
Cmnd_Alias SERVICES   = /usr/bin/systemctl restart *, /usr/bin/systemctl status *
Cmnd_Alias DANGEROUS  = /usr/bin/rm, /sbin/reboot, /sbin/shutdown

# Gán cho group
%noc-team  ALL=(root) NETWORKING, SERVICES, !DANGEROUS
```

**Evaluation**: **Last match wins** — khác biệt lớn so với các hệ thống khác. Rule cuối cùng match sẽ quyết định.

**Password** (qua PAM):
- `pam_pwquality`: `minlen=12`, `dcredit=-1` (ít nhất 1 digit), `ucredit=-1`, `ocredit=-1`
- `pam_unix`: `remember=12` (không trùng 12 pass cũ)
- `/etc/login.defs`: `PASS_MAX_DAYS=90`, `PASS_MIN_DAYS=7`, `PASS_WARN_AGE=14`

### 2.6. HashiCorp Vault — Path-Based Policies

**Mô hình**: Quyền gắn theo **đường dẫn** (path) với **capabilities**.

```hcl
# Đọc mọi thứ trong myapp
path "secret/data/myapp/*" {
  capabilities = ["read", "list"]
}

# Full quyền trên staging
path "secret/data/staging/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Cấm tuyệt đối zone production
path "secret/data/production/*" {
  capabilities = ["deny"]
}
```

Capabilities: `create`, `read`, `update`, `delete`, `list`, `sudo`, `deny`, `patch`.

**Evaluation**:
1. Token có nhiều policies → capabilities được **union** (cộng dồn)
2. `deny` **luôn thắng** — nếu bất kỳ policy nào deny, không policy nào có thể override
3. Path cụ thể hơn thắng path chung hơn

**Templated policies**: `path "secret/data/users/{{identity.entity.name}}/*"` — mỗi user chỉ access được data của mình.

### 2.7. Kubernetes RBAC — Resources + Verbs (Additive Only)

**Mô hình**: Role định nghĩa `resources` + `verbs`, bind vào user/group.

```yaml
kind: Role
rules:
- apiGroups: [""]
  resources: ["pods", "pods/log"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list"]
  resourceNames: ["nginx"]    # chỉ deployment tên "nginx"
```

Verbs: `get`, `list`, `watch`, `create`, `update`, `patch`, `delete`, `deletecollection`.

**Evaluation**: **Additive only — KHÔNG CÓ DENY**. Union tất cả Role/ClusterRole. Không grant = không có quyền.

**Ưu điểm**: Đơn giản, dễ reasoning.
**Nhược điểm**: Không thể tạo exception ("tất cả trừ X").

### 2.8. AWS IAM — Effect/Action/Resource/Condition

**Mô hình**: Policy documents với statements.

```json
{
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["ec2:Describe*", "ec2:Get*"],
      "Resource": "*"
    },
    {
      "Effect": "Deny",
      "Action": "ec2:TerminateInstances",
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "ec2:ResourceTag/Environment": "production"
        }
      }
    }
  ]
}
```

**Evaluation** (quan trọng nhất — industry standard):

```
1. Mặc định: DENY (implicit deny)
2. Nếu có Explicit Deny → DENY (không gì override được)
3. Nếu có Explicit Allow → ALLOW
4. Nếu không match gì → DENY (implicit deny)

Explicit Deny  >  Explicit Allow  >  Implicit Deny
```

**Wildcards**: `s3:Get*` match `s3:GetObject`, `s3:GetBucketPolicy`...

**Conditions**: Thêm điều kiện — IP source, MFA, tag, thời gian...

**Password policy**: Đầy đủ — `min-length`, complexity, `expiration (1-1095 days)`, `history (1-24)`, `allow self-change`.

### 2.9. FreeIPA — HBAC + Sudo Rules

**Mô hình**: Tách riêng 2 lớp:
1. **HBAC** (Host-Based Access Control): Ai được SSH vào máy nào
2. **Sudo Rules**: Trên máy đó, được chạy lệnh gì

```
# HBAC: team noc-team được SSH vào nhóm máy network-devices
ipa hbacrule-add "allow-noc-ssh"
ipa hbacrule-add-user "allow-noc-ssh" --groups=noc-team
ipa hbacrule-add-host "allow-noc-ssh" --hostgroups=network-devices

# Sudo: trên máy đó, được chạy nhóm lệnh networking-commands
ipa sudorule-add "noc-networking"
ipa sudorule-add-allow-command "noc-networking" --sudocmdgroups="networking-commands"
ipa sudorule-add-deny-command "noc-networking" --sudocmds="/sbin/reboot"
```

**Sudo command groups** (rất hay):
```
ipa sudocmdgroup-add "networking-commands"
ipa sudocmdgroup-add-member "networking-commands" --sudocmds="/sbin/ip"
ipa sudocmdgroup-add-member "networking-commands" --sudocmds="/usr/sbin/tcpdump"
```

**Password policy** (per group — đặc biệt hữu ích):
```
ipa pwpolicy-mod --maxlife=90 --history=12 --minlength=12 --minclasses=3 \
  --maxfail=5 --failinterval=60 --lockouttime=600
```
- `maxlife`: hết hạn (ngày)
- `history`: không trùng N pass cũ
- `minclasses`: ít nhất N loại ký tự (lower, upper, digit, special)
- `maxfail` + `failinterval` + `lockouttime`: lockout policy

---

## 3. So sánh các mô hình

### 3.1. Bảng so sánh tổng hợp

| Tiêu chí | Cisco | JUNOS | Huawei | Nokia | sudo | Vault | K8s | AWS IAM | FreeIPA |
|----------|-------|-------|--------|-------|------|-------|-----|---------|---------|
| **Cấu trúc** | Hierarchy 0-15 | Flat class + flags | Hierarchy 0-3 | Ordered entries | Flat rules | Path-based | Resource+verb | Policy docs | Rules+groups |
| **Deny** | Không (views có) | Deny > Allow | Không | First-match | Last-match | Deny wins | Không có | Deny wins | Allowlist only |
| **Wildcard** | Không | Regex | Không | Prefix | Glob | Glob | `*` | Glob | Không |
| **Command grouping** | Superview | Permission flags | Command level | Profile entries | Cmnd_Alias | Path prefix | apiGroups | Action prefix | sudocmdgroup |
| **Password built-in** | Minimal | Minimal | Tốt | Tốt | Via PAM | Token TTL | Không | Tốt | Rất tốt |
| **Phù hợp mgt-svc** | Trung bình | Cao | Trung bình | Cao | Trung bình | Trung bình | Thấp | Cao (logic) | Cao (concept) |

### 3.2. Pattern nào phù hợp nhất cho mgt-svc?

Xét context: mgt-svc là **trung tâm cấp phát quyền** cho cli-ne-config và cli-ne-command.
User không gõ lệnh trực tiếp trên mgt-svc mà trên các service khác.
mgt-svc cần trả lời: "User X được chạy lệnh Y trên NE Z không?"

**Pattern phù hợp nhất = Kết hợp**:

| Lấy từ | Concept | Áp dụng |
|--------|---------|---------|
| **FreeIPA** | Tách HBAC (NE access) + Sudo rules (command access) | Group → NE mapping + Group → Command mapping |
| **FreeIPA** | Sudo command groups | Gom lệnh thành nhóm (monitoring, configuration, admin...) |
| **Juniper** | Permission flags + deny patterns | Category-based allow + explicit deny patterns |
| **AWS IAM** | Explicit Deny > Allow > Implicit Deny | Evaluation logic an toàn nhất |
| **Huawei** | Password policy on-box | Built-in password expiration, complexity, history, lockout |
| **Nokia** | Prefix-based matching | Command matching đơn giản, không dùng regex |

---

## 4. Thiết kế đề xuất cho mgt-svc

### 4.1. Vấn đề cốt lõi — Tại sao cần NE-scoped commands?

Các NE khác nhau có tập lệnh hoàn toàn khác nhau:

```
SMF-01 (Session Management Function):
  - get subscriber        ← chỉ SMF có
  - get session            ← chỉ SMF có
  - show pdu-session       ← chỉ SMF có

AMF-01 (Access and Mobility Management Function):
  - show sctp              ← chỉ AMF có
  - show ue                ← chỉ AMF có
  - show registration      ← chỉ AMF có

UPF-01 (User Plane Function):
  - show pfcp-session      ← chỉ UPF có
  - show gtp-tunnel        ← chỉ UPF có
```

**Nếu chỉ allow `get *` ở mức global**:
- User được `get subscriber` → OK
- User cũng được `get session` → KHÔNG MUỐN

**→ Command permission phải gắn với NE hoặc nhóm NE cụ thể, không phải global.**

### 4.2. Giải pháp — NE Profile (phân loại NE theo function)

NE hiện tại có `system_type` (VD: "5GC") — nhưng cùng "5GC" mà AMF, SMF, UPF 
có lệnh khác nhau. Cần thêm **NE Profile** để phân loại NE theo chức năng:

```
NE Profile = loại NE theo tập lệnh mà nó hỗ trợ

Ví dụ:
  Profile "SMF"  → ne_name: SMF-01, SMF-02, SMF-HCM-01
  Profile "AMF"  → ne_name: AMF-01, AMF-HN-01
  Profile "UPF"  → ne_name: UPF-01, UPF-02
  Profile "generic-router" → ne_name: PE-01, PE-02
```

Mỗi NE thuộc 1 profile. Command definitions gắn với profile.
Permission gán theo: **Group → (NE scope) → (Command)**

### 4.3. Kiến trúc tổng quan

```
┌──────────────────────────────────────────────────────────────────────┐
│                              mgt-svc                                 │
│                                                                      │
│  ┌──────────┐    ┌───────────┐    ┌────────────────┐                 │
│  │  User    │◄──►│   Group   │◄──►│   NE           │                 │
│  │          │ M:N│           │ M:N│                 │                 │
│  └──────────┘    └─────┬─────┘    │ + ne_profile ──┼──► ┌──────────┐ │
│                        │          └────────────────┘    │NE Profile│ │
│                        │                                └────┬─────┘ │
│                        │ M:N                                 │       │
│               ┌────────┴──────────────┐                      │       │
│               │ Group Command Perm    │    ┌─────────────────┘       │
│               │                       │    │                         │
│               │ + ne_scope ───────────┼────┘  ← per NE/Profile/*    │
│               │ + effect (allow/deny) │                              │
│               │ + grant (group/cmd/…) │                              │
│               └────────┬──────────────┘                              │
│                        │ references                                  │
│               ┌────────┴─────────┐                                   │
│               │  Command Def     │                                   │
│               │                  │                                   │
│               │ + ne_profile ────┼──── lệnh thuộc loại NE nào       │
│               │ + pattern        │                                   │
│               │ + category       │                                   │
│               │                  │                                   │
│               │  ┌────────────┐  │                                   │
│               │  │ Cmd Group  │  │  ← gom nhiều cmd cùng profile    │
│               │  └────────────┘  │                                   │
│               └──────────────────┘                                   │
│                                                                      │
│  ┌──────────────────┐    ┌──────────────────┐                        │
│  │ Password Policy  │    │  Mgt Permission  │                        │
│  │ (per group)      │    │  (quản trị svc)  │                        │
│  └──────────────────┘    └──────────────────┘                        │
└──────────────────────────────────────────────────────────────────────┘
```

### 4.4. Ba lớp quyền tách biệt

```
Lớp 1: NE Access
  "User thuộc group nào? Group đó có NE nào?"
  → Giữ nguyên: Group → NE mapping (existing)

Lớp 2: Command Permission (có NE scope)
  "Trên NE đó (hoặc loại NE đó), user được chạy lệnh gì?"
  → MỚI: Group → Command Permission → gắn với ne_scope

Lớp 3: Mgt Permission
  "User được quản trị gì trên mgt-svc?"
  → MỚI: Group → Mgt Permission (thay thế account_type)
```

### 4.5. Command Registry — gắn với NE Profile

**Command Definition** — mỗi lệnh gắn với profile NE cụ thể:

```
# Lệnh chung cho mọi NE (profile = "*")
{ service: "ne-command", ne_profile: "*",   pattern: "show version",       category: "monitoring" }
{ service: "ne-command", ne_profile: "*",   pattern: "show running-config", category: "monitoring" }
{ service: "ne-command", ne_profile: "*",   pattern: "ping *",             category: "monitoring" }
{ service: "ne-command", ne_profile: "*",   pattern: "reload",             category: "admin", risk: 2 }

# Lệnh riêng SMF
{ service: "ne-command", ne_profile: "SMF", pattern: "get subscriber",     category: "monitoring" }
{ service: "ne-command", ne_profile: "SMF", pattern: "get session",        category: "monitoring" }
{ service: "ne-command", ne_profile: "SMF", pattern: "show pdu-session *", category: "monitoring" }
{ service: "ne-command", ne_profile: "SMF", pattern: "delete session *",   category: "admin", risk: 2 }

# Lệnh riêng AMF
{ service: "ne-command", ne_profile: "AMF", pattern: "show sctp",          category: "monitoring" }
{ service: "ne-command", ne_profile: "AMF", pattern: "show ue *",          category: "monitoring" }
{ service: "ne-command", ne_profile: "AMF", pattern: "show registration *", category: "monitoring" }
{ service: "ne-command", ne_profile: "AMF", pattern: "deregister ue *",    category: "admin", risk: 2 }

# Lệnh riêng UPF
{ service: "ne-command", ne_profile: "UPF", pattern: "show pfcp-session *", category: "monitoring" }
{ service: "ne-command", ne_profile: "UPF", pattern: "show gtp-tunnel *",   category: "monitoring" }
```

**Command Group** — gom lệnh CÙNG PROFILE:

```
command-group: "smf-subscriber-ops"
  ne_profile: "SMF"
  members:
    - get subscriber
    - show pdu-session *
    # KHÔNG có "get session" ← đây là cách kiểm soát chi tiết

command-group: "smf-session-ops"
  ne_profile: "SMF"
  members:
    - get session
    - delete session *

command-group: "amf-monitoring"
  ne_profile: "AMF"
  members:
    - show sctp
    - show ue *
    - show registration *

command-group: "common-monitoring"
  ne_profile: "*"
  members:
    - show version
    - show running-config
    - ping *
```

### 4.6. Permission Assignment — 3 cấp NE scope

Permission gán cho group với **ne_scope** chỉ định áp dụng ở đâu:

```
ne_scope có 3 mức:

  "*"          → tất cả NE (global)
  "profile:SMF" → tất cả NE có ne_profile = "SMF"
  "ne:SMF-01"   → chỉ NE cụ thể tên SMF-01
```

**Ví dụ thực tế:**

```
# Team SMF — chỉ được subscriber ops trên SMF, monitoring chung trên mọi NE
group "team-smf-l1":
  NE access: [SMF-01, SMF-02]
  permissions:
    allow command-group "common-monitoring"       ne_scope "*"           ← monitoring chung
    allow command-group "smf-subscriber-ops"      ne_scope "profile:SMF" ← get subscriber OK
    deny  command-group "smf-session-ops"         ne_scope "profile:SMF" ← get session KHÔNG

# Team SMF senior — được cả subscriber + session ops
group "team-smf-l2":
  NE access: [SMF-01, SMF-02, SMF-03]
  permissions:
    allow command-group "common-monitoring"       ne_scope "*"
    allow command-group "smf-subscriber-ops"      ne_scope "profile:SMF"
    allow command-group "smf-session-ops"         ne_scope "profile:SMF" ← get session OK

# Team AMF — chỉ monitoring trên AMF
group "team-amf-l1":
  NE access: [AMF-01, AMF-02]
  permissions:
    allow command-group "common-monitoring"       ne_scope "*"
    allow command-group "amf-monitoring"          ne_scope "profile:AMF"

# NOC admin — full access mọi NE
group "noc-admin":
  NE access: [ALL]
  permissions:
    allow category "*"                            ne_scope "*"
    deny  pattern  "delete session *"             ne_scope "ne:SMF-01"   ← cấm delete trên prod SMF

# Intern — chỉ show version trên 1 NE test
group "intern":
  NE access: [SMF-TEST-01]
  permissions:
    allow pattern "show version"                  ne_scope "ne:SMF-TEST-01"
    allow pattern "ping *"                        ne_scope "ne:SMF-TEST-01"
```

### 4.7. Evaluation Logic (3 bước)

```
Input: user "truongnv", service "ne-command", command "get session", ne_id 5 (SMF-01)

Step 1: NE Access Check
  ├── User groups: ["team-smf-l1"]
  ├── Group "team-smf-l1" NEs: [SMF-01(id=5), SMF-02(id=6)]
  ├── ne_id 5 (SMF-01) in list? → YES
  └── SMF-01.ne_profile = "SMF"

Step 2: Collect applicable rules (filter by ne_scope)
  Rules từ group "team-smf-l1":
  │
  ├── allow cmd-group "common-monitoring"    ne_scope "*"
  │   → expand: [show version, show running-config, ping *]
  │   → "get session" match? NO
  │
  ├── allow cmd-group "smf-subscriber-ops"   ne_scope "profile:SMF"
  │   → NE SMF-01 profile = "SMF" → scope match? YES
  │   → expand: [get subscriber, show pdu-session *]
  │   → "get session" match? NO
  │
  └── deny cmd-group "smf-session-ops"       ne_scope "profile:SMF"
      → NE SMF-01 profile = "SMF" → scope match? YES
      → expand: [get session, delete session *]
      → "get session" match? YES → DENY

Step 3: Evaluate (AWS IAM logic)
  - Explicit Deny found: "get session" denied by cmd-group "smf-session-ops"
  → DENY

Result: DENY "command 'get session' denied by rule:
  deny command-group 'smf-session-ops' (ne_scope: profile:SMF) in group 'team-smf-l1'"
```

**Ngược lại, user "senior" thuộc group "team-smf-l2":**

```
Input: user "senior", command "get session", ne_id 5 (SMF-01)

Step 2: Rules từ group "team-smf-l2":
  ├── allow cmd-group "smf-session-ops"      ne_scope "profile:SMF"
  │   → expand: [get session, delete session *]
  │   → "get session" match? YES → ALLOW
  │
  └── (không có deny rule nào match)

Step 3: Evaluate
  - No explicit deny
  - Explicit allow found
  → ALLOW
```

### 4.8. NE Scope matching — thứ tự ưu tiên

Khi nhiều rule match ở nhiều scope khác nhau, scope cụ thể hơn thắng:

```
Priority (cao → thấp):
  1. ne:<tên_ne>        → cụ thể nhất (VD: ne:SMF-01)
  2. profile:<profile>  → theo loại NE (VD: profile:SMF)
  3. *                  → global

Ví dụ:
  allow category "monitoring"      ne_scope "*"           ← global allow
  deny  pattern  "get session"     ne_scope "ne:SMF-01"   ← specific deny

  → User chạy "get session" trên SMF-01: DENY (specific deny wins)
  → User chạy "get session" trên SMF-02: cần check tiếp rule khác
```

**Quy tắc đầy đủ (kết hợp scope priority + effect priority):**

```
1. Tại cùng scope level: Deny > Allow (AWS IAM)
2. Scope cụ thể hơn override scope chung hơn (Vault path specificity)
3. Nếu không có rule nào match → Implicit Deny (default deny)

Bảng quyết định:

  ne:X deny   + profile:Y allow  → DENY  (cùng match, scope ne:X cụ thể hơn)
  ne:X allow  + profile:Y deny   → ALLOW (scope ne:X cụ thể hơn, override deny ở scope rộng)
  profile:Y deny + * allow       → DENY  (scope profile cụ thể hơn *)
  * deny + * allow               → DENY  (cùng scope, deny wins)
```

### 4.9. Command Pattern Matching (prefix-based, không regex)

Tại sao không dùng regex?
- Admin NOC không quen regex
- Dễ sai, khó debug
- Prefix/glob đủ cho command line

**Rules**:
```
Pattern              Input                        Match?
────────────────────────────────────────────────────────────
get subscriber       get subscriber               YES (exact)
get subscriber       get subscriber 123           YES (prefix — command bắt đầu bằng pattern)
get session          get session                  YES (exact)
get *                get subscriber               YES (wildcard)
get *                get session                  YES (wildcard)
show *               show running-config          YES
show *               get subscriber               NO  (prefix "show" ≠ "get")
show interface *     show interface Gi0/0         YES
show interface *     show running-config          NO
*                    anything                     YES (catch-all)
```

**Match priority** (khi nhiều pattern match):
1. Exact match (không có `*`)
2. Longer prefix match (cụ thể hơn)
3. Shorter prefix / catch-all `*`

### 4.11. Quản trị mgt-svc (thay thế account_type)

Thay vì hardcode SuperAdmin/Admin/Normal, dùng **mgt permissions** (như Vault capabilities):

```
resource: "user"     actions: create, read, update, delete
resource: "ne"       actions: create, read, update, delete
resource: "group"    actions: create, read, update, delete, assign
resource: "command"  actions: create, read, update, delete
resource: "policy"   actions: create, read, update, delete
resource: "history"  actions: read
```

System groups:
```
superadmin  → *.* (tất cả)
mgt-admin   → user.*, group.*, ne.*, command.*, policy.*
mgt-viewer  → user.read, group.read, ne.read, command.read, history.read
```

### 4.8. Password Policy (theo Huawei + FreeIPA)

Áp dụng **per-group** (như FreeIPA), user kế thừa policy strict nhất từ các group.

```
Policy "strict" (cho admin):
  max_age_days: 60
  min_length: 12
  require_uppercase: true
  require_lowercase: true
  require_digit: true
  require_special: true
  history_count: 12
  max_login_failure: 3
  lockout_minutes: 30

Policy "standard" (cho operator):
  max_age_days: 90
  min_length: 8
  require_uppercase: true
  require_lowercase: true
  require_digit: true
  require_special: false
  history_count: 6
  max_login_failure: 5
  lockout_minutes: 15

Policy "relaxed" (cho viewer):
  max_age_days: 180
  min_length: 8
  history_count: 3
  max_login_failure: 10
  lockout_minutes: 5
```

**Effective policy** khi user thuộc nhiều group: lấy giá trị **strict nhất** cho mỗi field:
- `max_age_days` → min(60, 90) = 60
- `min_length` → max(12, 8) = 12
- `max_login_failure` → min(3, 5) = 3

---

## 5. Database Schema

### 5.1. ERD tổng quan

```
                            ┌─────────────────────┐
                            │    tbl_account       │
                            │    (existing)        │
                            │                      │
                            │ + password_expires_at│  ← NEW
                            └──────────┬───────────┘
                                       │ M:N (existing)
                            ┌──────────┴───────────┐
                            │ cli_user_group_mapping│
                            │     (existing)        │
                            └──────────┬───────────┘
                                       │
                            ┌──────────┴───────────┐
                            │     cli_group         │
                            │     (existing)        │
                            │                       │
                            │ + is_system           │  ← NEW
                            │ + password_policy_id  │  ← NEW
                            └──┬────────┬───────┬───┘
                               │        │       │
               ┌───────────────┘        │       └──────────────────┐
               │                        │                          │
    ┌──────────┴──────────┐  ┌──────────┴──────────┐  ┌───────────┴──────────┐
    │cli_group_ne_mapping │  │cli_group_cmd_perm   │  │cli_group_mgt_perm    │
    │    (existing)       │  │       (NEW)         │  │       (NEW)          │
    └─────────────────────┘  │                     │  └──────────────────────┘
                             │ + ne_scope ─────────┼──── "*" | "profile:X" | "ne:Y"
                             └──────────┬──────────┘
                                        │ references
                             ┌──────────┴──────────┐
                             │  cli_command_def     │
                             │       (NEW)          │
                             │                      │
                             │ + ne_profile ────────┼──── lệnh thuộc loại NE nào
                             └──────────┬──────────┘
                                        │ M:N
                             ┌──────────┴──────────┐
                             │cli_command_group_    │
                             │    mapping (NEW)     │
                             └──────────┬──────────┘
                                        │
                             ┌──────────┴──────────┐
                             │ cli_command_group    │
                             │       (NEW)          │
                             │ + ne_profile         │
                             └─────────────────────┘

    ┌──────────────────────┐
    │  cli_ne_profile      │  ← NEW: phân loại NE theo tập lệnh
    │                      │
    │  "SMF", "AMF", "UPF" │
    └──────────────────────┘

    ┌──────────────────────┐  ┌─────────────────────┐
    │ cli_password_policy  │  │ cli_password_history │
    │       (NEW)          │  │       (NEW)          │
    └──────────────────────┘  └─────────────────────┘
```

### 5.2. Bảng mới chi tiết

#### `cli_ne_profile` — Phân loại NE theo function/tập lệnh

```sql
CREATE TABLE cli_ne_profile (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    name          VARCHAR(64)  NOT NULL UNIQUE,   -- 'SMF', 'AMF', 'UPF', 'generic-router'
    description   VARCHAR(512),
    created_at    TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);
```

> NE bảng `cli_ne` sẽ thêm cột `ne_profile_id` (FK → `cli_ne_profile`).
> Profile xác định loại NE → quyết định tập lệnh khả dụng.

#### `cli_command_def` — Registry lệnh, gắn với NE profile

```sql
CREATE TABLE cli_command_def (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    service       VARCHAR(32)  NOT NULL,  -- 'ne-command' | 'ne-config'
    ne_profile    VARCHAR(64)  NOT NULL DEFAULT '*',  -- 'SMF' | 'AMF' | '*' (chung cho mọi NE)
    pattern       VARCHAR(256) NOT NULL,  -- 'get subscriber' | 'show *'
    category      VARCHAR(32)  NOT NULL,  -- 'monitoring' | 'configuration' | 'admin' | 'debug'
    risk_level    INT          NOT NULL DEFAULT 0,  -- 0=safe, 1=normal, 2=dangerous
    description   VARCHAR(512),
    created_by    VARCHAR(64),
    created_at    TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uq_service_profile_pattern (service, ne_profile, pattern)
);
```

#### `cli_command_group` — Nhóm lệnh, gắn với NE profile

```sql
CREATE TABLE cli_command_group (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    name          VARCHAR(64)  NOT NULL UNIQUE,
    ne_profile    VARCHAR(64)  NOT NULL DEFAULT '*',  -- 'SMF' | 'AMF' | '*'
    service       VARCHAR(32)  NOT NULL,  -- 'ne-command' | 'ne-config' | '*'
    description   VARCHAR(512),
    created_by    VARCHAR(64),
    created_at    TIMESTAMP    DEFAULT CURRENT_TIMESTAMP
);
```

#### `cli_command_group_mapping` — Lệnh thuộc nhóm nào

```sql
CREATE TABLE cli_command_group_mapping (
    command_group_id  BIGINT NOT NULL,
    command_def_id    BIGINT NOT NULL,

    PRIMARY KEY (command_group_id, command_def_id),
    FOREIGN KEY (command_group_id) REFERENCES cli_command_group(id) ON DELETE CASCADE,
    FOREIGN KEY (command_def_id)   REFERENCES cli_command_def(id)   ON DELETE CASCADE
);
```

#### `cli_group_cmd_permission` — Group được phép/cấm gì, trên NE nào

```sql
CREATE TABLE cli_group_cmd_permission (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    group_id      BIGINT       NOT NULL,
    service       VARCHAR(32)  NOT NULL,     -- 'ne-command' | 'ne-config' | '*'
    ne_scope      VARCHAR(128) NOT NULL DEFAULT '*',
                  -- '*'           = tất cả NE
                  -- 'profile:SMF' = tất cả NE có profile SMF
                  -- 'ne:SMF-01'   = chỉ NE tên SMF-01
    grant_type    VARCHAR(16)  NOT NULL,     -- 'command_group' | 'category' | 'pattern'
    grant_value   VARCHAR(256) NOT NULL,     -- tên command_group | tên category | pattern string
    effect        VARCHAR(8)   NOT NULL,     -- 'allow' | 'deny'

    FOREIGN KEY (group_id) REFERENCES cli_group(id) ON DELETE CASCADE,
    UNIQUE KEY uq_perm (group_id, service, ne_scope, grant_type, grant_value)
);
```

#### `cli_group_mgt_permission` — Quyền quản trị mgt-svc

```sql
CREATE TABLE cli_group_mgt_permission (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    group_id      BIGINT       NOT NULL,
    resource      VARCHAR(32)  NOT NULL,     -- 'user' | 'ne' | 'group' | 'command' | 'policy' | 'history' | '*'
    action        VARCHAR(16)  NOT NULL,     -- 'create' | 'read' | 'update' | 'delete' | '*'

    FOREIGN KEY (group_id) REFERENCES cli_group(id) ON DELETE CASCADE,
    UNIQUE KEY uq_mgt_perm (group_id, resource, action)
);
```

#### `cli_password_policy`

```sql
CREATE TABLE cli_password_policy (
    id                BIGINT PRIMARY KEY AUTO_INCREMENT,
    name              VARCHAR(64)  NOT NULL UNIQUE,
    max_age_days      INT          NOT NULL DEFAULT 0,     -- 0 = never expires
    min_length        INT          NOT NULL DEFAULT 8,
    require_uppercase BOOLEAN      NOT NULL DEFAULT FALSE,
    require_lowercase BOOLEAN      NOT NULL DEFAULT FALSE,
    require_digit     BOOLEAN      NOT NULL DEFAULT FALSE,
    require_special   BOOLEAN      NOT NULL DEFAULT FALSE,
    history_count     INT          NOT NULL DEFAULT 0,     -- 0 = no history check
    max_login_failure INT          NOT NULL DEFAULT 0,     -- 0 = no lockout
    lockout_minutes   INT          NOT NULL DEFAULT 0
);
```

#### `cli_password_history`

```sql
CREATE TABLE cli_password_history (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id       BIGINT       NOT NULL,
    password_hash VARCHAR(256) NOT NULL,
    changed_at    TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES tbl_account(account_id) ON DELETE CASCADE,
    INDEX idx_user_changed (user_id, changed_at DESC)
);
```

### 5.3. Thay đổi trên bảng hiện tại

#### `tbl_account` — thêm cột

```sql
ALTER TABLE tbl_account
    ADD COLUMN password_expires_at TIMESTAMP NULL;
```

#### `cli_ne` — thêm cột NE profile

```sql
ALTER TABLE cli_ne
    ADD COLUMN ne_profile_id BIGINT NULL,
    ADD FOREIGN KEY (ne_profile_id) REFERENCES cli_ne_profile(id);
```

> `ne_profile_id` xác định loại NE (SMF, AMF, UPF...).
> NE chưa gán profile → chỉ match command có `ne_profile = "*"`.
> Khác với `system_type` (VD: "5GC" — quá rộng, cùng 5GC mà AMF/SMF/UPF lệnh khác nhau).

#### `cli_group` — thêm cột

```sql
ALTER TABLE cli_group
    ADD COLUMN is_system         BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN password_policy_id BIGINT  NULL,
    ADD FOREIGN KEY (password_policy_id) REFERENCES cli_password_policy(id);
```

---

## 6. Flow chi tiết

### 6.1. Login Flow

```
User gõ username/password
         │
         ▼
cli-ne-command/config
         │
         │  POST /aa/authenticate
         │  { username, password }
         │
         ▼
     mgt-svc
         │
         ├── 1. Tìm user trong DB
         │   └── Không tìm thấy → 401 "invalid credentials"
         │
         ├── 2. Check account locked?
         │   ├── login_failure_count >= policy.max_login_failure
         │   │   AND now < locked_time + policy.lockout_minutes
         │   └── YES → 403 "account locked, try after N minutes"
         │
         ├── 3. Verify password (bcrypt)
         │   ├── WRONG → login_failure_count++
         │   │   ├── if >= max → set locked_time = now
         │   │   └── 401 "invalid credentials (N attempts remaining)"
         │   └── OK → reset login_failure_count = 0
         │
         ├── 4. Check password expired?
         │   ├── password_expires_at != NULL AND now > password_expires_at
         │   └── YES → set force_change_pass = true
         │
         ├── 5. Collect groups
         │   └── SELECT group_id FROM cli_user_group_mapping WHERE user_id = ?
         │
         ├── 6. Issue JWT
         │   {
         │     sub: "truongnv",
         │     groups: ["noc-l2", "team-smf"],
         │     pwd_exp: false,
         │     exp: ...
         │   }
         │
         └── 7. Return token
              { token: "Basic eyJ...", password_expired: false }
```

### 6.2. Authorization Flow (khi cli-ne-* cần check quyền)

**Option A: Bulk query lúc đầu session**

```
cli-ne-command
    │
    │  GET /aa/authorize/effective
    │  Authorization: Basic <jwt>
    │
    ▼
mgt-svc
    │
    ├── 1. Parse JWT → username, groups
    │
    ├── 2. Check password_expired?
    │   └── YES → return { password_expired: true, ... }
    │       cli-ne-command bắt user đổi pass trước khi dùng
    │
    ├── 3. Collect NEs
    │   └── Union NE từ tất cả group (logic hiện tại)
    │
    ├── 4. Collect Command Permissions
    │   └── Từ mỗi group → cli_group_cmd_permission
    │       └── Resolve grant_type:
    │           - "command_group" → expand ra danh sách pattern trong nhóm đó
    │           - "category" → expand ra tất cả pattern có category đó
    │           - "pattern" → dùng trực tiếp
    │
    ├── 5. Build effective rules
    │   └── Tách allow vs deny, sắp xếp theo specificity
    │
    └── 6. Return response
         {
           username: "truongnv",
           password_expired: false,
           nes: [
             { ne_id: 5, ne_name: "SMF-01", ne_profile: "SMF", ... },
             { ne_id: 6, ne_name: "SMF-02", ne_profile: "SMF", ... }
           ],
           commands: {
             "ne-command": {
               // Rules grouped by NE scope for client-side evaluation
               rules: [
                 // Global rules (apply to all assigned NEs)
                 { ne_scope: "*",           effect: "allow", patterns: ["show version", "ping *"] },
                 // Profile-scoped rules (apply to NEs with matching profile)
                 { ne_scope: "profile:SMF", effect: "allow", patterns: ["get subscriber", "show pdu-session *"] },
                 { ne_scope: "profile:SMF", effect: "deny",  patterns: ["get session", "delete session *"] },
                 // NE-specific rules (apply to one NE only)
                 { ne_scope: "ne:SMF-01",   effect: "deny",  patterns: ["delete *"] }
               ]
             },
             "ne-config": {
               rules: [ ... ]
             }
           }
         }
```

cli-ne-command/config **cache kết quả này** trong session. Mỗi lệnh user gõ, check local trước.

**Option B: Per-command check (realtime)**

```
cli-ne-command
    │
    │  POST /aa/authorize/check-command
    │  {
    │    service: "ne-command",
    │    command: "show running-config",
    │    ne_id: 1
    │  }
    │
    ▼
mgt-svc
    │
    ├── 1. Check NE access
    │   └── ne_id 5 (SMF-01) thuộc group nào? User có trong group đó?
    │
    ├── 2. Resolve NE info
    │   └── ne_id 5 → ne_name: "SMF-01", ne_profile: "SMF"
    │
    ├── 3. Collect matching rules (filter by ne_scope)
    │   Applicable scopes for SMF-01: ["ne:SMF-01", "profile:SMF", "*"]
    │   
    │   Pattern match "get subscriber" against all applicable rules:
    │   - allow cmd-group "common-monitoring" (scope: *)
    │     → [show version, ping *] → "get subscriber" match? NO
    │   - allow cmd-group "smf-subscriber-ops" (scope: profile:SMF)
    │     → [get subscriber, show pdu-session *] → "get subscriber" match? YES → ALLOW
    │   - deny cmd-group "smf-session-ops" (scope: profile:SMF)
    │     → [get session, delete session *] → "get subscriber" match? NO
    │
    ├── 4. Evaluate (scope priority + AWS IAM logic)
    │   - Most specific scope with a match: "profile:SMF" → allow
    │   - No deny at same or more specific scope
    │   → ALLOW
    │
    └── 5. Return
         { allowed: true, matched_rule: "allow:get subscriber (smf-subscriber-ops, scope:profile:SMF)" }
```

**Recommendation**: Dùng **cả hai**:
- Option A: lúc login session, lấy full danh sách → cache trên client
- Option B: cho lệnh risk_level >= 2 hoặc khi cần double-check

### 6.3. Password Change Flow

```
User request change password
    │
    ├── 1. Determine effective policy
    │   └── Lấy policy từ tất cả group, pick strict nhất per field
    │       VD: user thuộc group A (policy "strict") + group B (policy "standard")
    │       → effective: max_age=60, min_length=12, require_all=true, history=12
    │
    ├── 2. Validate new password
    │   ├── Check min_length ≥ 12
    │   ├── Check uppercase, lowercase, digit, special
    │   ├── Check not same as current
    │   └── Check not in last 12 passwords (query cli_password_history)
    │       └── Nếu FAIL → return error chi tiết
    │
    ├── 3. Save old password to history
    │   └── INSERT cli_password_history (user_id, password_hash, now)
    │
    ├── 4. Update password
    │   └── tbl_account.password = bcrypt(new_password)
    │       tbl_account.password_expires_at = now + effective.max_age_days
    │       tbl_account.force_change_pass = false
    │       tbl_account.last_change_pass = now
    │
    └── 5. Trim password history
        └── Giữ lại max(history_count) bản ghi, xoá cũ hơn
```

### 6.4. Quản lý Command qua CLI/API

#### Khai báo lệnh mới (với NE profile)

```
CLI:
  # Lệnh chung cho mọi NE
  set command service ne-command ne_profile "*" pattern "show version" category monitoring

  # Lệnh riêng SMF
  set command service ne-command ne_profile SMF pattern "get subscriber" category monitoring
  set command service ne-command ne_profile SMF pattern "get session" category monitoring
  set command service ne-command ne_profile SMF pattern "delete session *" category admin risk_level 2

  # Lệnh riêng AMF
  set command service ne-command ne_profile AMF pattern "show sctp" category monitoring
  set command service ne-command ne_profile AMF pattern "show ue *" category monitoring

API:
  POST /aa/command-def/create
  {
    "service": "ne-command",
    "ne_profile": "SMF",
    "pattern": "get subscriber",
    "category": "monitoring",
    "risk_level": 0,
    "description": "Query subscriber information on SMF"
  }
```

#### Tạo nhóm lệnh (gắn profile)

```
CLI:
  set command-group name "smf-subscriber-ops" ne_profile SMF service ne-command
  map command-group "smf-subscriber-ops" command "get subscriber"
  map command-group "smf-subscriber-ops" command "show pdu-session *"
  # KHÔNG thêm "get session" ← kiểm soát ở đây

  set command-group name "smf-session-ops" ne_profile SMF service ne-command
  map command-group "smf-session-ops" command "get session"
  map command-group "smf-session-ops" command "delete session *"

API:
  POST /aa/command-group/create
  { "name": "smf-subscriber-ops", "ne_profile": "SMF", "service": "ne-command" }

  POST /aa/command-group/1/commands
  { "command_def_ids": [10, 11] }
```

#### Gán quyền cho group (với NE scope)

```
CLI:
  # team-smf-l1: subscriber ops trên SMF, monitoring chung
  map group "team-smf-l1" allow command-group "common-monitoring"      ne_scope "*"
  map group "team-smf-l1" allow command-group "smf-subscriber-ops"     ne_scope "profile:SMF"
  map group "team-smf-l1" deny  command-group "smf-session-ops"        ne_scope "profile:SMF"

  # team-smf-l2: cả subscriber + session
  map group "team-smf-l2" allow command-group "common-monitoring"      ne_scope "*"
  map group "team-smf-l2" allow command-group "smf-subscriber-ops"     ne_scope "profile:SMF"
  map group "team-smf-l2" allow command-group "smf-session-ops"        ne_scope "profile:SMF"

  # cấm delete trên NE production cụ thể
  map group "team-smf-l2" deny  pattern "delete *"                     ne_scope "ne:SMF-01"

API:
  POST /aa/group/5/cmd-permissions
  {
    "service": "ne-command",
    "ne_scope": "profile:SMF",
    "grant_type": "command_group",
    "grant_value": "smf-subscriber-ops",
    "effect": "allow"
  }
```

#### Import bulk (với NE profile)

```
API:
  POST /aa/command-def/import
  Content-Type: application/json

  [
    { "service": "ne-command", "ne_profile": "*",   "pattern": "show version",       "category": "monitoring" },
    { "service": "ne-command", "ne_profile": "*",   "pattern": "ping *",             "category": "monitoring" },
    { "service": "ne-command", "ne_profile": "*",   "pattern": "reload",             "category": "admin", "risk_level": 2 },
    { "service": "ne-command", "ne_profile": "SMF", "pattern": "get subscriber",     "category": "monitoring" },
    { "service": "ne-command", "ne_profile": "SMF", "pattern": "get session",        "category": "monitoring" },
    { "service": "ne-command", "ne_profile": "SMF", "pattern": "delete session *",   "category": "admin", "risk_level": 2 },
    { "service": "ne-command", "ne_profile": "AMF", "pattern": "show sctp",          "category": "monitoring" },
    { "service": "ne-command", "ne_profile": "AMF", "pattern": "show ue *",          "category": "monitoring" },
    { "service": "ne-config",  "ne_profile": "*",   "pattern": "get-config *",       "category": "monitoring" },
    { "service": "ne-config",  "ne_profile": "*",   "pattern": "edit-config *",      "category": "configuration" }
  ]
```

---

## 7. API Contract

### 7.1. Authentication (giữ nguyên + mở rộng response)

```
POST /aa/authenticate
Request:  { "username": "truongnv", "password": "..." }
Response: {
  "response_data": "Basic eyJ...",
  "password_expired": false,
  "groups": ["noc-l2", "team-smf"]
}
```

### 7.2. Authorization (MỚI)

```
GET  /aa/authorize/effective
  → Full NE list (with profile) + NE-scoped command permissions
  → Response bao gồm ne_profile cho mỗi NE, rules grouped by ne_scope

POST /aa/authorize/check-command
  → Check 1 lệnh cụ thể trên 1 NE cụ thể
  Request:  { "service": "ne-command", "command": "get session", "ne_id": 5 }
  Response (denied):
  {
    "allowed": false,
    "ne_name": "SMF-01",
    "ne_profile": "SMF",
    "reason": "denied by command-group 'smf-session-ops' (scope: profile:SMF) in group 'team-smf-l1'"
  }

  Response (allowed):
  {
    "allowed": true,
    "ne_name": "SMF-01",
    "ne_profile": "SMF",
    "matched_rule": "allow:get subscriber (smf-subscriber-ops, scope:profile:SMF)",
    "risk_level": 0
  }
```

### 7.3. NE Profile Management (MỚI)

```
GET    /aa/ne-profile/list                    ← liệt kê profiles (SMF, AMF, UPF...)
POST   /aa/ne-profile/create                  ← tạo profile mới
POST   /aa/ne-profile/update                  ← sửa profile
DELETE /aa/ne-profile/{id}                    ← xoá profile
POST   /aa/ne/{ne_id}/profile                 ← gán profile cho NE
```

### 7.4. Command Definition Management

```
GET    /aa/command-def/list[?service=ne-command&ne_profile=SMF&category=monitoring]
POST   /aa/command-def/create                 ← bắt buộc có ne_profile
POST   /aa/command-def/update
DELETE /aa/command-def/{id}
POST   /aa/command-def/import                 ← bulk import với ne_profile
```

### 7.5. Command Group Management

```
GET    /aa/command-group/list[?service=ne-command&ne_profile=SMF]
POST   /aa/command-group/create               ← bắt buộc có ne_profile
POST   /aa/command-group/update
DELETE /aa/command-group/{id}
GET    /aa/command-group/{id}/commands        ← liệt kê lệnh trong nhóm
POST   /aa/command-group/{id}/commands        ← thêm lệnh vào nhóm
DELETE /aa/command-group/{id}/commands/{cmd_id}  ← gỡ lệnh khỏi nhóm
```

### 7.6. Group Command Permission

```
GET    /aa/group/{id}/cmd-permissions         ← xem quyền lệnh (bao gồm ne_scope)
POST   /aa/group/{id}/cmd-permissions         ← gán quyền (bắt buộc có ne_scope)
DELETE /aa/group/{id}/cmd-permissions/{perm_id}  ← gỡ quyền
```

### 7.7. Group Mgt Permission

```
GET    /aa/group/{id}/mgt-permissions         ← xem quyền quản trị
POST   /aa/group/{id}/mgt-permissions         ← gán quyền
DELETE /aa/group/{id}/mgt-permissions/{perm_id}  ← gỡ quyền
```

### 7.8. Password Policy

```
GET    /aa/password-policy/list
POST   /aa/password-policy/create
POST   /aa/password-policy/update
DELETE /aa/password-policy/{id}
POST   /aa/group/{id}/password-policy         ← gán policy cho group
```

---

## 8. Migration Plan

### Phase 1: Database Migration (không phá gì)

1. Tạo bảng mới: `cli_ne_profile`, `cli_command_def`, `cli_command_group`,
   `cli_command_group_mapping`, `cli_group_cmd_permission`, `cli_group_mgt_permission`,
   `cli_password_policy`, `cli_password_history`
2. Thêm cột mới vào `tbl_account` (`password_expires_at`), `cli_ne` (`ne_profile_id`),
   `cli_group` (`is_system`, `password_policy_id`)
3. Không xoá hay đổi gì cũ

### Phase 2: Seed Data

1. Tạo NE profiles: "SMF", "AMF", "UPF" (+ các loại NE khác nếu có)
2. Gán profile cho NE hiện tại dựa trên ne_name prefix (SMF-* → "SMF", AMF-* → "AMF"...)
3. Tạo password policies: "strict", "standard", "relaxed"
4. Tạo system groups: "superadmin", "mgt-admin", "noc-admin", "noc-l2", "noc-l1"
5. Gán mgt permissions cho system groups
6. Seed command definitions cho mỗi NE profile

### Phase 3: User Migration

```sql
-- SuperAdmin → superadmin group
INSERT INTO cli_user_group_mapping (user_id, group_id)
SELECT account_id, (SELECT id FROM cli_group WHERE name = 'superadmin')
FROM tbl_account WHERE account_type = 0;

-- Admin → noc-admin + mgt-admin groups
INSERT INTO cli_user_group_mapping (user_id, group_id)
SELECT account_id, (SELECT id FROM cli_group WHERE name = 'noc-admin')
FROM tbl_account WHERE account_type = 1;

INSERT INTO cli_user_group_mapping (user_id, group_id)
SELECT account_id, (SELECT id FROM cli_group WHERE name = 'mgt-admin')
FROM tbl_account WHERE account_type = 1;

-- Normal → noc-l1 group
INSERT INTO cli_user_group_mapping (user_id, group_id)
SELECT account_id, (SELECT id FROM cli_group WHERE name = 'noc-l1')
FROM tbl_account WHERE account_type = 2;
```

### Phase 4: Cập nhật Code

1. Service layer: thêm logic resolve permissions từ groups
2. Middleware: thay `CheckRole` bằng `CheckMgtPermission(resource, action)`
3. Handler: thêm endpoints mới cho command/group/policy management
4. JWT: thêm `groups` claim, giữ `aud` để backward-compat
5. Authorize endpoints: implement `/aa/authorize/effective` và `/aa/authorize/check-command`

### Phase 5: Deprecate account_type

1. Giữ field `account_type` 1 version, không dùng cho logic nào nữa
2. API cũ trả `aud: "admin"/"user"` dựa trên mgt permissions thay vì account_type
3. Version sau: xoá field

---

## Tham khảo thêm

| Hệ thống | Tài liệu chính |
|-----------|-----------------|
| Cisco RBAC | IOS Security Configuration Guide — CLI Views |
| Juniper Login Classes | JUNOS System Basics — Configuring Login Classes |
| Huawei VRP | Configuration Guide — User Management |
| Nokia SR OS | System Management Guide — User Profiles |
| AWS IAM | IAM Policy Evaluation Logic |
| HashiCorp Vault | Vault Policies Documentation |
| Kubernetes RBAC | K8s Authorization Documentation |
| FreeIPA | FreeIPA HBAC + Sudo Rules |
