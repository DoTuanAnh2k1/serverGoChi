# Java client for cli-mgt-svc

Single-package, zero-dependency Java 17+ client. Drop the three files into any
project, change the package name, and use.

## Files

- `MgtServiceClient.java` — all-static client; one method per HTTP endpoint plus typed wrappers.
- `MgtModels.java` — record types for every response shape + a minimal JSON parser.
- `MgtHistoryClient.java` — standalone example covering only `POST /aa/history/save`.

## State model

- **Shared, set once:** `baseUrl` and the shared `HttpClient`, configured via
  `MgtServiceClient.init(url)`.
- **Per call:** the caller passes the `token` as the first argument to every
  authorized method. The client does not own or cache it — you keep the token
  wherever it fits your session model.

The token string already contains the `"Basic "` prefix — use it verbatim.

## Quick start

```java
// Once at startup
MgtServiceClient.init("http://mgt-svc:3000");

// 1. Login — caller keeps the returned token per user / session.
MgtModels.AuthResult auth = MgtServiceClient.authenticateTyped("alice", "pass");
String token = auth.token();          // "Basic eyJ..."

// 2. Authorized calls: token as the first argument.
List<MgtModels.Ne> nes = MgtServiceClient.listNeTyped(token);
nes.forEach(n -> System.out.println(n.ne() + " → " + n.confMasterIp()));

// 3. Save command history — endpoint is unauthenticated; always pass the
//    actor's username so the audit trail is meaningful.
MgtServiceClient.historySave(
        "show running-config", "HTSMF01", "10.10.1.1", "ne-command", "success", "alice");

// 4. RBAC — cache effective permissions once per session, check per-command for high-risk ops.
MgtServiceClient.Response effRes = MgtServiceClient.authorizeEffective(token);
MgtModels.EffectiveResponse eff = effRes.as(MgtModels.EffectiveResponse::from);

MgtServiceClient.Response ckRes = MgtServiceClient.authorizeCheckCommand(
        token, "ne-command", "delete session 1", /*ne_id*/ 5L);
MgtModels.CheckCommandResult ck = ckRes.as(MgtModels.CheckCommandResult::from);
if (!Boolean.TRUE.equals(ck.allowed())) { throw new RuntimeException("denied: " + ck.reason()); }

// 4. Typed history query.
List<MgtModels.History> recent =
        MgtServiceClient.historyListTyped(token, 50, "ne-command", null, "alice");
```

## Raw vs. typed methods

Every endpoint has a **raw** variant returning `Response(int status, String body)`.
Endpoints that return JSON objects/arrays also have a `*Typed` variant that
parses the body into a `MgtModels.*` record and throws `MgtApiException` if the
status is not 2xx.

| Raw                      | Typed                       | Returns                                 |
|--------------------------|-----------------------------|-----------------------------------------|
| `authenticate(u, p)`     | `authenticateTyped(u, p)`   | `MgtModels.AuthResult`                  |
| `validateToken(t)`       | `validateTokenTyped(t)`     | `MgtModels.ValidateTokenResult`         |
| `showUsers(tok)`         | `showUsersTyped(tok)`       | `List<MgtModels.UserShow>`              |
| `authorizeUserShow(tok)` | `authorizeUserShowTyped`    | `List<MgtModels.UserPermission>`        |
| `neShow(tok)`            | `neShowTyped(tok)`          | `List<MgtModels.NeShow>`                |
| `listNe(tok)`            | `listNeTyped(tok)`          | `List<MgtModels.Ne>`                    |
| `listNeMonitor(tok)`     | `listNeMonitorTyped(tok)`   | `List<MgtModels.NeMonitor>`             |
| `neConfigList(tok)`      | `neConfigListTyped(tok)`    | `List<MgtModels.NeConfig>`              |
| `configBackupSave(tok,…)`| `configBackupSaveTyped`     | `MgtModels.ConfigBackupSaveResult`      |
| `configBackupList(tok,…)`| `configBackupListTyped`     | `List<MgtModels.ConfigBackup>`          |
| `configBackupGet(tok,id)`| `configBackupGetTyped`      | `MgtModels.ConfigBackupDetail`          |
| `historyList(tok,…)`     | `historyListTyped`          | `List<MgtModels.History>`               |
| `adminUserList(tok)`     | `adminUserListTyped`        | `List<MgtModels.AdminUser>`             |
| `adminUserFull(tok)`     | `adminUserFullTyped`        | `List<MgtModels.AdminUserFull>`         |
| `adminNeList(tok)`       | `adminNeListTyped`          | `List<MgtModels.CliNe>`                 |
| `groupList(tok)`         | `groupListTyped`            | `List<MgtModels.Group>`                 |
| `groupShow(tok, id)`     | `groupShowTyped`            | `MgtModels.GroupDetail`                 |
| `userGroupList(tok, u)`  | `userGroupListTyped`        | `List<MgtModels.Group>`                 |
| `groupNeList(tok, gid)`  | `groupNeListTyped`          | `List<Long>` — NE ids in the group      |
| `importBulk(tok, body)`  | `importBulkTyped`           | `List<MgtModels.ImportResult>`          |
| `subscribersFiles(tok)`  | `subscribersFilesTyped`     | `List<MgtModels.SubscriberFile>`        |
| `subscribersFile(tok,i)` | `subscribersFileTyped`      | `MgtModels.SubscriberFileContent`       |

