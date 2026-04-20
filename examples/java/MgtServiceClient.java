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
 * Shared state:
 *   - {@code baseUrl} and the shared {@link HttpClient} — configured once via {@link #init}.
 *
 * Token handling:
 *   - The caller owns the token. Pass it as the first argument to every authorized
 *     method. The token string already includes the {@code "Basic "} prefix — use
 *     verbatim what {@link #authenticate} returned.
 *
 * Typical usage:
 * <pre>{@code
 *   // Once at startup
 *   MgtServiceClient.init("http://mgt-svc:3000");
 *
 *   // Login — caller keeps the returned token per user / session.
 *   String token = MgtServiceClient.authenticateTyped("alice", "pass").token();
 *
 *   // Subsequent authorized calls take the token explicitly.
 *   List<MgtModels.Ne> nes = MgtServiceClient.listNeTyped(token);
 *   MgtServiceClient.historySave(token, "show running-config", "HTSMF01",
 *                                "10.10.1.1", "ne-command", "success");
 * }</pre>
 *
 * Java 17+ — uses java.net.http.HttpClient and records, no external dependencies.
 */
public final class MgtServiceClient {

    private static final HttpClient HTTP = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(5))
            .build();

    private static volatile String baseUrl;

    private MgtServiceClient() {}

    /** Configure the mgt-service base URL. Call once at startup. */
    public static void init(String url) {
        if (url == null || url.isEmpty()) throw new IllegalArgumentException("baseUrl required");
        baseUrl = url.endsWith("/") ? url.substring(0, url.length() - 1) : url;
    }

    /** @return the configured base URL, or {@code null} if {@link #init} was not called. */
    public static String getBaseUrl() { return baseUrl; }

    // ── Response container ──────────────────────────────────────────────────

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

        /** Parse body as a JSON array of numbers. Returns empty list if body is not an array. */
        public List<Long> asLongList() {
            Object v = MgtModels.parse(body);
            if (!(v instanceof List)) return List.of();
            List<Long> out = new ArrayList<>();
            for (Object item : (List<?>) v) {
                if (item instanceof Number) out.add(((Number) item).longValue());
                else if (item != null) {
                    try { out.add(Long.parseLong(String.valueOf(item))); } catch (NumberFormatException ignored) {}
                }
            }
            return out;
        }
    }

    // ── Health / Docs (no auth) ──────────────────────────────────────────────

    public static Response health()   throws Exception { return send(unauth("/health").GET().build()); }
    public static Response healthDb() throws Exception { return send(unauth("/aa/heath-check-db").GET().build()); }

    // ── Auth (no token required to call these) ───────────────────────────────

    /** POST /aa/authenticate. Returns the raw response — extract the token with {@link Response#as}. */
    public static Response authenticate(String username, String password) throws Exception {
        return postUnauth("/aa/authenticate", toJson(map("username", username, "password", password)));
    }

    /** POST /aa/validate-token. {@code tokenToValidate} is the value being checked. */
    public static Response validateToken(String tokenToValidate) throws Exception {
        return postUnauth("/aa/validate-token", toJson(map("token", tokenToValidate)));
    }

    // ── Auth-required endpoints ──────────────────────────────────────────────
    // Every method below takes the caller's token as the first argument.

    /** POST /aa/change-password/. Self-service — {@code username} must equal the caller. */
    public static Response changePassword(String token, String username, String oldPassword, String newPassword) throws Exception {
        return postAuth(token, "/aa/change-password/", toJson(map(
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
    public static Response createUser(String token, Map<String, Object> userFields) throws Exception {
        return postAuth(token, "/aa/authenticate/user/set", toJson(userFields));
    }

    public static Response disableUser(String token, String accountName) throws Exception {
        return postAuth(token, "/aa/authenticate/user/delete", toJson(map("account_name", accountName)));
    }

    public static Response showUsers(String token) throws Exception {
        return getAuth(token, "/aa/authenticate/user/show");
    }

    public static Response adminResetPassword(String token, String username, String newPassword) throws Exception {
        return postAuth(token, "/aa/authenticate/user/reset-password", toJson(map(
                "username", username,
                "new_password", newPassword)));
    }

    // ── User Authorization ───────────────────────────────────────────────────

    public static Response authorizeUserSet(String token, String username, String permission) throws Exception {
        return postAuth(token, "/aa/authorize/user/set", toJson(map(
                "username", username,
                "permission", permission)));
    }

    public static Response authorizeUserDelete(String token, String username) throws Exception {
        return postAuth(token, "/aa/authorize/user/delete", toJson(map("username", username)));
    }

    public static Response authorizeUserShow(String token) throws Exception {
        return getAuth(token, "/aa/authorize/user/show");
    }

    // ── Network Element ──────────────────────────────────────────────────────

    public static Response neShow(String token) throws Exception {
        return getAuth(token, "/aa/authorize/ne/show");
    }

    /**
     * POST /aa/authorize/ne/create — minimum {ne_name}.
     * Optional: namespace, site_name, system_type, description, command_url,
     * conf_mode, conf_master_ip, conf_slave_ip, conf_port_master_ssh,
     * conf_port_master_tcp, conf_port_slave_ssh, conf_port_slave_tcp,
     * conf_username, conf_password.
     */
    public static Response neCreate(String token, Map<String, Object> neFields) throws Exception {
        return postAuth(token, "/aa/authorize/ne/create", toJson(neFields));
    }

    public static Response neUpdate(String token, Map<String, Object> neFields) throws Exception {
        return postAuth(token, "/aa/authorize/ne/update", toJson(neFields));
    }

    public static Response neRemove(String token, long id) throws Exception {
        return postAuth(token, "/aa/authorize/ne/remove", toJson(map("id", id)));
    }

    public static Response neAssignToUser(String token, String username, String neId) throws Exception {
        return postAuth(token, "/aa/authorize/ne/set", toJson(map(
                "username", username,
                "neid", neId)));
    }

    public static Response neUnassignFromUser(String token, String username, String neId) throws Exception {
        return postAuth(token, "/aa/authorize/ne/delete", toJson(map(
                "username", username,
                "neid", neId)));
    }

    // ── List endpoints ───────────────────────────────────────────────────────

    public static Response listNe(String token) throws Exception {
        return getAuth(token, "/aa/list/ne");
    }

    public static Response listNeMonitor(String token) throws Exception {
        return getAuth(token, "/aa/list/ne/monitor");
    }

    // ── NE Config ────────────────────────────────────────────────────────────

    public static Response neConfigCreate(String token, Map<String, Object> fields) throws Exception {
        return postAuth(token, "/aa/authorize/ne/config/create", toJson(fields));
    }

    public static Response neConfigList(String token) throws Exception {
        return getAuth(token, "/aa/authorize/ne/config/list");
    }

    public static Response neConfigUpdate(String token, Map<String, Object> fields) throws Exception {
        return postAuth(token, "/aa/authorize/ne/config/update", toJson(fields));
    }

    public static Response neConfigDelete(String token, long id) throws Exception {
        return postAuth(token, "/aa/authorize/ne/config/delete", toJson(map("id", id)));
    }

    // ── Config Backup ────────────────────────────────────────────────────────

    public static Response configBackupSave(String token, String neName, String neIp, String configXml) throws Exception {
        return postAuth(token, "/aa/config-backup/save", toJson(map(
                "ne_name", neName,
                "ne_ip", neIp,
                "config_xml", configXml)));
    }

    /** GET /aa/config-backup/list?ne_name=&lt;optional&gt; — pass null for no filter. */
    public static Response configBackupList(String token, String neNameFilter) throws Exception {
        String q = (neNameFilter == null || neNameFilter.isEmpty())
                ? "" : "?ne_name=" + urlEncode(neNameFilter);
        return getAuth(token, "/aa/config-backup/list" + q);
    }

    public static Response configBackupGet(String token, long id) throws Exception {
        return getAuth(token, "/aa/config-backup/" + id);
    }

    // ── History ──────────────────────────────────────────────────────────────

    /**
     * GET /aa/history/list — all filter params optional.
     * @param limit   1..500 (default 100). 0 to omit.
     * @param scope   "cli-config" | "ne-command" | "ne-config" or null.
     * @param neName  filter by NE name, or null.
     * @param account filter by username, or null.
     */
    public static Response historyList(String token, int limit, String scope, String neName, String account) throws Exception {
        StringBuilder q = new StringBuilder();
        if (limit > 0)                                    appendQuery(q, "limit",   String.valueOf(limit));
        if (scope   != null && !scope.isEmpty())          appendQuery(q, "scope",   scope);
        if (neName  != null && !neName.isEmpty())         appendQuery(q, "ne_name", neName);
        if (account != null && !account.isEmpty())        appendQuery(q, "account", account);
        return getAuth(token, "/aa/history/list" + q);
    }

    /**
     * POST /aa/history/save — used by SSH / cli-netconf to log a CLI command.
     * account is taken from the JWT — do not send it.
     */
    public static Response historySave(String token, String cmdName, String neName, String neIp,
                                       String scope, String result) throws Exception {
        return postAuth(token, "/aa/history/save", toJson(map(
                "cmd_name", cmdName,
                "ne_name",  neName,
                "ne_ip",    neIp,
                "scope",    scope,
                "result",   result)));
    }

    // ── Admin (frontend API) ─────────────────────────────────────────────────

    public static Response adminUserList(String token) throws Exception {
        return getAuth(token, "/aa/admin/user/list");
    }

    public static Response adminUserUpdate(String token, Map<String, Object> fields) throws Exception {
        return postAuth(token, "/aa/admin/user/update", toJson(fields));
    }

    public static Response adminNeList(String token) throws Exception {
        return getAuth(token, "/aa/admin/ne/list");
    }

    public static Response adminNeCreate(String token, Map<String, Object> fields) throws Exception {
        return postAuth(token, "/aa/admin/ne/create", toJson(fields));
    }

    public static Response adminNeUpdate(String token, Map<String, Object> fields) throws Exception {
        return postAuth(token, "/aa/admin/ne/update", toJson(fields));
    }

    // ── Groups ───────────────────────────────────────────────────────────────

    public static Response groupList(String token) throws Exception {
        return getAuth(token, "/aa/group/list");
    }

    public static Response groupShow(String token, long id) throws Exception {
        return postAuth(token, "/aa/group/show", toJson(Map.of("id", id)));
    }

    public static Response groupCreate(String token, String name, String description) throws Exception {
        return postAuth(token, "/aa/group/create", toJson(Map.of("name", name, "description", description == null ? "" : description)));
    }

    public static Response groupUpdate(String token, long id, String name, String description) throws Exception {
        return postAuth(token, "/aa/group/update", toJson(Map.of("id", id, "name", name == null ? "" : name, "description", description == null ? "" : description)));
    }

    public static Response groupDelete(String token, long id) throws Exception {
        return postAuth(token, "/aa/group/delete", toJson(Map.of("id", id)));
    }

    public static Response userGroupList(String token, String username) throws Exception {
        return getAuth(token, "/aa/group/user?username=" + URLEncoder.encode(username, StandardCharsets.UTF_8));
    }

    public static Response userGroupAssign(String token, String username, long groupId) throws Exception {
        return postAuth(token, "/aa/group/user/assign", toJson(Map.of("username", username, "group_id", groupId)));
    }

    public static Response userGroupUnassign(String token, String username, long groupId) throws Exception {
        return postAuth(token, "/aa/group/user/unassign", toJson(Map.of("username", username, "group_id", groupId)));
    }

    public static Response groupNeList(String token, long groupId) throws Exception {
        return getAuth(token, "/aa/group/ne?group_id=" + groupId);
    }

    public static Response groupNeAssign(String token, long groupId, long neId) throws Exception {
        return postAuth(token, "/aa/group/ne/assign", toJson(Map.of("group_id", groupId, "ne_id", neId)));
    }

    public static Response groupNeUnassign(String token, long groupId, long neId) throws Exception {
        return postAuth(token, "/aa/group/ne/unassign", toJson(Map.of("group_id", groupId, "ne_id", neId)));
    }

    // ── Import ───────────────────────────────────────────────────────────────

    /** POST /aa/import/ — body is plain text in section format ([users], [nes], ...). */
    public static Response importBulk(String token, String plainTextBody) throws Exception {
        requireBaseUrl();
        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + "/aa/import/"))
                .timeout(Duration.ofSeconds(15))
                .header("Content-Type", "text/plain; charset=utf-8")
                .header("Authorization", token)
                .POST(BodyPublishers.ofString(plainTextBody, StandardCharsets.UTF_8))
                .build();
        return send(req);
    }

    // ── Subscribers ──────────────────────────────────────────────────────────

    public static Response subscribersFiles(String token) throws Exception {
        return getAuth(token, "/aa/subscribers/files");
    }

    public static Response subscribersFile(String token, int index) throws Exception {
        return getAuth(token, "/aa/subscribers/files/" + index);
    }

    // ── Typed convenience wrappers ───────────────────────────────────────────
    //
    // These call the matching raw endpoint above, then parse the body into a
    // record from MgtModels. They throw MgtApiException if status is not 2xx —
    // use the raw Response-returning variants if you want to inspect failures.

    public static MgtModels.AuthResult authenticateTyped(String u, String p) throws Exception {
        return must(authenticate(u, p)).as(MgtModels.AuthResult::from);
    }

    public static MgtModels.ValidateTokenResult validateTokenTyped(String tokenToValidate) throws Exception {
        return must(validateToken(tokenToValidate)).as(MgtModels.ValidateTokenResult::from);
    }

    public static List<MgtModels.UserShow> showUsersTyped(String token) throws Exception {
        return must(showUsers(token)).asList(MgtModels.UserShow::from);
    }

    public static List<MgtModels.UserPermission> authorizeUserShowTyped(String token) throws Exception {
        return must(authorizeUserShow(token)).asList(MgtModels.UserPermission::from);
    }

    public static List<MgtModels.NeShow> neShowTyped(String token) throws Exception {
        return must(neShow(token)).asList(MgtModels.NeShow::from);
    }

    public static List<MgtModels.Ne> listNeTyped(String token) throws Exception {
        return must(listNe(token)).asList(MgtModels.Ne::from);
    }

    public static List<MgtModels.NeMonitor> listNeMonitorTyped(String token) throws Exception {
        return must(listNeMonitor(token)).asList(MgtModels.NeMonitor::from);
    }

    public static List<MgtModels.NeConfig> neConfigListTyped(String token) throws Exception {
        return must(neConfigList(token)).asList(MgtModels.NeConfig::from);
    }

    /** Parsed {@link #configBackupList}. Returns only the {@code backups} array. */
    public static List<MgtModels.ConfigBackup> configBackupListTyped(String token, String neNameFilter) throws Exception {
        MgtModels.ConfigBackupListResult wrapped =
                must(configBackupList(token, neNameFilter)).as(MgtModels.ConfigBackupListResult::from);
        return wrapped == null || wrapped.backups() == null ? List.of() : wrapped.backups();
    }

    public static MgtModels.ConfigBackupDetail configBackupGetTyped(String token, long id) throws Exception {
        return must(configBackupGet(token, id)).as(MgtModels.ConfigBackupDetail::from);
    }

    public static MgtModels.ConfigBackupSaveResult configBackupSaveTyped(String token, String neName, String neIp, String xml) throws Exception {
        return must(configBackupSave(token, neName, neIp, xml)).as(MgtModels.ConfigBackupSaveResult::from);
    }

    public static List<MgtModels.History> historyListTyped(String token, int limit, String scope, String neName, String account) throws Exception {
        return must(historyList(token, limit, scope, neName, account)).asList(MgtModels.History::from);
    }

    public static List<MgtModels.AdminUser> adminUserListTyped(String token) throws Exception {
        return must(adminUserList(token)).asList(MgtModels.AdminUser::from);
    }

    public static List<MgtModels.Group> groupListTyped(String token) throws Exception {
        return must(groupList(token)).asList(MgtModels.Group::from);
    }

    public static MgtModels.GroupDetail groupShowTyped(String token, long id) throws Exception {
        return must(groupShow(token, id)).as(MgtModels.GroupDetail::from);
    }

    public static List<MgtModels.Group> userGroupListTyped(String token, String username) throws Exception {
        return must(userGroupList(token, username)).asList(MgtModels.Group::from);
    }

    public static List<Long> groupNeListTyped(String token, long groupId) throws Exception {
        return must(groupNeList(token, groupId)).asLongList();
    }

    public static List<MgtModels.CliNe> adminNeListTyped(String token) throws Exception {
        return must(adminNeList(token)).asList(MgtModels.CliNe::from);
    }

    public static List<MgtModels.ImportResult> importBulkTyped(String token, String body) throws Exception {
        return must(importBulk(token, body)).asList(MgtModels.ImportResult::from);
    }

    public static List<MgtModels.SubscriberFile> subscribersFilesTyped(String token) throws Exception {
        return must(subscribersFiles(token)).asList(MgtModels.SubscriberFile::from);
    }

    public static MgtModels.SubscriberFileContent subscribersFileTyped(String token, int index) throws Exception {
        return must(subscribersFile(token, index)).as(MgtModels.SubscriberFileContent::from);
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

    /** GET {path} with Authorization header. */
    public static Response getAuth(String token, String path) throws Exception {
        return send(authed(token, path).GET().build());
    }

    /** POST {path} (JSON body) with Authorization header. */
    public static Response postAuth(String token, String path, String jsonBody) throws Exception {
        HttpRequest req = authed(token, path)
                .header("Content-Type", "application/json")
                .POST(BodyPublishers.ofString(jsonBody, StandardCharsets.UTF_8))
                .build();
        return send(req);
    }

    /** POST {path} (JSON body) without Authorization header — for /aa/authenticate and /aa/validate-token. */
    public static Response postUnauth(String path, String jsonBody) throws Exception {
        HttpRequest req = unauth(path)
                .header("Content-Type", "application/json")
                .POST(BodyPublishers.ofString(jsonBody, StandardCharsets.UTF_8))
                .build();
        return send(req);
    }

    private static HttpRequest.Builder authed(String token, String path) {
        if (token == null || token.isEmpty()) throw new IllegalArgumentException("token required");
        return unauth(path).header("Authorization", token);
    }

    private static HttpRequest.Builder unauth(String path) {
        requireBaseUrl();
        return HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + path))
                .timeout(Duration.ofSeconds(15));
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
