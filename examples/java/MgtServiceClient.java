// MgtServiceClient — pure-JDK (no Gson/Jackson) client for cli-mgt v2.
//
// Drop into any JDK 11+ project. The JSON parser is a tiny hand-written one
// good enough for the v2 API shape (flat maps / arrays of primitives). If
// you already use Jackson, feel free to swap toJson/parseJson for
// ObjectMapper — the public API signatures stay the same.

package examples.java;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;

public class MgtServiceClient {
    private final String baseURL;
    private final HttpClient http;
    private String token;

    public MgtServiceClient(String baseURL) {
        this.baseURL = baseURL.endsWith("/") ? baseURL.substring(0, baseURL.length() - 1) : baseURL;
        this.http = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(10)).build();
    }

    public String getToken() { return token; }
    public void setToken(String t) { this.token = t; }

    // ── Auth ────────────────────────────────────────────────────────────

    /** Login. On success the token is stored on this instance. */
    public void authenticate(String username, String password) throws IOException, InterruptedException {
        Map<String, Object> body = new HashMap<>();
        body.put("username", username);
        body.put("password", password);
        Map<String, Object> res = postJson("/aa/authenticate", body);
        Object tok = res.get("token");
        if (!(tok instanceof String)) throw new IOException("authenticate: no token in response: " + res);
        this.token = (String) tok;
    }

    public Map<String, Object> validateToken(String tok) throws IOException, InterruptedException {
        Map<String, Object> body = new HashMap<>(); body.put("token", tok);
        return postJson("/aa/validate-token", body);
    }

    public void changePassword(String oldPw, String newPw) throws IOException, InterruptedException {
        Map<String, Object> body = new HashMap<>();
        body.put("old_password", oldPw);
        body.put("new_password", newPw);
        postJson("/aa/change-password", body);
    }

    // ── Users ───────────────────────────────────────────────────────────

    public List<Map<String, Object>> listUsers() throws IOException, InterruptedException {
        return asList(getJson("/aa/users"));
    }

    public Map<String, Object> createUser(String username, String password, String email, String fullName) throws IOException, InterruptedException {
        Map<String, Object> body = new HashMap<>();
        body.put("username", username);
        body.put("password", password);
        body.put("email", email);
        body.put("full_name", fullName);
        return postJson("/aa/users", body);
    }

    public Map<String, Object> getUser(long id) throws IOException, InterruptedException {
        return asMap(getJson("/aa/users/" + id));
    }

    public void updateUser(long id, Map<String, Object> patch) throws IOException, InterruptedException {
        request("PUT", "/aa/users/" + id, patch);
    }

    public void deleteUser(long id) throws IOException, InterruptedException {
        request("DELETE", "/aa/users/" + id, null);
    }

    public void resetPassword(long id, String newPassword) throws IOException, InterruptedException {
        Map<String, Object> body = new HashMap<>(); body.put("new_password", newPassword);
        postJson("/aa/users/" + id + "/reset-password", body);
    }

    // ── NEs ─────────────────────────────────────────────────────────────

    public List<Map<String, Object>> listNEs() throws IOException, InterruptedException {
        return asList(getJson("/aa/nes"));
    }

    public Map<String, Object> createNE(Map<String, Object> ne) throws IOException, InterruptedException {
        return postJson("/aa/nes", ne);
    }

    public void updateNE(long id, Map<String, Object> ne) throws IOException, InterruptedException {
        request("PUT", "/aa/nes/" + id, ne);
    }

    public void deleteNE(long id) throws IOException, InterruptedException {
        request("DELETE", "/aa/nes/" + id, null);
    }

    // ── Commands ────────────────────────────────────────────────────────

    public List<Map<String, Object>> listCommands() throws IOException, InterruptedException {
        return asList(getJson("/aa/commands"));
    }

    public List<Map<String, Object>> listCommandsFor(long neId, String service) throws IOException, InterruptedException {
        StringBuilder q = new StringBuilder("/aa/commands?");
        if (neId > 0) q.append("ne_id=").append(neId).append("&");
        if (service != null && !service.isEmpty()) q.append("service=").append(service);
        return asList(getJson(q.toString()));
    }

    public Map<String, Object> createCommand(long neId, String service, String cmdText, String description) throws IOException, InterruptedException {
        Map<String, Object> body = new HashMap<>();
        body.put("ne_id", neId);
        body.put("service", service);
        body.put("cmd_text", cmdText);
        body.put("description", description);
        return postJson("/aa/commands", body);
    }

    public void deleteCommand(long id) throws IOException, InterruptedException {
        request("DELETE", "/aa/commands/" + id, null);
    }

    // ── Groups ──────────────────────────────────────────────────────────

    public List<Map<String, Object>> listNeAccessGroups() throws IOException, InterruptedException {
        return asList(getJson("/aa/ne-access-groups"));
    }

    public Map<String, Object> createNeAccessGroup(String name, String description) throws IOException, InterruptedException {
        return postJson("/aa/ne-access-groups", group(name, description));
    }

    public void addUserToNeAccessGroup(long groupId, long userId) throws IOException, InterruptedException {
        Map<String, Object> b = new HashMap<>(); b.put("user_id", userId);
        postJson("/aa/ne-access-groups/" + groupId + "/users", b);
    }

    public void addNeToNeAccessGroup(long groupId, long neId) throws IOException, InterruptedException {
        Map<String, Object> b = new HashMap<>(); b.put("ne_id", neId);
        postJson("/aa/ne-access-groups/" + groupId + "/nes", b);
    }

    public List<Map<String, Object>> listCmdExecGroups() throws IOException, InterruptedException {
        return asList(getJson("/aa/cmd-exec-groups"));
    }

    public Map<String, Object> createCmdExecGroup(String name, String description) throws IOException, InterruptedException {
        return postJson("/aa/cmd-exec-groups", group(name, description));
    }

    public void addUserToCmdExecGroup(long groupId, long userId) throws IOException, InterruptedException {
        Map<String, Object> b = new HashMap<>(); b.put("user_id", userId);
        postJson("/aa/cmd-exec-groups/" + groupId + "/users", b);
    }

    public void addCommandToCmdExecGroup(long groupId, long commandId) throws IOException, InterruptedException {
        Map<String, Object> b = new HashMap<>(); b.put("command_id", commandId);
        postJson("/aa/cmd-exec-groups/" + groupId + "/commands", b);
    }

    private Map<String, Object> group(String name, String description) {
        Map<String, Object> g = new HashMap<>();
        g.put("name", name);
        g.put("description", description);
        return g;
    }

    // ── Policy + Access List + Authorize ────────────────────────────────

    public Map<String, Object> getPasswordPolicy() throws IOException, InterruptedException {
        return asMap(getJson("/aa/password-policy"));
    }

    public Map<String, Object> upsertPasswordPolicy(Map<String, Object> policy) throws IOException, InterruptedException {
        return asMap(request("PUT", "/aa/password-policy", policy));
    }

    public List<Map<String, Object>> listAccessList(String listType) throws IOException, InterruptedException {
        String path = "/aa/access-list" + (listType != null && !listType.isEmpty() ? ("?list_type=" + listType) : "");
        return asList(getJson(path));
    }

    /** "Can user X execute command Y on NE Z?" */
    public AuthorizeDecision authorizeCheck(String username, long neId, long commandId) throws IOException, InterruptedException {
        Map<String, Object> body = new HashMap<>();
        body.put("username", username);
        body.put("ne_id", neId);
        body.put("command_id", commandId);
        Map<String, Object> res = postJson("/aa/authorize/check", body);
        AuthorizeDecision d = new AuthorizeDecision();
        d.allowed            = Boolean.TRUE.equals(res.get("allowed"));
        d.reason             = (String) res.getOrDefault("reason", "");
        d.userExists         = Boolean.TRUE.equals(res.get("user_exists"));
        d.userEnabled        = Boolean.TRUE.equals(res.get("user_enabled"));
        d.neReachable        = Boolean.TRUE.equals(res.get("ne_reachable"));
        d.commandOnNe        = Boolean.TRUE.equals(res.get("command_on_ne"));
        d.commandExecAllowed = Boolean.TRUE.equals(res.get("command_exec_allowed"));
        return d;
    }

    public static class AuthorizeDecision {
        public boolean allowed;
        public String reason;
        public boolean userExists, userEnabled, neReachable, commandOnNe, commandExecAllowed;
    }

    // ── History + Config Backup ─────────────────────────────────────────

    public List<Map<String, Object>> listHistory(int limit, String scope, String neNamespace, String account) throws IOException, InterruptedException {
        StringBuilder q = new StringBuilder("/aa/history?");
        if (limit > 0) q.append("limit=").append(limit).append("&");
        if (scope != null && !scope.isEmpty()) q.append("scope=").append(scope).append("&");
        if (neNamespace != null && !neNamespace.isEmpty()) q.append("ne_namespace=").append(neNamespace).append("&");
        if (account != null && !account.isEmpty()) q.append("account=").append(account);
        return asList(getJson(q.toString()));
    }

    /** /history/save is deliberately unauthenticated — used by the cli-gate proxy. */
    public void saveHistory(String account, String cmdText, String neNamespace, String neIp, String scope, String result) throws IOException, InterruptedException {
        Map<String, Object> body = new HashMap<>();
        body.put("account", account);
        body.put("cmd_text", cmdText);
        body.put("ne_namespace", neNamespace);
        body.put("ne_ip", neIp);
        body.put("scope", scope);
        body.put("result", result);
        postJson("/aa/history/save", body);
    }

    public Map<String, Object> saveConfigBackup(String neName, String neIp, String configXML) throws IOException, InterruptedException {
        Map<String, Object> body = new HashMap<>();
        body.put("ne_name", neName);
        body.put("ne_ip", neIp);
        body.put("config_xml", configXML);
        return postJson("/aa/config-backup/save", body);
    }

    public Map<String, Object> listConfigBackups(String neName) throws IOException, InterruptedException {
        String path = "/aa/config-backup/list" + (neName != null && !neName.isEmpty() ? ("?ne_name=" + neName) : "");
        return asMap(getJson(path));
    }

    public Map<String, Object> getConfigBackup(long id) throws IOException, InterruptedException {
        return asMap(getJson("/aa/config-backup/" + id));
    }

    // ── HTTP / JSON plumbing ────────────────────────────────────────────

    private Object getJson(String path) throws IOException, InterruptedException {
        return request("GET", path, null);
    }

    private Map<String, Object> postJson(String path, Map<String, Object> body) throws IOException, InterruptedException {
        return asMap(request("POST", path, body));
    }

    private Object request(String method, String path, Object body) throws IOException, InterruptedException {
        HttpRequest.Builder b = HttpRequest.newBuilder()
                .uri(URI.create(baseURL + path))
                .timeout(Duration.ofSeconds(15))
                .header("Content-Type", "application/json");
        if (token != null && !token.isEmpty()) b.header("Authorization", token);
        switch (method) {
            case "GET":    b.GET(); break;
            case "DELETE": b.DELETE(); break;
            default:       b.method(method, HttpRequest.BodyPublishers.ofString(body == null ? "" : toJson(body)));
        }
        HttpResponse<String> res = http.send(b.build(), HttpResponse.BodyHandlers.ofString());
        if (res.statusCode() < 200 || res.statusCode() >= 300) {
            throw new IOException(method + " " + path + " → HTTP " + res.statusCode() + " " + res.body());
        }
        if (res.body() == null || res.body().isEmpty()) return null;
        return parseJson(res.body());
    }

    @SuppressWarnings("unchecked")
    private static Map<String, Object> asMap(Object v) {
        return v instanceof Map ? (Map<String, Object>) v : new HashMap<>();
    }
    @SuppressWarnings("unchecked")
    private static List<Map<String, Object>> asList(Object v) {
        if (!(v instanceof List)) return new ArrayList<>();
        List<?> raw = (List<?>) v;
        List<Map<String, Object>> out = new ArrayList<>(raw.size());
        for (Object o : raw) if (o instanceof Map) out.add((Map<String, Object>) o);
        return out;
    }

    // ── Minimal JSON encoder / decoder ──────────────────────────────────

    private static String toJson(Object o) {
        StringBuilder sb = new StringBuilder();
        writeJson(sb, o);
        return sb.toString();
    }

    private static void writeJson(StringBuilder sb, Object o) {
        if (o == null) { sb.append("null"); return; }
        if (o instanceof Boolean) { sb.append((boolean) o); return; }
        if (o instanceof Number)  { sb.append(o); return; }
        if (o instanceof Map) {
            sb.append('{');
            boolean first = true;
            for (Map.Entry<?, ?> e : ((Map<?, ?>) o).entrySet()) {
                if (!first) sb.append(',');
                first = false;
                writeString(sb, String.valueOf(e.getKey()));
                sb.append(':');
                writeJson(sb, e.getValue());
            }
            sb.append('}');
            return;
        }
        if (o instanceof Iterable) {
            sb.append('[');
            boolean first = true;
            for (Object v : (Iterable<?>) o) {
                if (!first) sb.append(',');
                first = false;
                writeJson(sb, v);
            }
            sb.append(']');
            return;
        }
        writeString(sb, String.valueOf(o));
    }

    private static void writeString(StringBuilder sb, String s) {
        sb.append('"');
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            switch (c) {
                case '"':  sb.append("\\\""); break;
                case '\\': sb.append("\\\\"); break;
                case '\n': sb.append("\\n"); break;
                case '\r': sb.append("\\r"); break;
                case '\t': sb.append("\\t"); break;
                default:
                    if (c < 0x20) sb.append(String.format(Locale.ROOT, "\\u%04x", (int) c));
                    else sb.append(c);
            }
        }
        sb.append('"');
    }

    private static Object parseJson(String s) throws IOException {
        Parser p = new Parser(s);
        p.skipWS();
        Object v = p.readValue();
        p.skipWS();
        if (p.pos != s.length()) throw new IOException("trailing data at pos " + p.pos);
        return v;
    }

    private static final class Parser {
        final String s; int pos;
        Parser(String s) { this.s = s; }
        void skipWS() { while (pos < s.length() && Character.isWhitespace(s.charAt(pos))) pos++; }
        Object readValue() throws IOException {
            skipWS();
            if (pos >= s.length()) throw new IOException("unexpected EOF");
            char c = s.charAt(pos);
            if (c == '{') return readObject();
            if (c == '[') return readArray();
            if (c == '"') return readString();
            if (c == 't' || c == 'f') return readBool();
            if (c == 'n') { expect("null"); return null; }
            return readNumber();
        }
        Map<String, Object> readObject() throws IOException {
            expect('{');
            Map<String, Object> m = new HashMap<>();
            skipWS();
            if (peek() == '}') { pos++; return m; }
            while (true) {
                skipWS();
                String key = readString();
                skipWS();
                expect(':');
                Object v = readValue();
                m.put(key, v);
                skipWS();
                if (peek() == ',') { pos++; continue; }
                expect('}');
                return m;
            }
        }
        List<Object> readArray() throws IOException {
            expect('[');
            List<Object> list = new ArrayList<>();
            skipWS();
            if (peek() == ']') { pos++; return list; }
            while (true) {
                list.add(readValue());
                skipWS();
                if (peek() == ',') { pos++; continue; }
                expect(']');
                return list;
            }
        }
        String readString() throws IOException {
            expect('"');
            StringBuilder sb = new StringBuilder();
            while (pos < s.length()) {
                char c = s.charAt(pos++);
                if (c == '"') return sb.toString();
                if (c != '\\') { sb.append(c); continue; }
                char esc = s.charAt(pos++);
                switch (esc) {
                    case '"':  sb.append('"');  break;
                    case '\\': sb.append('\\'); break;
                    case '/':  sb.append('/');  break;
                    case 'n':  sb.append('\n'); break;
                    case 'r':  sb.append('\r'); break;
                    case 't':  sb.append('\t'); break;
                    case 'b':  sb.append('\b'); break;
                    case 'f':  sb.append('\f'); break;
                    case 'u':
                        sb.append((char) Integer.parseInt(s.substring(pos, pos + 4), 16));
                        pos += 4;
                        break;
                    default: throw new IOException("bad escape \\" + esc);
                }
            }
            throw new IOException("unterminated string");
        }
        Boolean readBool() throws IOException {
            if (peek() == 't') { expect("true"); return true; }
            expect("false"); return false;
        }
        Object readNumber() {
            int start = pos;
            if (peek() == '-') pos++;
            while (pos < s.length() && "0123456789.eE+-".indexOf(s.charAt(pos)) >= 0) pos++;
            String num = s.substring(start, pos);
            if (num.indexOf('.') >= 0 || num.indexOf('e') >= 0 || num.indexOf('E') >= 0) return Double.parseDouble(num);
            try { return Long.parseLong(num); } catch (NumberFormatException e) { return Double.parseDouble(num); }
        }
        char peek() { return pos < s.length() ? s.charAt(pos) : '\0'; }
        void expect(char c) throws IOException {
            if (pos >= s.length() || s.charAt(pos) != c) throw new IOException("expected '" + c + "' at pos " + pos);
            pos++;
        }
        void expect(String t) throws IOException {
            if (!s.startsWith(t, pos)) throw new IOException("expected \"" + t + "\" at pos " + pos);
            pos += t.length();
        }
    }

    // ── Demo main ───────────────────────────────────────────────────────

    public static void main(String[] args) throws Exception {
        String base = System.getenv().getOrDefault("MGT_BASE", "http://localhost:3000");
        String user = System.getenv().getOrDefault("MGT_USER", "admin");
        String pass = System.getenv().getOrDefault("MGT_PASS", "admin");

        MgtServiceClient c = new MgtServiceClient(base);
        c.authenticate(user, pass);
        System.out.println("logged in");

        List<Map<String, Object>> users = c.listUsers();
        System.out.println("users: " + users.size());
        for (Map<String, Object> u : users) {
            System.out.println("  " + u.get("id") + " " + u.get("username") + " enabled=" + u.get("is_enabled"));
        }

        List<Map<String, Object>> nes = c.listNEs();
        System.out.println("NEs: " + nes.size());

        List<Map<String, Object>> cmds = c.listCommands();
        System.out.println("commands: " + cmds.size());

        if (!users.isEmpty() && !nes.isEmpty() && !cmds.isEmpty()) {
            Map<String, Object> u = users.get(0);
            Map<String, Object> n = nes.get(0);
            Map<String, Object> cmd = cmds.get(0);
            AuthorizeDecision d = c.authorizeCheck(
                    (String) u.get("username"),
                    ((Number) n.get("id")).longValue(),
                    ((Number) cmd.get("id")).longValue());
            System.out.println("authorize(" + u.get("username") + ", ne=" + n.get("namespace") + ", cmd=#" + cmd.get("id") + ") → "
                    + (d.allowed ? "ALLOW" : "DENY: " + d.reason));
        }

        System.out.println("done. Supported endpoints demonstrated:\n  " +
                String.join("\n  ", Arrays.asList(
                        "POST /aa/authenticate", "GET  /aa/users",
                        "GET  /aa/nes", "GET  /aa/commands",
                        "POST /aa/authorize/check")));
    }
}
