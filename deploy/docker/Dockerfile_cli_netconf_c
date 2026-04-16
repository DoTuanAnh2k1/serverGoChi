# =============================================================================
# Dockerfile_cli_netconf_c — CLI NETCONF C rewrite (feature/c-maapi-rewrite)
#
# SSH Server Mode: OpenSSH server + xác thực qua mgt-service API.
# User SSH vào container → PAM xác thực → cli-netconf tự hỏi login →
# lấy danh sách NE từ mgt-service → user chọn NE → MAAPI connect.
#
# Build args:
#   CLI_NETCONF_REPO   — Git repo URL         (default: GitHub public)
#   CLI_NETCONF_BRANCH — Git branch           (default: feature/c-maapi-rewrite)
#
# Runtime env vars (đặt trong docker-compose):
#   MGT_SVC_BASE     — URL gốc mgt-service   (default: http://mgt-service:3000)
#   LOG_LEVEL        — Log level              (default: info)
#   SEED_USERNAME    — User để sync users     (default: anhdt195)
#   SEED_PASSWORD    — Password seed user     (default: 123)
#
# Truy cập:
#   ssh <username>@<host> -p 2222
# =============================================================================

# ── Build stage ──────────────────────────────────────────────────────────────
# Khi CLI_NETCONF_SRC được set (local path), build từ local source.
# Mặc định vẫn clone từ GitHub nếu dùng bên ngoài docker-compose.
FROM ubuntu:24.04 AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
        gcc make libxml2-dev libreadline-dev ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

# Copy toàn bộ source từ build context (local cli-netconf repo)
COPY src/ src/
COPY include/ include/
COPY Makefile .
COPY libconfd-server.so libconfd.so
COPY libcrypto-server.so libcrypto.so

# Build với rpath trỏ thẳng tới /usr/lib/cli trong runtime container
RUN make CONFD_LIB=/build/libconfd.so \
         LDFLAGS="-lreadline -lxml2 -L/build -lconfd -lcrypto -Wl,-rpath,/usr/lib/cli"

# ── Runtime stage ────────────────────────────────────────────────────────────
FROM ubuntu:24.04

RUN apt-get update && apt-get install -y --no-install-recommends \
        openssh-server \
        libxml2 \
        libreadline8 \
        curl \
        jq \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /var/run/sshd /usr/lib/cli

# Binary + shared libraries
COPY --from=builder /build/cli-netconf          /usr/local/bin/cli-netconf
COPY --from=builder /build/libconfd.so          /usr/lib/cli/libconfd.so
COPY --from=builder /build/libcrypto.so         /usr/lib/cli/libcrypto.so.1.0.0

RUN ln -sf /usr/lib/cli/libcrypto.so.1.0.0 /usr/lib/cli/libcrypto.so.10 \
    && echo "/usr/lib/cli" > /etc/ld.so.conf.d/cli.conf \
    && ldconfig

# ── PAM auth script — xác thực user qua mgt-service API ─────────────────────
COPY <<'AUTH_SCRIPT' /usr/local/bin/auth-mgt.sh
#!/bin/bash
# Gọi POST /aa/authenticate lên mgt-service.
# Input: PAM_USER (env), password (stdin từ pam_exec expose_authtok).
# Output: exit 0 = auth OK, exit 1 = auth fail.
# Đọc MGT URL từ /etc/mgt-url vì sshd child không kế thừa env container.

MGT_URL=$(cat /etc/mgt-url 2>/dev/null)
MGT_URL="${MGT_URL:-http://mgt-service:3000}"
read -r PASS

RESP=$(curl -sf --max-time 5 -X POST "$MGT_URL/aa/authenticate" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$PAM_USER\",\"password\":\"$PASS\"}" 2>/dev/null)

[ $? -ne 0 ] && exit 1

STATUS=$(printf '%s' "$RESP" | jq -r '.status // empty' 2>/dev/null)
TOKEN=$(printf '%s'  "$RESP" | jq -r '.response_data // empty' 2>/dev/null)

if [ "$STATUS" = "success" ] && [ -n "$TOKEN" ]; then
    mkdir -p /tmp/cli-tokens
    printf '%s' "$TOKEN" > "/tmp/cli-tokens/$PAM_USER"
    exit 0
fi
exit 1
AUTH_SCRIPT

RUN chmod +x /usr/local/bin/auth-mgt.sh

# ── CLI wrapper — SSH Server Mode (không set CONFD env → CLI tự login + chọn NE)
COPY <<'CLI_WRAPPER' /usr/local/bin/cli-wrapper.sh
#!/bin/bash
# Wrapper chạy cli-netconf ở SSH Server Mode.
# KHÔNG set CONFD_IPC_ADDR/PORT → CLI chạy SSH server mode.
# Truyền MGT_SVC_TOKEN + MGT_SVC_USER từ PAM → CLI skip login, vào thẳng chọn NE.

export LD_LIBRARY_PATH=/usr/lib/cli
export MGT_SVC_BASE=$(cat /etc/mgt-url 2>/dev/null || echo "http://mgt-service:3000")
export MAAPI_USER="${USER:-admin}"
export LOG_LEVEL=$(cat /etc/cli-log-level 2>/dev/null || echo "info")
export LOG_FILE="/var/log/cli-netconf/${USER:-unknown}.log"

