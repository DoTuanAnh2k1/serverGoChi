# C client for cli-mgt-svc (v2)

Tiny libcurl-backed client in plain C11 — designed to live inside the
cli-gate / ne-command proxy (which is C). No JSON dependency; the one-shot
parser in [mgt_client.c](mgt_client.c) handles the flat-object shape the v2
API returns.

## Files

| File | What |
|---|---|
| [mgt_client.h](mgt_client.h) | Public API — `mgt_client_new`, `mgt_authenticate`, `mgt_request`, `mgt_authorize_check`, `mgt_save_history` |
| [mgt_client.c](mgt_client.c) | libcurl implementation + minimal JSON string/bool lookup |
| [demo.c](demo.c) | End-to-end example: login, list users/NEs/commands, probe authorize, push audit |
| [Makefile](Makefile) | `make` / `make run` |

## Build

```bash
cd examples/c
make               # produces ./demo
# or:
gcc -Wall -O2 -o demo mgt_client.c demo.c -lcurl
```

Requires libcurl-dev (Debian: `apt install libcurl4-openssl-dev`, Alpine:
`apk add curl-dev`, macOS: already present).

## Run

```bash
MGT_BASE=http://localhost:3000 MGT_USER=admin MGT_PASS=admin ./demo
```

## Minimal usage

```c
#include "mgt_client.h"

mgt_client_t *c = mgt_client_new("http://localhost:3000");
mgt_authenticate(c, "admin", "admin");       /* token stored on c */

/* The one question that matters — answered by the service */
authorize_decision_t d;
mgt_authorize_check(c, "alice", /*ne_id*/7, /*command_id*/19, &d);
if (d.allowed) { /* proceed */ }
else           { fprintf(stderr, "denied: %s\n", d.reason); }

/* Audit push (no JWT required by /aa/history/save) */
mgt_save_history(c, "alice", "show version", "htsmf01", "10.0.0.1",
                 "ne-command", "ok");

mgt_client_free(c);
```

## Raw API escape hatch

If you need an endpoint that isn't wrapped, use `mgt_request`:

```c
char *body = NULL;
long  status = 0;
mgt_request(c, "GET", "/aa/ne-access-groups/5/users", NULL, &body, &status);
/* parse body yourself with cJSON / jansson / json-c */
free(body);
```

## Notes

- The token returned by `/aa/authenticate` is `Basic <jwt>` — the client
  sends it back verbatim in `Authorization`. That's what the Go middleware
  expects (it strips the `Basic ` prefix internally).
- The JSON helpers (`json_find_string`, `json_find_bool`) walk the
  top-level object only. They're enough for simple responses
  (`{ "token": "...", "status": "ok" }`) and the authorize decision. For
  anything nested, link a real JSON library.
- No retries, no connection pooling. One `CURL*` per client is reused
  across calls, but there's no multithreading safety — serialize calls per
  client, or create one client per thread.
