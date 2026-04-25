# Java client for cli-mgt-svc (v2)

Single-file, zero-dependency Java 11+ client with **9 typed model classes** and
**100% API coverage**. Drop [MgtServiceClient.java](MgtServiceClient.java)
into your project, adjust the package if needed, and use it.

## Quick sample

```java
MgtServiceClient c = new MgtServiceClient("http://localhost:3000");
String role = c.authenticate("admin", "admin");
System.out.println("logged in as " + role);  // "super_admin"

// Typed models — not raw maps
List<MgtServiceClient.User> users = c.listUsers();
for (User u : users) {
    System.out.println(u.username + " role=" + u.role);
}

// Create user with role
User newUser = c.createUser("operator1", "Pass1234!", "op1@vht.com", "Op 1", "admin");

// Wire up a group
MgtServiceClient.Group g = c.createNeAccessGroup("site-hcmc", "HCMC operators");
c.addUserToNeAccessGroup(g.id, newUser.id);

// The one question that matters
MgtServiceClient.AuthorizeDecision d = c.authorizeCheck("operator1", 7L, 19L);
if (d.allowed) { /* proceed */ }
else { System.err.println("denied: " + d.reason); }
```

## Typed models

| Class | Fields |
|-------|--------|
| `User` | id, username, email, fullName, phone, **role**, isEnabled, lockedAt, lastLoginAt, ... |
| `NE` | id, namespace, neType, siteName, masterIp, masterPort, confMode, ... |
| `Command` | id, neId, service, cmdText, description |
| `Group` | id, name, description |
| `AccessListEntry` | id, listType, matchType, pattern, reason |
| `PasswordPolicy` | minLength, maxAgeDays, historyCount, requireUppercase, ... |
| `AuthorizeDecision` | allowed, reason, userExists, userEnabled, neReachable, commandOnNe, commandExecAllowed |
| `HistoryEntry` | id, account, cmdText, neNamespace, scope, result, ... |
| `ConfigBackup` | id, neName, neIp, configXml, createdAt |

Each model has `fromMap()`, `toMap()`, and `toString()`.

## What's covered

Every v2 endpoint under `/aa/*` (100% coverage):

- **Auth**: `authenticate` (returns role), `validateToken`, `changePassword`
- **Users**: full CRUD + `resetPassword` — with **role** field support
- **NEs**: full CRUD
- **Commands**: full CRUD, filter by `(ne_id, service)`
- **NE access groups**: full CRUD + add/remove user + add/remove NE + list members
- **Cmd exec groups**: full CRUD + add/remove user + add/remove command + list members
- **Password policy**: get / upsert
- **Access list**: list / create / delete
- **Authorize check**: typed `AuthorizeDecision` with trace flags
- **History**: `listHistory`, unauthenticated `saveHistory` (for proxy audit push)
- **Config backup**: save / list / get

## Running the demo main

```bash
# From the repo root
javac -d /tmp/mgt-client examples/java/MgtServiceClient.java
MGT_BASE=http://localhost:3000 MGT_USER=admin MGT_PASS=admin \
    java -cp /tmp/mgt-client examples.java.MgtServiceClient
```

The included `main` logs in, prints the user / NE / command counts, and probes
`authorizeCheck` against the first row of each table.

## JSON / HTTP

The JSON encoder / decoder is hand-written and only covers what the v2 API
returns: `null`, booleans, numbers (long or double), strings, maps, and
lists. No Gson / Jackson dependency. If you already use a JSON lib, replace
`toJson(Object)` / `parseJson(String)` — the public method signatures don't
change.

HTTP uses JDK's built-in `java.net.http.HttpClient`. The auth token is sent
verbatim in the `Authorization` header (the `Basic <jwt>` prefix is returned
by `/aa/authenticate` — the client passes it back as-is).
