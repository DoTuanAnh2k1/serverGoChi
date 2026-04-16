#!/bin/bash
# Custom entrypoint for cli-netconf local dev.
# Syncs users from mgt-service before starting sshd,
# so SSH doesn't reject "invalid user".

set -e

ssh-keygen -A 2>/dev/null

MGT_URL="${MGT_SERVICE_URL:-http://mgt-service:3000}"

# Write config to files so PAM auth and cli-wrapper can read them
# (pam_exec and ForceCommand strip environment variables)
printf '%s' "$MGT_URL" > /etc/mgt-service-url
# Resolve hostname to IP (C binary can't do DNS resolution)
CONFD_ADDR="${CONFD_IPC_ADDR:-127.0.0.1}"
RESOLVED_IP=$(getent hosts "$CONFD_ADDR" 2>/dev/null | awk '{print $1}')
if [ -n "$RESOLVED_IP" ]; then
    printf '%s' "$RESOLVED_IP" > /etc/confd-ipc-addr
    echo "  Resolved $CONFD_ADDR -> $RESOLVED_IP"
else
    printf '%s' "$CONFD_ADDR" > /etc/confd-ipc-addr
fi
printf '%s' "${CONFD_IPC_PORT:-4565}" > /etc/confd-ipc-port
printf '%s' "${NE_NAME:-confd}" > /etc/ne-name

echo "=== CLI NETCONF (C/MAAPI) Full Mode ==="
echo "  CONFD_IPC_ADDR:  ${CONFD_IPC_ADDR:-127.0.0.1}"
echo "  CONFD_IPC_PORT:  ${CONFD_IPC_PORT:-4565}"
echo "  MGT_SERVICE_URL: $MGT_URL"
echo "  NE_NAME:         ${NE_NAME:-confd}"
echo "  SSH port:        22"
echo ""

# Sync users from mgt-service: create system users for each account
echo "Syncing users from mgt-service..."

# Login to get token first (use seed user)
TOKEN=$(curl -sf --max-time 10 -X POST "$MGT_URL/aa/authenticate" \
  -H "Content-Type: application/json" \
  -d '{"username":"anhdt195","password":"123"}' 2>/dev/null \
  | jq -r '.response_data // empty' 2>/dev/null)

USERS_JSON=""
if [ -n "$TOKEN" ]; then
    USERS_JSON=$(curl -sf --max-time 10 "$MGT_URL/aa/authenticate/user/show" \
      -H "Authorization: $TOKEN" 2>/dev/null || echo "")
fi

if [ -n "$USERS_JSON" ]; then
    # Extract usernames from response (array of {username, ...})
    USERNAMES=$(printf '%s' "$USERS_JSON" | jq -r '
        if type == "array" then .[].username
        else empty end
    ' 2>/dev/null)

    for u in $USERNAMES; do
        if [ -n "$u" ] && ! id "$u" &>/dev/null; then
            useradd -m -s /bin/bash "$u" 2>/dev/null && echo "  created user: $u"
        fi
    done
    echo "User sync done."
else
    echo "Warning: could not fetch users from mgt-service, creating default user..."
    useradd -m -s /bin/bash anhdt195 2>/dev/null || true
fi

echo "Starting sshd..."
exec /usr/sbin/sshd -D -e
