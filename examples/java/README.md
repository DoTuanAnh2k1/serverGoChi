# Java client for cli-mgt-svc (v2)

Single-file, zero-dependency Java 11+ client. Drop [MgtServiceClient.java](MgtServiceClient.java)
into your project, adjust the package if needed, and use it.

## Quick sample

```java
MgtServiceClient c = new MgtServiceClient("http://localhost:3000");
c.authenticate("admin", "admin");

// List inventory
c.listUsers();
c.listNEs();
c.listCommands();

// Wire up a group
Map<String, Object> g = c.createNeAccessGroup("site-hcmc", "HCMC operators");
c.addUserToNeAccessGroup((long) g.get("id"), 42L);
c.addNeToNeAccessGroup((long) g.get("id"), 7L);

// Ask the service the one question that matters
MgtServiceClient.AuthorizeDecision d = c.authorizeCheck("alice", 7L, 19L);
if (d.allowed) { /* proceed */ }
else { System.err.println("denied: " + d.reason); }
```

## What's covered

Every v2 endpoint under `/aa/*`:

- Auth: `authenticate`, `validateToken`, `changePassword`
- Users: CRUD + `resetPassword`
- NEs: CRUD
- Commands: CRUD, filter by `(ne_id, service)`
- NE access groups: CRUD + user/NE membership
- Cmd exec groups: CRUD + user/command membership
- Password policy: get / upsert
- Access list: list / create / delete
- **Authorize check**: `authorizeCheck(username, neId, commandId)` returns a
  typed decision with the trace flags (`userExists`, `userEnabled`,
  `neReachable`, `commandOnNe`, `commandExecAllowed`)
- History: `listHistory`, unauthenticated `saveHistory` (for proxy audit push)
- Config backup: save / list / get

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