# PAM auth đã lưu JWT token → truyền cho CLI để skip login prompt
TOKEN_FILE="/tmp/cli-tokens/$USER"
if [ -f "$TOKEN_FILE" ]; then
    export MGT_SVC_TOKEN="$(cat "$TOKEN_FILE")"
    export MGT_SVC_USER="$USER"
fi

mkdir -p /var/log/cli-netconf

exec /usr/local/bin/cli-netconf
CLI_WRAPPER

RUN chmod +x /usr/local/bin/cli-wrapper.sh

# ── Cấu hình SSH server ──────────────────────────────────────────────────────
RUN sed -i \
        -e 's/#PasswordAuthentication yes/PasswordAuthentication yes/' \
        -e 's/#PermitRootLogin.*/PermitRootLogin no/' \
        -e 's/#UsePAM.*/UsePAM yes/' \
        /etc/ssh/sshd_config \
    && echo "ForceCommand /usr/local/bin/cli-wrapper.sh" >> /etc/ssh/sshd_config \
    && echo "PrintMotd no"      >> /etc/ssh/sshd_config \
    && echo "PrintLastLog no"   >> /etc/ssh/sshd_config \
    && echo "AllowAgentForwarding no" >> /etc/ssh/sshd_config \
    && echo "AllowTcpForwarding no"   >> /etc/ssh/sshd_config

# ── PAM: xác thực qua script thay vì /etc/shadow ────────────────────────────
COPY <<'PAM_CFG' /etc/pam.d/sshd
auth    required  pam_exec.so  expose_authtok  /usr/local/bin/auth-mgt.sh
auth    optional  pam_permit.so
account required  pam_permit.so
session required  pam_limits.so
session required  pam_permit.so
PAM_CFG

# PAM cần user tồn tại trong system để tạo session
# Tạo một system user dùng chung; ForceCommand quyết định gì chạy
RUN useradd -m -s /bin/bash -G sudo cliuser \
    && echo "cliuser:!" | chpasswd

# ── Sync users script — fetch từ mgt-service, tạo system users cho SSH ───────
COPY <<'SYNC_SCRIPT' /usr/local/bin/sync-users.sh
#!/bin/bash
# sshd cần user tồn tại trong /etc/passwd trước khi PAM auth chạy.
# Script này lấy danh sách user từ mgt-service và tạo system users.

MGT_URL="${MGT_SVC_BASE:-http://mgt-service:3000}"
SEED_USER="${SEED_USERNAME:-anhdt195}"
SEED_PASS="${SEED_PASSWORD:-123}"

TOKEN=$(curl -sf --max-time 10 -X POST "$MGT_URL/aa/authenticate" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$SEED_USER\",\"password\":\"$SEED_PASS\"}" 2>/dev/null \
  | jq -r '.response_data // empty' 2>/dev/null)

if [ -z "$TOKEN" ]; then
  echo "sync-users: cannot login to mgt-service, creating seed user only"
  id "$SEED_USER" &>/dev/null || useradd -m -s /bin/bash "$SEED_USER" 2>/dev/null
  exit 0
fi

USERS=$(curl -sf --max-time 10 -H "Authorization: $TOKEN" \
  "$MGT_URL/aa/authenticate/user/show" 2>/dev/null)

if [ -n "$USERS" ]; then
  printf '%s' "$USERS" | jq -r '
    if type == "array" then .[].username
    else empty end
  ' 2>/dev/null | while read -r USERNAME; do
    if [ -n "$USERNAME" ] && ! id "$USERNAME" &>/dev/null; then
      useradd -m -s /bin/bash "$USERNAME" 2>/dev/null
      echo "  sync-users: created $USERNAME"
    fi
  done
  echo "sync-users: done"
else
  echo "sync-users: cannot fetch user list, creating seed user only"
  id "$SEED_USER" &>/dev/null || useradd -m -s /bin/bash "$SEED_USER" 2>/dev/null
fi
SYNC_SCRIPT

RUN chmod +x /usr/local/bin/sync-users.sh

# ── Entrypoint ───────────────────────────────────────────────────────────────
COPY <<'ENTRYPOINT' /entrypoint.sh
#!/bin/bash
ssh-keygen -A 2>/dev/null

# Ghi config ra file (sshd child processes không kế thừa env container)
printf '%s' "${MGT_SVC_BASE:-http://mgt-service:3000}" > /etc/mgt-url
printf '%s' "${LOG_LEVEL:-info}" > /etc/cli-log-level
mkdir -p /var/log/cli-netconf
chmod 1777 /var/log/cli-netconf

echo "=== CLI NETCONF SSH Server Mode ==="
echo "  MGT_SVC_BASE:  ${MGT_SVC_BASE:-http://mgt-service:3000}"
echo "  LOG_LEVEL:     ${LOG_LEVEL:-info}"
echo "  SSH port:      22"
echo ""

# Sync users từ mgt-service
/usr/local/bin/sync-users.sh

echo ""
echo "  SSH listening on port 22"
echo ""

exec /usr/sbin/sshd -D -e
ENTRYPOINT

RUN chmod +x /entrypoint.sh

EXPOSE 22

ENV MGT_SVC_BASE=http://mgt-service:3000 \
    LOG_LEVEL=info \
    SEED_USERNAME=anhdt195 \
    SEED_PASSWORD=123

ENTRYPOINT ["/entrypoint.sh"]
