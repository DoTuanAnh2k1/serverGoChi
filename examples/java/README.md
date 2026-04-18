# Java client for cli-mgt-svc

Single-package, zero-dependency Java 17+ client. Drop the three files into any
project, change the package name, and use.

## Files

- `MgtServiceClient.java` — full client; one method per HTTP endpoint plus typed wrappers.
- `MgtModels.java` — record types for every response shape + a minimal JSON parser.
- `MgtHistoryClient.java` — standalone example covering only `POST /aa/history/save`.

## Quick start

```java
MgtServiceClient.init("http://localhost:3000");

// 1. Login — token is stored on the client, auto-attached on subsequent calls.
MgtServiceClient.authenticate("anhdt195", "123");

// 2a. Raw style — returns Response(status, body) with the JSON body as a String.
MgtServiceClient.Response r = MgtServiceClient.listNe();
System.out.println(r.status() + " " + r.body());

// 2b. Typed style — returns parsed records, throws MgtApiException on non-2xx.
List<MgtModels.Ne> nes = MgtServiceClient.listNeTyped();
for (MgtModels.Ne ne : nes) {
    System.out.println(ne.ne() + " → " + ne.confMasterIp() + ":" + ne.confPortMasterSsh());
}

// 3. Write endpoints with many fields — pass a Map to the raw method.
MgtServiceClient.adminNeCreate(MgtServiceClient.map(
    "ne_name", "HTSMF03",
    "namespace", "hcm-5gc",
    "conf_master_ip", "10.10.1.1",
    "conf_port_master_tcp", 830,
    "command_url", "http://10.10.1.1:8080"
));

// 4. Save command history (called from SSH server / cli-netconf).
MgtServiceClient.historySave("show running-config", "HTSMF01", "10.10.1.1", "ne-command", "success");

// 5. Typed history query.
List<MgtModels.History> recent =
    MgtServiceClient.historyListTyped(50, "ne-command", null, "anhdt195");
recent.forEach(h -> System.out.printf("%s  %-10s  %s%n",
        h.createdDate(), h.account(), h.cmdName()));
```

## Raw vs. typed methods

Every endpoint has a raw variant returning `Response(int status, String body)`.
Endpoints that return JSON objects/arrays also have a `*Typed` variant that
parses the body into a `MgtModels.*` record and throws `MgtApiException` if the
status is not 2xx.

| Raw                        | Typed                          | Returns                                 |
|----------------------------|--------------------------------|-----------------------------------------|
| `authenticate`             | `authenticateTyped`            | `MgtModels.AuthResult`                  |
| `validateToken`            | `validateTokenTyped`           | `MgtModels.ValidateTokenResult`         |
| `showUsers`                | `showUsersTyped`               | `List<MgtModels.UserShow>`              |
| `authorizeUserShow`        | `authorizeUserShowTyped`       | `List<MgtModels.UserPermission>`        |
| `neShow`                   | `neShowTyped`                  | `List<MgtModels.NeShow>`                |
| `listNe`                   | `listNeTyped`                  | `List<MgtModels.Ne>`                    |
| `listNeMonitor`            | `listNeMonitorTyped`           | `List<MgtModels.NeMonitor>`             |
| `neConfigList`             | `neConfigListTyped`            | `List<MgtModels.NeConfig>`              |
| `configBackupSave`         | `configBackupSaveTyped`        | `MgtModels.ConfigBackupSaveResult`      |
| `configBackupList`         | `configBackupListTyped`        | `List<MgtModels.ConfigBackup>`          |
| `configBackupGet`          | `configBackupGetTyped`         | `MgtModels.ConfigBackupDetail`          |
| `historyList`              | `historyListTyped`             | `List<MgtModels.History>`               |
| `adminUserList`            | `adminUserListTyped`           | `List<MgtModels.AdminUser>`             |
| `adminNeList`              | `adminNeListTyped`             | `List<MgtModels.CliNe>`                 |
| `importBulk`               | `importBulkTyped`              | `List<MgtModels.ImportResult>`          |
| `subscribersFiles`         | `subscribersFilesTyped`        | `List<MgtModels.SubscriberFile>`        |
| `subscribersFile`          | `subscribersFileTyped`         | `MgtModels.SubscriberFileContent`       |

Write/update endpoints (`*Create`, `*Update`, `*Delete`, `*Assign`, etc.) return
the generic `Response` — inspect `status()` or `as(MgtModels.Envelope::from)`
for `{status, code, message}` payloads.

### Ad-hoc parsing

If an endpoint isn't covered by a `*Typed` variant, parse on the fly:

```java
MgtServiceClient.Response r = MgtServiceClient.post("/aa/admin/user/update", ...);
MgtModels.Envelope env = r.as(MgtModels.Envelope::from);
if (Boolean.TRUE.equals(env.status())) ...
```

Or drop to the map/list level:

```java
Object parsed = MgtModels.parse(r.body());  // Map / List / String / Long / Double / Boolean / null
```

## Thread safety & state

`MgtServiceClient` keeps **a single global `baseUrl` + `token`** per JVM. It is
intended for service accounts (e.g. the SSH / cli-netconf side logging into
mgt-service with one account). Don't use it to multiplex many user sessions —
for that, drop the statics and reintroduce instance fields.

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
MgtServiceClient.get("/aa/something");
MgtServiceClient.post("/aa/something", MgtServiceClient.toJson(MgtServiceClient.map("k", "v")));
```
