#!/bin/bash
export LD_LIBRARY_PATH=/usr/lib/cli

# Read config from files (env vars stripped by PAM/sshd)
if [ -f /etc/confd-ipc-addr ]; then
    export CONFD_IPC_ADDR=$(cat /etc/confd-ipc-addr)
else
    export CONFD_IPC_ADDR="${CONFD_IPC_ADDR:-127.0.0.1}"
fi

if [ -f /etc/confd-ipc-port ]; then
    export CONFD_IPC_PORT=$(cat /etc/confd-ipc-port)
else
    export CONFD_IPC_PORT="${CONFD_IPC_PORT:-4565}"
fi

if [ -f /etc/ne-name ]; then
    export NE_NAME=$(cat /etc/ne-name)
else
    export NE_NAME="${NE_NAME:-confd}"
fi

export MAAPI_USER="${USER:-admin}"

# Read MGT service URL
if [ -f /etc/mgt-service-url ]; then
    export MGT_SERVICE_URL=$(cat /etc/mgt-service-url)
fi

# Load JWT token if available (saved by auth script)
TOKEN_FILE="/tmp/cli-tokens/$USER"
[ -f "$TOKEN_FILE" ] && export MGT_JWT_TOKEN="$(cat "$TOKEN_FILE")"

exec /usr/local/bin/cli-netconf