Write/update endpoints (`*Create`, `*Update`, `*Delete`, `*Assign`, etc.) return
the generic `Response` — inspect `status()` or `as(MgtModels.Envelope::from)`
for `{status, code, message}` payloads.

### No-auth endpoints

`health()`, `healthDb()`, `authenticate(u, p)`, and `validateToken(t)` do not
take a token parameter — they do not require an Authorization header.

### Ad-hoc calls

If an endpoint isn't covered here, use the generic helpers:

```java
// With auth
MgtServiceClient.getAuth(token,  "/aa/something");
MgtServiceClient.postAuth(token, "/aa/something",
        MgtServiceClient.toJson(MgtServiceClient.map("k", "v")));

// No auth
MgtServiceClient.postUnauth("/health-like-path", "...");
```

And parse:

```java
Object parsed = MgtModels.parse(response.body());  // Map / List / String / Long / Double / Boolean / null
```

## Endpoint coverage

All endpoints in `api.yaml`, grouped:

- Auth: `authenticate`, `validateToken`, `changePassword`
- Health: `health`, `healthDb`
- User mgmt (admin): `createUser`, `disableUser`, `showUsers`, `adminResetPassword`
- User authz: `authorizeUserSet`, `authorizeUserDelete`, `authorizeUserShow`
- Network Element: `neShow`, `neCreate`, `neUpdate`, `neRemove`, `neAssignToUser`, `neUnassignFromUser`
- Lists: `listNe`, `listNeMonitor`
- NE Config: `neConfigCreate`, `neConfigList`, `neConfigUpdate`, `neConfigDelete`
- Config Backup: `configBackupSave`, `configBackupList`, `configBackupGet`
- History: `historyList`, `historySave`
- Admin (frontend API): `adminUserList`, `adminUserFull`, `adminUserUpdate`, `adminNeList`, `adminNeCreate`, `adminNeUpdate`
- Groups: `groupList`, `groupShow`, `groupCreate`, `groupUpdate`, `groupDelete`,
  `userGroupList`, `userGroupAssign`, `userGroupUnassign`,
  `groupNeList`, `groupNeAssign`, `groupNeUnassign`
- RBAC (docs/rbac-design.md §4.7):
  - NE Profile: `neProfileList`, `neProfileCreate`, `neProfileUpdate`, `neProfileDelete`, `neAssignProfile`
  - Command Def: `commandDefList`, `commandDefCreate`, `commandDefUpdate`, `commandDefDelete`, `commandDefImport`
  - Command Group: `commandGroupList`, `commandGroupCreate`, `commandGroupUpdate`, `commandGroupDelete`,
    `commandGroupMembers`, `commandGroupAddMember`, `commandGroupRemoveMember`
  - Group cmd permission: `groupCmdPermissionList`, `groupCmdPermissionAdd`, `groupCmdPermissionDelete`
  - Authorize: `authorizeEffective`, `authorizeCheckCommand` — downstream services (ne-command / ne-config) should
    cache `authorizeEffective` per session and call `authorizeCheckCommand` realtime for risky operations
- Password Policy (docs/rbac-design.md §4.8): `passwordPolicyList`, `passwordPolicyCreate`, `passwordPolicyUpdate`,
  `passwordPolicyDelete`, `groupAssignPasswordPolicy`
- Mgt Permission (docs/rbac-design.md §4.11): `mgtPermissionList`, `mgtPermissionAdd`, `mgtPermissionDelete`
- Import: `importBulk`
- Subscribers: `subscribersFiles`, `subscribersFile`

### History save — unauthenticated + account field

`POST /aa/history/save` no longer requires a JWT. The caller should always
supply `account` = username of the actor that ran the command; when omitted
the server records `"unknown"` and the audit trail loses identity.

```java
// Correct — downstream resolves the username from its own auth layer and passes it here.
MgtServiceClient.historySave(
        "show running-config", "HTSMF01", "10.10.1.1", "ne-command", "success", "alice");

// Also works for minimal services (use MgtHistoryClient) — same account semantics.
MgtHistoryClient.saveHistory(
        "http://mgt-svc:3000",
        "show running-config", "HTSMF01", "10.10.1.1", "ne-command", "success", "alice");
```
