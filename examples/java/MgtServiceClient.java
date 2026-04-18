package com.example.mgt;

import java.net.URI;
import java.net.URLEncoder;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpRequest.BodyPublishers;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.ArrayList;
import java.util.Collection;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.function.Function;

/**
 * Static Java client for cli-mgt-svc HTTP API.
 *
 * Usage:
 *   MgtServiceClient.init("http://mgt-svc:3000");          // once at startup
 *   MgtServiceClient.authenticate("svc-account", "***");   // sets token internally
 *   MgtServiceClient.Response r = MgtServiceClient.listNe();
 *   System.out.println(r.status() + " " + r.body());
 *
 * Notes:
 *  - Single global state — one baseUrl + one token per JVM. NOT thread-safe across
 *    concurrent users; intended for service accounts (e.g. SSH server logging into
 *    mgt-service with a single account).
 *  - Token from /aa/authenticate already contains the "Basic " prefix; this client
 *    stores and forwards it verbatim in the Authorization header.
 *  - Every method returns Response(status, body); body is the raw JSON string.
 *    Pair this with Jackson/Gson if you want typed parsing on the caller side.
 *  - Java 17+ — uses java.net.http.HttpClient and records, no external dependencies.
 */
public final class MgtServiceClient {

    private static final HttpClient HTTP = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(5))
            .build();

    private static volatile String baseUrl;
    private static volatile String token; // includes "Basic " prefix

    private MgtServiceClient() {}

    /** Configure the mgt-service base URL. Call once at startup. */
    public static void init(String url) {
        if (url == null || url.isEmpty()) throw new IllegalArgumentException("baseUrl required");
        baseUrl = url.endsWith("/") ? url.substring(0, url.length() - 1) : url;
    }

    /** Inject a token previously obtained from /aa/authenticate (must include "Basic " prefix). */
    public static void setToken(String t) { token = t; }
    public static String getToken()       { return token; }

    /** Drop stored token (useful for tests / logout). */
    public static void clearToken() { token = null; }

    /** Status + body container with typed parse helpers. */
    public record Response(int status, String body) {
        public boolean ok() { return status >= 200 && status < 300; }

        /** Parse body and map through factory (e.g. {@code MgtModels.AuthResult::from}). */
        public <T> T as(Function<Object, T> factory) {
            return factory.apply(MgtModels.parse(body));
        }

        /** Parse body as a JSON array and map each item through factory. Returns empty list if body is not an array. */
        public <T> List<T> asList(Function<Object, T> factory) {
            Object v = MgtModels.parse(body);
            if (!(v instanceof List)) return List.of();
            List<T> out = new ArrayList<>();
            for (Object item : (List<?>) v) {
                T parsed = factory.apply(item);
                if (parsed != null) out.add(parsed);
            }
            return out;
        }
    }

    // ── Health / Docs ────────────────────────────────────────────────────────

    public static Response health() throws Exception {
        return get("/health");
    }

    public static Response healthDb() throws Exception {
        return get("/aa/heath-check-db");
    }

    // ── Auth ─────────────────────────────────────────────────────────────────

    /** POST /aa/authenticate. On 200, stores the token statically and returns the response. */
    public static Response authenticate(String username, String password) throws Exception {
        Response r = postRaw("/aa/authenticate", toJson(map("username", username, "password", password)), false);
        if (r.ok()) {
            MgtModels.AuthResult auth = r.as(MgtModels.AuthResult::from);
            if (auth != null && auth.token() != null && !auth.token().isEmpty()) token = auth.token();
        }
        return r;
    }

    /** POST /aa/validate-token. Body: {"token": "<basic-prefixed jwt>"} */
    public static Response validateToken(String tokenWithBasicPrefix) throws Exception {
        return postRaw("/aa/validate-token", toJson(map("token", tokenWithBasicPrefix)), false);
    }

    /** POST /aa/change-password/. Self-service — username must equal the caller. */
    public static Response changePassword(String username, String oldPassword, String newPassword) throws Exception {
        return post("/aa/change-password/", toJson(map(
                "username", username,
                "old_password", oldPassword,
                "new_password", newPassword)));
    }

    // ── User Management (admin) ──────────────────────────────────────────────

    /**
     * POST /aa/authenticate/user/set — create or re-enable a user.
     * Required keys in extras: account_name, password. Optional: full_name, email, address,
     * phone_number, avatar, description, account_type (0/1/2), only_ad.
     */
    public static Response createUser(Map<String, Object> userFields) throws Exception {
        return post("/aa/authenticate/user/set", toJson(userFields));
    }

    /** POST /aa/authenticate/user/delete — disable a user. */
    public static Response disableUser(String accountName) throws Exception {
        return post("/aa/authenticate/user/delete", toJson(map("account_name", accountName)));
    }

    /** GET /aa/authenticate/user/show — list users with NEs and roles. */
    public static Response showUsers() throws Exception {
        return get("/aa/authenticate/user/show");
    }

    /** POST /aa/authenticate/user/reset-password — admin reset. */
    public static Response adminResetPassword(String username, String newPassword) throws Exception {
        return post("/aa/authenticate/user/reset-password", toJson(map(
                "username", username,
                "new_password", newPassword)));
    }

    // ── User Authorization ───────────────────────────────────────────────────

    /** POST /aa/authorize/user/set — set permission ("admin" → account_type=1, "user" → 2). */
    public static Response authorizeUserSet(String username, String permission) throws Exception {
        return post("/aa/authorize/user/set", toJson(map(
                "username", username,
                "permission", permission)));
    }

    /** POST /aa/authorize/user/delete — reset permission to "user". */
    public static Response authorizeUserDelete(String username) throws Exception {
        return post("/aa/authorize/user/delete", toJson(map("username", username)));
    }

    /** GET /aa/authorize/user/show — list user-permission entries. */
    public static Response authorizeUserShow() throws Exception {
        return get("/aa/authorize/user/show");
    }

    // ── Network Element ──────────────────────────────────────────────────────

    /** GET /aa/authorize/ne/show — list 5GC NEs (basic fields). */
    public static Response neShow() throws Exception {
        return get("/aa/authorize/ne/show");
    }

    /**
     * POST /aa/authorize/ne/create — minimum {ne_name}.
     * Optional: namespace, site_name, system_type, description, command_url,
     * conf_mode, conf_master_ip, conf_slave_ip, conf_port_master_ssh,
     * conf_port_master_tcp, conf_port_slave_ssh, conf_port_slave_tcp,
     * conf_username, conf_password.
     */
    public static Response neCreate(Map<String, Object> neFields) throws Exception {
        return post("/aa/authorize/ne/create", toJson(neFields));
    }

    /** POST /aa/authorize/ne/update — id required, other fields optional (partial update). */
    public static Response neUpdate(Map<String, Object> neFields) throws Exception {
        return post("/aa/authorize/ne/update", toJson(neFields));
    }

    /** POST /aa/authorize/ne/remove — delete NE + cascade mappings/configs/backups. */
    public static Response neRemove(long id) throws Exception {
        return post("/aa/authorize/ne/remove", toJson(map("id", id)));
    }

    /** POST /aa/authorize/ne/set — assign NE to user. neid is the NE ID as a string. */
    public static Response neAssignToUser(String username, String neId) throws Exception {
        return post("/aa/authorize/ne/set", toJson(map(
                "username", username,
                "neid", neId)));
    }

    /** POST /aa/authorize/ne/delete — remove NE from user. */
    public static Response neUnassignFromUser(String username, String neId) throws Exception {
        return post("/aa/authorize/ne/delete", toJson(map(
                "username", username,
                "neid", neId)));
    }

    // ── List endpoints (any authenticated user) ──────────────────────────────

    /** GET /aa/list/ne — full NE fields the caller is authorized to access. */
    public static Response listNe() throws Exception {
        return get("/aa/list/ne");
    }

    /** GET /aa/list/ne/monitor — NE name + monitor URL. */
    public static Response listNeMonitor() throws Exception {
        return get("/aa/list/ne/monitor");
    }

    // ── NE Config ────────────────────────────────────────────────────────────

    /**
     * POST /aa/authorize/ne/config/create.
     * Required: ne_id, ip_address. Optional: port (default 22), username, password,
     * conf_mode, description.
     */
    public static Response neConfigCreate(Map<String, Object> fields) throws Exception {
        return post("/aa/authorize/ne/config/create", toJson(fields));
    }

    /** GET /aa/authorize/ne/config/list — list NE configs (includes ne_name + password). */
    public static Response neConfigList() throws Exception {
        return get("/aa/authorize/ne/config/list");
    }

    /** POST /aa/authorize/ne/config/update — id required, other fields optional. */
    public static Response neConfigUpdate(Map<String, Object> fields) throws Exception {
        return post("/aa/authorize/ne/config/update", toJson(fields));
    }

    /** POST /aa/authorize/ne/config/delete. */
    public static Response neConfigDelete(long id) throws Exception {
        return post("/aa/authorize/ne/config/delete", toJson(map("id", id)));
    }

    // ── Config Backup ────────────────────────────────────────────────────────

    /** POST /aa/config-backup/save — XML stored on disk, metadata in DB. */
    public static Response configBackupSave(String neName, String neIp, String configXml) throws Exception {
        return post("/aa/config-backup/save", toJson(map(
                "ne_name", neName,
                "ne_ip", neIp,
                "config_xml", configXml)));
    }

    /** GET /aa/config-backup/list?ne_name=&lt;optional&gt; — pass null to list all. */
    public static Response configBackupList(String neNameFilter) throws Exception {
        String q = (neNameFilter == null || neNameFilter.isEmpty())
                ? "" : "?ne_name=" + urlEncode(neNameFilter);
        return get("/aa/config-backup/list" + q);
    }

    /** GET /aa/config-backup/{id} — returns metadata + full XML. */
    public static Response configBackupGet(long id) throws Exception {
        return get("/aa/config-backup/" + id);
    }

    // ── History ──────────────────────────────────────────────────────────────

    /**
     * GET /aa/history/list — all params optional.
     * @param limit    1..500 (default 100). 0 to omit.
     * @param scope    "cli-config" | "ne-command" | "ne-config" or null.
     * @param neName   filter by NE name, or null.
     * @param account  filter by username, or null.
     */
    public static Response historyList(int limit, String scope, String neName, String account) throws Exception {
        StringBuilder q = new StringBuilder();
        if (limit > 0)                                    appendQuery(q, "limit",   String.valueOf(limit));
        if (scope   != null && !scope.isEmpty())          appendQuery(q, "scope",   scope);
        if (neName  != null && !neName.isEmpty())         appendQuery(q, "ne_name", neName);
        if (account != null && !account.isEmpty())        appendQuery(q, "account", account);
        return get("/aa/history/list" + q);
    }

    /**
     * POST /aa/history/save — used by SSH server / cli-netconf-svc to log a command.
     * cmd_name + ne_name are required; ne_ip / scope / result optional.
     * account is taken from the JWT — do not send it.
     */
    public static Response historySave(String cmdName, String neName, String neIp,
                                        String scope, String result) throws Exception {
        return post("/aa/history/save", toJson(map(
                "cmd_name", cmdName,
                "ne_name",  neName,
                "ne_ip",    neIp,
                "scope",    scope,
                "result",   result)));
    }

    // ── Admin (frontend API) ─────────────────────────────────────────────────

    /** GET /aa/admin/user/list — full user objects, no password. */
    public static Response adminUserList() throws Exception {
        return get("/aa/admin/user/list");
    }

    /**
     * POST /aa/admin/user/update — update metadata.
     * Required: account_name. Optional: full_name, email, phone_number, address,
     * description, account_type (1 or 2 — cannot set 0).
     * Normal users may call this only for their own account_name; account_type
     * is preserved server-side for non-admins.
     */
    public static Response adminUserUpdate(Map<String, Object> fields) throws Exception {
        return post("/aa/admin/user/update", toJson(fields));
    }

    /** GET /aa/admin/ne/list — full NE objects (includes credentials). */
    public static Response adminNeList() throws Exception {
        return get("/aa/admin/ne/list");
    }

    /**
     * POST /aa/admin/ne/create — create with full CliNe schema.
     * Required: ne_name. See CliNe schema for all fields.
     */
    public static Response adminNeCreate(Map<String, Object> fields) throws Exception {
        return post("/aa/admin/ne/create", toJson(fields));
    }

    /** POST /aa/admin/ne/update — id required, other fields optional. */
    public static Response adminNeUpdate(Map<String, Object> fields) throws Exception {
        return post("/aa/admin/ne/update", toJson(fields));
    }

    // ── Import ───────────────────────────────────────────────────────────────

    /** POST /aa/import/ — body is plain text in section format ([users], [nes], ...). */
    public static Response importBulk(String plainTextBody) throws Exception {
        return postRawText("/aa/import/", plainTextBody);
    }

    // ── Subscribers (TCP-collected files) ────────────────────────────────────

    /** GET /aa/subscribers/files — list available subscriber result files. */
    public static Response subscribersFiles() throws Exception {
        return get("/aa/subscribers/files");
    }

    /** GET /aa/subscribers/files/{index} — fetch file content by index. */
    public static Response subscribersFile(int index) throws Exception {
        return get("/aa/subscribers/files/" + index);
    }

    // ── Typed convenience wrappers ───────────────────────────────────────────
    //
    // These call the matching raw endpoint above, then parse the body into a
    // record from MgtModels. They throw MgtApiException if status is not 2xx —
    // use the raw Response-returning variants if you want to inspect failures.

    /** Like {@link #authenticate} but returns the parsed body. Token is stored internally as usual. */
    public static MgtModels.AuthResult authenticateTyped(String u, String p) throws Exception {
        return must(authenticate(u, p)).as(MgtModels.AuthResult::from);
    }

    /** Parsed {@link #validateToken}. */
    public static MgtModels.ValidateTokenResult validateTokenTyped(String tokenWithBasicPrefix) throws Exception {
        return must(validateToken(tokenWithBasicPrefix)).as(MgtModels.ValidateTokenResult::from);
    }

    /** Parsed {@link #showUsers}. */
    public static List<MgtModels.UserShow> showUsersTyped() throws Exception {
        return must(showUsers()).asList(MgtModels.UserShow::from);
    }

    /** Parsed {@link #authorizeUserShow}. */
    public static List<MgtModels.UserPermission> authorizeUserShowTyped() throws Exception {
        return must(authorizeUserShow()).asList(MgtModels.UserPermission::from);
    }

    /** Parsed {@link #neShow}. */
    public static List<MgtModels.NeShow> neShowTyped() throws Exception {
        return must(neShow()).asList(MgtModels.NeShow::from);
    }

    /** Parsed {@link #listNe}. */
    public static List<MgtModels.Ne> listNeTyped() throws Exception {
        return must(listNe()).asList(MgtModels.Ne::from);
    }

    /** Parsed {@link #listNeMonitor}. */
    public static List<MgtModels.NeMonitor> listNeMonitorTyped() throws Exception {
        return must(listNeMonitor()).asList(MgtModels.NeMonitor::from);
    }

    /** Parsed {@link #neConfigList}. */
    public static List<MgtModels.NeConfig> neConfigListTyped() throws Exception {
        return must(neConfigList()).asList(MgtModels.NeConfig::from);
    }

    /** Parsed {@link #configBackupList}. Returns only the {@code backups} array. */
    public static List<MgtModels.ConfigBackup> configBackupListTyped(String neNameFilter) throws Exception {
        MgtModels.ConfigBackupListResult wrapped =
                must(configBackupList(neNameFilter)).as(MgtModels.ConfigBackupListResult::from);
        return wrapped == null || wrapped.backups() == null ? List.of() : wrapped.backups();
    }

    /** Parsed {@link #configBackupGet}. */
    public static MgtModels.ConfigBackupDetail configBackupGetTyped(long id) throws Exception {
        return must(configBackupGet(id)).as(MgtModels.ConfigBackupDetail::from);
    }

    /** Parsed {@link #configBackupSave}. */
    public static MgtModels.ConfigBackupSaveResult configBackupSaveTyped(String neName, String neIp, String xml) throws Exception {
        return must(configBackupSave(neName, neIp, xml)).as(MgtModels.ConfigBackupSaveResult::from);
    }

    /** Parsed {@link #historyList}. */
    public static List<MgtModels.History> historyListTyped(int limit, String scope, String neName, String account) throws Exception {
        return must(historyList(limit, scope, neName, account)).asList(MgtModels.History::from);
    }

    /** Parsed {@link #adminUserList}. */
    public static List<MgtModels.AdminUser> adminUserListTyped() throws Exception {
        return must(adminUserList()).asList(MgtModels.AdminUser::from);
    }

    /** Parsed {@link #adminNeList}. */
    public static List<MgtModels.CliNe> adminNeListTyped() throws Exception {
        return must(adminNeList()).asList(MgtModels.CliNe::from);
    }

    /** Parsed {@link #importBulk}. */
    public static List<MgtModels.ImportResult> importBulkTyped(String body) throws Exception {
        return must(importBulk(body)).asList(MgtModels.ImportResult::from);
    }

    /** Parsed {@link #subscribersFiles}. */
    public static List<MgtModels.SubscriberFile> subscribersFilesTyped() throws Exception {
        return must(subscribersFiles()).asList(MgtModels.SubscriberFile::from);
    }

    /** Parsed {@link #subscribersFile}. */
    public static MgtModels.SubscriberFileContent subscribersFileTyped(int index) throws Exception {
        return must(subscribersFile(index)).as(MgtModels.SubscriberFileContent::from);
    }

    /** Thrown by *Typed methods when the HTTP status is not 2xx. */
    public static final class MgtApiException extends RuntimeException {
        public final int status;
        public final String responseBody;
        public MgtApiException(int status, String body) {
            super("mgt-svc call failed: " + status + " " + body);
            this.status = status;
            this.responseBody = body;
        }
    }

    private static Response must(Response r) {
        if (!r.ok()) throw new MgtApiException(r.status(), r.body());
        return r;
    }

    // ── Generic helpers (also exposed for ad-hoc calls) ──────────────────────

    public static Response get(String path) throws Exception {
        return send(builder(path).GET().build());
    }

    public static Response post(String path, String jsonBody) throws Exception {
        return postRaw(path, jsonBody, true);
    }

    private static Response postRaw(String path, String jsonBody, boolean auth) throws Exception {
        requireBaseUrl();
        HttpRequest.Builder b = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + path))
                .timeout(Duration.ofSeconds(15))
                .header("Content-Type", "application/json");
        if (auth && token != null) b.header("Authorization", token);
        return send(b.POST(BodyPublishers.ofString(jsonBody, StandardCharsets.UTF_8)).build());
    }

    private static Response postRawText(String path, String body) throws Exception {
        HttpRequest req = builder(path)
                .header("Content-Type", "text/plain; charset=utf-8")
                .POST(BodyPublishers.ofString(body, StandardCharsets.UTF_8))
                .build();
        return send(req);
    }

    private static HttpRequest.Builder builder(String path) {
        requireBaseUrl();
        HttpRequest.Builder b = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + path))
                .timeout(Duration.ofSeconds(15));
        if (token != null) b.header("Authorization", token);
        return b;
    }

    private static Response send(HttpRequest req) throws Exception {
        HttpResponse<String> r = HTTP.send(req, HttpResponse.BodyHandlers.ofString(StandardCharsets.UTF_8));
        return new Response(r.statusCode(), r.body());
    }

    private static void requireBaseUrl() {
        if (baseUrl == null) throw new IllegalStateException("MgtServiceClient.init(baseUrl) must be called first");
    }

    // ── Tiny JSON / utility helpers (no external deps) ───────────────────────

    /** Build a Map literal — convenient for one-line JSON bodies. */
    public static Map<String, Object> map(Object... kv) {
        if (kv.length % 2 != 0) throw new IllegalArgumentException("map() requires key/value pairs");
        Map<String, Object> m = new LinkedHashMap<>();
        for (int i = 0; i < kv.length; i += 2) m.put(String.valueOf(kv[i]), kv[i + 1]);
        return m;
    }

    /** Serialize Map / Collection / String / Number / Boolean / null to compact JSON. */
    public static String toJson(Object v) {
        StringBuilder sb = new StringBuilder();
        writeJson(sb, v);
        return sb.toString();
    }

    private static void writeJson(StringBuilder sb, Object v) {
        if (v == null) {
            sb.append("null");
        } else if (v instanceof Boolean || v instanceof Number) {
            sb.append(v);
        } else if (v instanceof Map) {
            Map<?, ?> m = (Map<?, ?>) v;
            sb.append('{');
            boolean first = true;
            for (Map.Entry<?, ?> e : m.entrySet()) {
                if (!first) sb.append(',');
                first = false;
                writeJsonString(sb, String.valueOf(e.getKey()));
                sb.append(':');
                writeJson(sb, e.getValue());
            }
            sb.append('}');
        } else if (v instanceof Collection) {
            Collection<?> c = (Collection<?>) v;
            sb.append('[');
            boolean first = true;
            for (Object item : c) {
                if (!first) sb.append(',');
                first = false;
                writeJson(sb, item);
            }
            sb.append(']');
        } else {
            writeJsonString(sb, String.valueOf(v));
        }
    }

    private static void writeJsonString(StringBuilder sb, String s) {
        sb.append('"');
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            switch (c) {
                case '"':  sb.append("\\\""); break;
                case '\\': sb.append("\\\\"); break;
                case '\n': sb.append("\\n");  break;
                case '\r': sb.append("\\r");  break;
                case '\t': sb.append("\\t");  break;
                default:
                    if (c < 0x20) sb.append(String.format("\\u%04x", (int) c));
                    else sb.append(c);
            }
        }
        sb.append('"');
    }

    private static void appendQuery(StringBuilder q, String key, String value) {
        q.append(q.length() == 0 ? '?' : '&').append(key).append('=').append(urlEncode(value));
    }

    private static String urlEncode(String s) {
        return URLEncoder.encode(s, StandardCharsets.UTF_8);
    }
}
