# Java client for cli-mgt-svc

Single-file, zero-dependency Java 17+ client. Drop into any project, change the package, and use.

## Files

- `MgtServiceClient.java` — full client; one method per HTTP endpoint.
- `MgtHistoryClient.java` — minimal example covering only `POST /aa/history/save`.

## Quick start

```java
MgtServiceClient c = new MgtServiceClient("http://localhost:3000");

// 1. Login — token is stored on the client, auto-attached on subsequent calls.
c.authenticate("anhdt195", "123");

// 2. Read endpoints — return Response(status, body).
MgtServiceClient.Response r = c.listNe();
System.out.println(r.status() + " " + r.body());

// 3. Write endpoints with many fields — pass a Map.
c.adminNeCreate(MgtServiceClient.map(
    "ne_name", "HTSMF03",
    "namespace", "hcm-5gc",
    "conf_master_ip", "10.10.1.1",
    "conf_port_master_tcp", 830,
    "command_url", "http://10.10.1.1:8080"
));

// 4. Save command history (called from SSH server / cli-netconf).
c.historySave("show running-config", "HTSMF01", "10.10.1.1", "ne-command", "success");
```

## What you get

`Response` is a record `(int status, String body)`. Body is the raw JSON string —
parse it with Jackson / Gson on your side if you want typed objects.

The client uses only `java.net.http.HttpClient` and `java.util.Map`. No external
dependencies, no annotations, no codegen.

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
c.get("/aa/something");
c.post("/aa/something", MgtServiceClient.toJson(MgtServiceClient.map("k", "v")));
```
