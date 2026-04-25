# C client for cli-mgt-svc (v2)

Libcurl-backed client in plain C11 with **6 domain structs**, **41 dedicated
functions**, and **100% API coverage**. Designed to live inside the cli-gate /
ne-command proxy (which is C).

## Files

| File | What |
|---|---|
| [mgt_client.h](mgt_client.h) | Public API — 41 functions + 6 structs (`mgt_user_t`, `mgt_ne_t`, `mgt_command_t`, `mgt_group_t`, `mgt_access_entry_t`, `authorize_decision_t`) |
| [mgt_client.c](mgt_client.c) | libcurl implementation + JSON helpers |
| [demo.c](demo.c) | End-to-end: login, create user/NE/command, create groups, add members, authorize check, cleanup |
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

## What's covered

Every v2 endpoint under `/aa/*` (100% coverage, 41 functions):

- **Auth**: `mgt_authenticate` (stores token + role), `mgt_get_role`
- **Users**: `mgt_list_users`, `mgt_create_user` (with role), `mgt_update_user`, `mgt_delete_user`, `mgt_reset_password`
- **NEs**: `mgt_list_nes`, `mgt_create_ne`, `mgt_update_ne`, `mgt_delete_ne`
- **Commands**: `mgt_list_commands`, `mgt_create_command`, `mgt_update_command`, `mgt_delete_command`
- **NE access groups**: full CRUD + add/remove user + add/remove NE (8 functions)
- **Cmd exec groups**: full CRUD + add/remove user + add/remove command (8 functions)
- **Policy**: `mgt_get_password_policy`, `mgt_upsert_password_policy`
- **Access list**: `mgt_list_access_list`, `mgt_create_access_entry`, `mgt_delete_access_entry`
- **Authorize**: `mgt_authorize_check` with `authorize_decision_t` trace
- **History**: `mgt_save_history` (unauthenticated audit push)

## Minimal usage

```c
#include "mgt_client.h"

mgt_client_t *c = mgt_client_new("http://localhost:3000");
mgt_authenticate(c, "admin", "admin");
printf("role: %s\n", mgt_get_role(c));  /* "super_admin" */

/* Create a user with role */
char *json = NULL;
mgt_create_user(c, "operator1", "Pass1234!", "op1@vht.com", "Op 1", "admin", &json);
free(json);

/* Authorize check */
authorize_decision_t d;
mgt_authorize_check(c, "operator1", 7, 19, &d);
if (d.allowed) { /* proceed */ }
else           { fprintf(stderr, "denied: %s\n", d.reason); }

/* Audit push (no JWT required) */
mgt_save_history(c, "operator1", "show version", "htsmf01", "10.0.0.1",
                 "ne-command", "ok");

mgt_client_free(c);
```

## Raw API escape hatch

For endpoints without a dedicated wrapper, use `mgt_request`:

```c
char *body = NULL;
long  status = 0;
mgt_request(c, "GET", "/aa/ne-access-groups/5/users", NULL, &body, &status);
free(body);
```

## Notes

- Token format: `Basic <jwt>` — stored on client, sent in `Authorization` header.
- Role is stored after `mgt_authenticate` — get it with `mgt_get_role()`.
- JSON helpers (`json_find_string`, `json_find_bool`, `json_find_int`) handle
  top-level flat objects. For nested responses, link cJSON / jansson / json-c.
- One `CURL*` per client, reused across calls. Not thread-safe — serialize
  calls per client or create one per thread.
