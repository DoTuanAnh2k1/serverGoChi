# Authorization Design — v2

> v1's complex RBAC has been replaced with a 3-tier permission model +
> two-layer group authorization.

## The one question

Every authorization decision reduces to:

> **Can user `X` execute command `Y` on NE `Z`?**

The evaluator at [pkg/service/authorize.go](../pkg/service/authorize.go) answers it
with no other inputs.

## The rule

Allowed ⇔ all of:

1. `X` exists, `is_enabled = TRUE`, `locked_at IS NULL`
2. `Y` (by `command.id`) has `ne_id = Z.id` — i.e. the command is registered
   against that specific NE; there's no pattern match or per-profile
   inheritance
3. ∃ some `ne_access_group` where `X` is a member *and* `Z` is a member
4. ∃ some `cmd_exec_group` where `X` is a member *and* `Y` is a member

Any failure denies. The evaluator returns a small trace so the UI/CLI can
render which step tripped.

## Two independent groups — why

The split is deliberate: the *NE reachability* question ("does this operator
have any business touching that box?") is different from the *command
eligibility* question ("is this command approved for them to run?"), and a
fleet usually wires them up on different axes — ops teams get NE access by
site/region, command permissions by job function.

Collapsing both into a single group would force a combinatorial explosion
(`N × M`) of groups, or punch holes that are hard to audit.

## Management permission tiers

Independent of the authorize rule (which governs NE/command access), the
**management surface** (API write endpoints + admin frontend) is gated by
the user's `role` field:

| Role | Management surface | Restriction |
|------|-------------------|-------------|
| `super_admin` | Full access — CRUD everything | None |
| `admin` | CRUD users, NEs, commands, groups, policy, access-list | Cannot modify/delete/reset-pw `super_admin` accounts |
| `user` | Read-only (all GET endpoints) | All POST/PUT/DELETE return 403 |

The `role` claim is embedded in the JWT token and checked by the
`RequireAdmin` middleware on all write routes. Fine-grained super_admin
protection is enforced per-handler (e.g. `HandlerDeleteUser` checks target
user's role before allowing deletion).

## Auth-time gates

Independent of the authorize rule, **login** itself is gated by:

- **Account state** — disabled or locked accounts never authenticate.
- **Password policy** — singleton row at `password_policy.id = 1`. Covers
  length, complexity flags, max age, history count, lockout threshold.
- **Blacklist** — any matching entry denies.
- **Whitelist** — when any whitelist entry exists for a `match_type` the
  identity has a value for, the identity must match at least one.
  Whitelist with zero entries of that `match_type` = "allow all".

`match_type` is one of:

| `match_type`    | Compares against |
|-----------------|----------------------------|
| `username`      | user's exact username (case-insensitive) |
| `ip_cidr`       | client IP (literal or CIDR) |
| `email_domain`  | `@<pattern>` suffix of the user's email |

## Schema

See [db.sql](../db.sql) and the `models/db_models/` package. Key points:

- All foreign keys use `ON DELETE CASCADE` — deleting a group removes its
  pivots; deleting an NE removes its commands and the pivots pointing at it.
- `UNIQUE(ne_id, service, cmd_text)` on `command` catches duplicate
  registration at insert time.
- `UNIQUE(list_type, match_type, pattern)` on `user_access_list` prevents
  two "blacklist username eve" rows.

## HTTP surface

The minimal set needed to administer + query the v2 model lives under `/aa`.
See the top-level [README.md](../README.md) for the full list, and
[pkg/handler/frontend.html](../pkg/handler/frontend.html)'s Guide tab for the
same list rendered inline.

## Import / Export

CSV import/export is available for users, NEs, and commands:

- **Export** (`GET /aa/export/{users,nes,commands}`): any authenticated user,
  returns CSV download. Auth via `Authorization` header or `_token` query param.
- **Import** (`POST /aa/import/{users,nes,commands}`): admin+ required,
  multipart file upload. Skips duplicates, reports created/skipped/errors.

The frontend has `⬇ Export CSV` and `⬆ Import CSV` buttons on the Users, NEs,
and Commands tabs (import button hidden for `user` role).

## Evaluator tests

[pkg/service/authorize_test.go](../pkg/service/authorize_test.go) covers:

1. Deny when user is in no groups at all
2. Allow when both layers grant
3. Deny when the user is locked
4. Deny when the command is registered on a different NE than the target

[pkg/service/auth_test.go](../pkg/service/auth_test.go) covers the login-time
gates: wrong-password failure-count, lockout threshold, disabled account,
username blacklist.

Both suites run against the in-memory [`testutil.MockStore`](../pkg/testutil/mock_store.go).
