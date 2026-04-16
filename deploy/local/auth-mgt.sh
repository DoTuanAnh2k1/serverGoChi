#!/bin/bash
# PAM auth script — xác thực user qua mgt-service API.
# Input: PAM_USER (env), password (stdin từ pam_exec expose_authtok).
# Output: exit 0 = auth OK, exit 1 = auth fail.
# Local dev version: tự tạo system user nếu chưa tồn tại.

# PAM exec strips env vars, so read from config file first
if [ -f /etc/mgt-service-url ]; then
    MGT_URL=$(cat /etc/mgt-service-url)
else
    MGT_URL="${MGT_SERVICE_URL:-http://cli-mgt-svc:3000}"
fi
read -r PASS

RESP=$(curl -sf --max-time 5 -X POST "$MGT_URL/aa/authenticate" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$PAM_USER\",\"password\":\"$PASS\"}" 2>/dev/null)

[ $? -ne 0 ] && exit 1

STATUS=$(printf '%s' "$RESP" | jq -r '.status // empty' 2>/dev/null)
TOKEN=$(printf '%s'  "$RESP" | jq -r '.response_data // empty' 2>/dev/null)

if [ "$STATUS" = "success" ] && [ -n "$TOKEN" ]; then
    # Auto-create system user if not exists (needed for SSH session)
    if ! id "$PAM_USER" &>/dev/null; then
        useradd -m -s /bin/bash "$PAM_USER" 2>/dev/null
    fi
    mkdir -p /tmp/cli-tokens
    printf '%s' "$TOKEN" > "/tmp/cli-tokens/$PAM_USER"
    exit 0
fi
exit 1
