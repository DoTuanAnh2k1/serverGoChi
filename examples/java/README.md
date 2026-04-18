# Java client for cli-mgt-svc

Single-package, zero-dependency Java 17+ client. Drop the three files into any
project, change the package name, and use.

## Files

- `MgtServiceClient.java` — client class. One **instance per user** (holds their token).
- `MgtModels.java` — record types for every response shape + a minimal JSON parser.
- `MgtHistoryClient.java` — standalone example covering only `POST /aa/history/save`.

## Shape of state

- **Shared (process-wide, static):** `baseUrl` + the underlying `HttpClient`. Set once at
  startup via `MgtServiceClient.init(url)`.
- **Per-user (instance field):** the `token` (JWT with `"Basic "` prefix). Each logged-in
  user holds their own `MgtServiceClient` instance.

This separation is important: do **not** share one client across users — the
token is per-instance, so concurrent users get their own.

## Quick start

```java
// 1. Once at startup — configure where mgt-svc lives.
MgtServiceClient.init("http://mgt-svc:3000");

// 2. Per user — login creates a dedicated client with the user's token already stored.
MgtServiceClient alice = MgtServiceClient.login("alice", "pass");

// 3. Use the client — the token rides on every request automatically.
List<MgtModels.Ne> nes = alice.listNeTyped();
for (MgtModels.Ne ne : nes) {
    System.out.println(ne.ne() + " → " + ne.confMasterIp() + ":" + ne.confPortMasterSsh());
}

// 4. Need the raw token string (e.g. to forward elsewhere)?
String jwt = alice.getToken();     // "Basic eyJ..."

// 5. Already have a token (e.g. forwarded from browser / another service)?
MgtServiceClient bob = MgtServiceClient.withToken("Basic eyJ...");
bob.historySave("show running-config", "HTSMF01", "10.10.1.1", "ne-command", "success");
```

## Raw vs. typed methods

Every endpoint has a **raw** variant returning `Response(int status, String body)`.
Endpoints that return JSON objects/arrays also have a `*Typed` variant that
parses the body into a `MgtModels.*` record and throws `MgtApiException` if the
status is not 2xx.

| Raw                     | Typed                     | Returns                                 |
|-------------------------|---------------------------|-----------------------------------------|
| `authenticate`          | `authenticateTyped`       | `MgtModels.AuthResult`                  |
| `validateToken`         | `validateTokenTyped`      | `MgtModels.ValidateTokenResult`         |
| `showUsers`             | `showUsersTyped`          | `List<MgtModels.UserShow>`              |
| `authorizeUserShow`     | `authorizeUserShowTyped`  | `List<MgtModels.UserPermission>`        |
| `neShow`                | `neShowTyped`             | `List<MgtModels.NeShow>`                |
| `listNe`                | `listNeTyped`             | `List<MgtModels.Ne>`                    |
| `listNeMonitor`         | `listNeMonitorTyped`      | `List<MgtModels.NeMonitor>`             |
| `neConfigList`          | `neConfigListTyped`       | `List<MgtModels.NeConfig>`              |
| `configBackupSave`      | `configBackupSaveTyped`   | `MgtModels.ConfigBackupSaveResult`      |
| `configBackupList`      | `configBackupListTyped`   | `List<MgtModels.ConfigBackup>`          |
| `configBackupGet`       | `configBackupGetTyped`    | `MgtModels.ConfigBackupDetail`          |
| `historyList`           | `historyListTyped`        | `List<MgtModels.History>`               |
| `adminUserList`         | `adminUserListTyped`      | `List<MgtModels.AdminUser>`             |
| `adminNeList`           | `adminNeListTyped`        | `List<MgtModels.CliNe>`                 |
| `importBulk`            | `importBulkTyped`         | `List<MgtModels.ImportResult>`          |
| `subscribersFiles`      | `subscribersFilesTyped`   | `List<MgtModels.SubscriberFile>`        |
| `subscribersFile`       | `subscribersFileTyped`    | `MgtModels.SubscriberFileContent`       |

Write/update endpoints (`*Create`, `*Update`, `*Delete`, `*Assign`, etc.) return
the generic `Response` — inspect `status()` or `as(MgtModels.Envelope::from)`
for `{status, code, message}` payloads.

### Ad-hoc parsing

If an endpoint isn't covered by a `*Typed` variant, parse on the fly:

```java
MgtServiceClient.Response r = client.post("/aa/admin/user/update", ...);
MgtModels.Envelope env = r.as(MgtModels.Envelope::from);
if (Boolean.TRUE.equals(env.status())) ...
```

Or drop to the map/list level:

```java
Object parsed = MgtModels.parse(r.body());  // Map / List / String / Long / Double / Boolean / null
```

## Building a per-user client cache

A common pattern is "one logged-in user → one `MgtServiceClient`". Keep a map
keyed by username (or session id) and look up per request:

```java
private static final Map<String, MgtServiceClient> SESSIONS = new ConcurrentHashMap<>();

public MgtServiceClient clientFor(String user, String pass) throws Exception {
    MgtServiceClient c = SESSIONS.get(user);
    if (c == null || !c.isAuthenticated()) {
        c = MgtServiceClient.login(user, pass);
        SESSIONS.put(user, c);
    }
    return c;
}
```

## Endpoint coverage

All 35 endpoints in `api.yaml`, grouped:

- Auth: `authenticate`, `validateToken`, `changePassword`
- Health: `health`, `healthDb`
- User mgmt (admin): `createUser`, `disableUser`, `showUsers`, `adminResetPassword`
- User authz: `authorizeUserSet`, `authorizeUserDelete`, `authorizeUserShow`
- Network Element: `neShow`, `neCreate`, `neUpdate`, `neRemove`, `neAssignToUser`, `neUnassignFromUser`
- Lists: `listNe`, `listNeMonitor`
- NE Config: `neConfigCreate`, `neConfigList`, `neConfigUpdate`, `neConfigDelete`
- Config Backup: `configBackupSave`, `configBackupList`, `configBackupGet`
- History: `historyList`, `historySave`
- Admin (frontend API): `adminUserList`, `adminUserUpdate`, `adminNeList`, `adminNeCreate`, `adminNeUpdate`
- Import: `importBulk`
- Subscribers: `subscribersFiles`, `subscribersFile`

Need an endpoint not listed here? Use the generic helpers:

```java
client.get("/aa/something");
client.post("/aa/something", MgtServiceClient.toJson(MgtServiceClient.map("k", "v")));
```
