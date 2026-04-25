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
    private String role;

    public MgtServiceClient(String baseURL) {
        this.baseURL = baseURL.endsWith("/") ? baseURL.substring(0, baseURL.length() - 1) : baseURL;
        this.http = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(10)).build();
    }

    public String getToken() { return token; }
    public void setToken(String t) { this.token = t; }
    public String getRole() { return role; }

    // ── Map extraction helpers ─────────────────────────────────────────

    private static String str(Map<String, Object> m, String k) { Object v = m.get(k); return v == null ? "" : v.toString(); }
    private static long lng(Map<String, Object> m, String k) { Object v = m.get(k); return v instanceof Number ? ((Number) v).longValue() : 0; }
    private static int num(Map<String, Object> m, String k) { Object v = m.get(k); return v instanceof Number ? ((Number) v).intValue() : 0; }
    private static boolean bool(Map<String, Object> m, String k) { return Boolean.TRUE.equals(m.get(k)); }

    // ── Typed model classes ────────────────────────────────────────────

    public static class User {
        public long id;
        public String username, email, fullName, phone, role;
        public boolean isEnabled;
        public String passwordExpiresAt;
        public int loginFailureCount;
        public String lockedAt, lastLoginAt, createdAt, updatedAt;

        public static User fromMap(Map<String, Object> m) {
            User u = new User();
            u.id = lng(m, "id");
            u.username = str(m, "username");
            u.email = str(m, "email");
            u.fullName = str(m, "full_name");
            u.phone = str(m, "phone");
            u.role = str(m, "role");
            u.isEnabled = bool(m, "is_enabled");
            u.passwordExpiresAt = str(m, "password_expires_at");
            u.loginFailureCount = num(m, "login_failure_count");
            u.lockedAt = str(m, "locked_at");
            u.lastLoginAt = str(m, "last_login_at");
            u.createdAt = str(m, "created_at");
            u.updatedAt = str(m, "updated_at");
            return u;
        }

        public Map<String, Object> toMap() {
            Map<String, Object> m = new HashMap<>();
            m.put("id", id);
            m.put("username", username);
            m.put("email", email);
            m.put("full_name", fullName);
            m.put("phone", phone);
            m.put("role", role);
            m.put("is_enabled", isEnabled);
            m.put("password_expires_at", passwordExpiresAt);
            m.put("login_failure_count", loginFailureCount);
            m.put("locked_at", lockedAt);
            m.put("last_login_at", lastLoginAt);
            m.put("created_at", createdAt);
            m.put("updated_at", updatedAt);
            return m;
        }

        @Override public String toString() {
            return "User{id=" + id + ", username=" + username + ", email=" + email
                    + ", role=" + role + ", isEnabled=" + isEnabled + "}";
        }
    }

    public static class NE {
        public long id;
        public String namespace, neType, siteName, description;
        public String masterIp;
        public int masterPort;
        public String sshUsername, sshPassword, commandUrl, confMode;

        public static NE fromMap(Map<String, Object> m) {
            NE n = new NE();
            n.id = lng(m, "id");
            n.namespace = str(m, "namespace");
            n.neType = str(m, "ne_type");
            n.siteName = str(m, "site_name");
            n.description = str(m, "description");
            n.masterIp = str(m, "master_ip");
            n.masterPort = num(m, "master_port");
            n.sshUsername = str(m, "ssh_username");
            n.sshPassword = str(m, "ssh_password");
            n.commandUrl = str(m, "command_url");
            n.confMode = str(m, "conf_mode");
            return n;
        }

        public Map<String, Object> toMap() {
            Map<String, Object> m = new HashMap<>();
            m.put("id", id);
            m.put("namespace", namespace);
            m.put("ne_type", neType);
            m.put("site_name", siteName);
            m.put("description", description);
            m.put("master_ip", masterIp);
            m.put("master_port", masterPort);
            m.put("ssh_username", sshUsername);
            m.put("ssh_password", sshPassword);
            m.put("command_url", commandUrl);
            m.put("conf_mode", confMode);
            return m;
        }

        @Override public String toString() {
            return "NE{id=" + id + ", namespace=" + namespace + ", neType=" + neType
                    + ", masterIp=" + masterIp + ":" + masterPort + "}";
        }
    }

    public static class Command {
        public long id, neId;
        public String service, cmdText, description;

        public static Command fromMap(Map<String, Object> m) {
            Command c = new Command();
            c.id = lng(m, "id");
            c.neId = lng(m, "ne_id");
            c.service = str(m, "service");
            c.cmdText = str(m, "cmd_text");
            c.description = str(m, "description");
            return c;
        }

        public Map<String, Object> toMap() {
            Map<String, Object> m = new HashMap<>();
            m.put("id", id);
            m.put("ne_id", neId);
            m.put("service", service);
            m.put("cmd_text", cmdText);
            m.put("description", description);
            return m;
        }

        @Override public String toString() {
            return "Command{id=" + id + ", neId=" + neId + ", service=" + service
                    + ", cmdText=" + cmdText + "}";
        }
    }

    public static class Group {
        public long id;
        public String name, description;

        public static Group fromMap(Map<String, Object> m) {
            Group g = new Group();
            g.id = lng(m, "id");
            g.name = str(m, "name");
            g.description = str(m, "description");
            return g;
        }

        public Map<String, Object> toMap() {
            Map<String, Object> m = new HashMap<>();
            m.put("id", id);
            m.put("name", name);
            m.put("description", description);
            return m;
        }

        @Override public String toString() {
            return "Group{id=" + id + ", name=" + name + "}";
        }
    }

    public static class AccessListEntry {
        public long id;
        public String listType, matchType, pattern, reason, createdAt;

        public static AccessListEntry fromMap(Map<String, Object> m) {
            AccessListEntry e = new AccessListEntry();
            e.id = lng(m, "id");
            e.listType = str(m, "list_type");
            e.matchType = str(m, "match_type");
            e.pattern = str(m, "pattern");
            e.reason = str(m, "reason");
            e.createdAt = str(m, "created_at");
            return e;
        }

        public Map<String, Object> toMap() {
            Map<String, Object> m = new HashMap<>();
            m.put("id", id);
            m.put("list_type", listType);
            m.put("match_type", matchType);
            m.put("pattern", pattern);
            m.put("reason", reason);
            m.put("created_at", createdAt);
            return m;
        }

        @Override public String toString() {
            return "AccessListEntry{id=" + id + ", listType=" + listType
                    + ", pattern=" + pattern + "}";
        }
    }

    public static class HistoryEntry {
        public long id;
        public String account, cmdText, neNamespace, neIp, scope, result, createdDate;

        public static HistoryEntry fromMap(Map<String, Object> m) {
            HistoryEntry h = new HistoryEntry();
            h.id = lng(m, "id");
            h.account = str(m, "account");
            h.cmdText = str(m, "cmd_text");
            h.neNamespace = str(m, "ne_namespace");
            h.neIp = str(m, "ne_ip");
            h.scope = str(m, "scope");
            h.result = str(m, "result");
            h.createdDate = str(m, "created_date");
            return h;
        }

        public Map<String, Object> toMap() {
            Map<String, Object> m = new HashMap<>();
            m.put("id", id);
            m.put("account", account);
            m.put("cmd_text", cmdText);
            m.put("ne_namespace", neNamespace);
            m.put("ne_ip", neIp);
            m.put("scope", scope);
            m.put("result", result);
            m.put("created_date", createdDate);
            return m;
        }

        @Override public String toString() {
            return "HistoryEntry{id=" + id + ", account=" + account
                    + ", cmdText=" + cmdText + ", scope=" + scope + "}";
        }
    }

    public static class PasswordPolicy {
        public int minLength, maxAgeDays, historyCount, maxLoginFailure, lockoutMinutes;
        public boolean requireUppercase, requireLowercase, requireDigit, requireSpecial;

        public static PasswordPolicy fromMap(Map<String, Object> m) {
            PasswordPolicy p = new PasswordPolicy();
            p.minLength = num(m, "min_length");
            p.maxAgeDays = num(m, "max_age_days");
            p.historyCount = num(m, "history_count");
            p.maxLoginFailure = num(m, "max_login_failure");
            p.lockoutMinutes = num(m, "lockout_minutes");
            p.requireUppercase = bool(m, "require_uppercase");
            p.requireLowercase = bool(m, "require_lowercase");
            p.requireDigit = bool(m, "require_digit");
            p.requireSpecial = bool(m, "require_special");
            return p;
        }

        public Map<String, Object> toMap() {
            Map<String, Object> m = new HashMap<>();
            m.put("min_length", minLength);
            m.put("max_age_days", maxAgeDays);
            m.put("history_count", historyCount);
            m.put("max_login_failure", maxLoginFailure);
            m.put("lockout_minutes", lockoutMinutes);
            m.put("require_uppercase", requireUppercase);
            m.put("require_lowercase", requireLowercase);
            m.put("require_digit", requireDigit);
            m.put("require_special", requireSpecial);
            return m;
        }

        @Override public String toString() {
            return "PasswordPolicy{minLength=" + minLength + ", maxAgeDays=" + maxAgeDays
                    + ", maxLoginFailure=" + maxLoginFailure + ", lockoutMinutes=" + lockoutMinutes + "}";
        }
    }

    public static class AuthorizeDecision {
        public boolean allowed;
        public String reason;
        public boolean userExists, userEnabled, neReachable, commandOnNe, commandExecAllowed;

        public static AuthorizeDecision fromMap(Map<String, Object> m) {
            AuthorizeDecision d = new AuthorizeDecision();
            d.allowed = bool(m, "allowed");
            d.reason = str(m, "reason");
            d.userExists = bool(m, "user_exists");
            d.userEnabled = bool(m, "user_enabled");
            d.neReachable = bool(m, "ne_reachable");
            d.commandOnNe = bool(m, "command_on_ne");
            d.commandExecAllowed = bool(m, "command_exec_allowed");
            return d;
        }

        public Map<String, Object> toMap() {
            Map<String, Object> m = new HashMap<>();
            m.put("allowed", allowed);
            m.put("reason", reason);
            m.put("user_exists", userExists);
            m.put("user_enabled", userEnabled);
            m.put("ne_reachable", neReachable);
            m.put("command_on_ne", commandOnNe);
            m.put("command_exec_allowed", commandExecAllowed);
            return m;
        }

        @Override public String toString() {
            return "AuthorizeDecision{allowed=" + allowed + ", reason=" + reason + "}";
        }
    }

    public static class ConfigBackup {
        public long id;
        public String neName, neIp, configXml, createdAt;

        public static ConfigBackup fromMap(Map<String, Object> m) {
            ConfigBackup b = new ConfigBackup();
            b.id = lng(m, "id");
            b.neName = str(m, "ne_name");
            b.neIp = str(m, "ne_ip");
            b.configXml = str(m, "config_xml");
            b.createdAt = str(m, "created_at");
            return b;
        }

        public Map<String, Object> toMap() {
            Map<String, Object> m = new HashMap<>();
            m.put("id", id);
            m.put("ne_name", neName);
            m.put("ne_ip", neIp);
            m.put("config_xml", configXml);
            m.put("created_at", createdAt);
            return m;
        }

        @Override public String toString() {
            return "ConfigBackup{id=" + id + ", neName=" + neName + ", neIp=" + neIp + "}";
        }
    }

    // ── Typed list conversion helpers ──────────────────────────────────

    private static List<User> toUsers(Object v) {
        List<User> out = new ArrayList<>();
        for (Map<String, Object> m : asList(v)) out.add(User.fromMap(m));
        return out;
    }

    private static List<NE> toNEs(Object v) {
        List<NE> out = new ArrayList<>();
        for (Map<String, Object> m : asList(v)) out.add(NE.fromMap(m));
        return out;
    }

    private static List<Command> toCommands(Object v) {
        List<Command> out = new ArrayList<>();
        for (Map<String, Object> m : asList(v)) out.add(Command.fromMap(m));
        return out;
    }

    private static List<Group> toGroups(Object v) {
        List<Group> out = new ArrayList<>();
        for (Map<String, Object> m : asList(v)) out.add(Group.fromMap(m));
        return out;
    }

    private static List<AccessListEntry> toAccessList(Object v) {
        List<AccessListEntry> out = new ArrayList<>();
        for (Map<String, Object> m : asList(v)) out.add(AccessListEntry.fromMap(m));
        return out;
    }

    private static List<HistoryEntry> toHistory(Object v) {
        List<HistoryEntry> out = new ArrayList<>();
        for (Map<String, Object> m : asList(v)) out.add(HistoryEntry.fromMap(m));
        return out;
    }

    private static List<ConfigBackup> toConfigBackups(Object v) {
        List<ConfigBackup> out = new ArrayList<>();
        for (Map<String, Object> m : asList(v)) out.add(ConfigBackup.fromMap(m));
        return out;
    }

    @SuppressWarnings("unchecked")
    private static List<Long> toLongList(Object v) {
        List<Long> out = new ArrayList<>();
        if (v instanceof List) {
            for (Object o : (List<?>) v) {
                if (o instanceof Number) out.add(((Number) o).longValue());
            }
        }
        return out;
    }

    // ── Auth ────────────────────────────────────────────────────────────

    /** Login. On success the token and role are stored on this instance. Returns the role. */
    public String authenticate(String username, String password) throws IOException, InterruptedException {
        Map<String, Object> body = new HashMap<>();
        body.put("username", username);
        body.put("password", password);
        Map<String, Object> res = postJson("/aa/authenticate", body);
        Object tok = res.get("token");
        if (!(tok instanceof String)) throw new IOException("authenticate: no token in response: " + res);
        this.token = (String) tok;
        this.role = str(res, "role");
        return this.role;
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

    public List<User> listUsers() throws IOException, InterruptedException {
        return toUsers(getJson("/aa/users"));
    }

    public User createUser(String username, String password, String email, String fullName, String role) throws IOException, InterruptedException {
        Map<String, Object> body = new HashMap<>();
        body.put("username", username);
        body.put("password", password);
        body.put("email", email);
        body.put("full_name", fullName);
        body.put("role", role);
        return User.fromMap(postJson("/aa/users", body));
    }

    public User getUser(long id) throws IOException, InterruptedException {
        return User.fromMap(asMap(getJson("/aa/users/" + id)));
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

    public List<NE> listNEs() throws IOException, InterruptedException {
        return toNEs(getJson("/aa/nes"));
    }

    public NE createNE(Map<String, Object> ne) throws IOException, InterruptedException {
        return NE.fromMap(postJson("/aa/nes", ne));
    }

    public void updateNE(long id, Map<String, Object> ne) throws IOException, InterruptedException {
        request("PUT", "/aa/nes/" + id, ne);
    }

    public void deleteNE(long id) throws IOException, InterruptedException {
        request("DELETE", "/aa/nes/" + id, null);
    }

    // ── Commands ────────────────────────────────────────────────────────

    public List<Command> listCommands() throws IOException, InterruptedException {
        return toCommands(getJson("/aa/commands"));
    }

    public List<Command> listCommandsFor(long neId, String service) throws IOException, InterruptedException {
        StringBuilder q = new StringBuilder("/aa/commands?");
        if (neId > 0) q.append("ne_id=").append(neId).append("&");
        if (service != null && !service.isEmpty()) q.append("service=").append(service);
        return toCommands(getJson(q.toString()));
    }

    public Command createCommand(long neId, String service, String cmdText, String description) throws IOException, InterruptedException {
        Map<String, Object> body = new HashMap<>();
        body.put("ne_id", neId);
        body.put("service", service);
        body.put("cmd_text", cmdText);
        body.put("description", description);
        return Command.fromMap(postJson("/aa/commands", body));
    }

    public Command getCommand(long id) throws IOException, InterruptedException {
        return Command.fromMap(asMap(getJson("/aa/commands/" + id)));
    }

    public void updateCommand(long id, Map<String, Object> patch) throws IOException, InterruptedException {
        request("PUT", "/aa/commands/" + id, patch);
    }

    public void deleteCommand(long id) throws IOException, InterruptedException {
        request("DELETE", "/aa/commands/" + id, null);
    }

    // ── Groups — NE Access Groups ──────────────────────────────────────

    public List<Group> listNeAccessGroups() throws IOException, InterruptedException {
        return toGroups(getJson("/aa/ne-access-groups"));
    }

    public Group createNeAccessGroup(String name, String description) throws IOException, InterruptedException {
        return Group.fromMap(postJson("/aa/ne-access-groups", groupBody(name, description)));
    }

    public void updateNeAccessGroup(long id, String name, String description) throws IOException, InterruptedException {
        request("PUT", "/aa/ne-access-groups/" + id, groupBody(name, description));
    }

    public void deleteNeAccessGroup(long id) throws IOException, InterruptedException {
        request("DELETE", "/aa/ne-access-groups/" + id, null);
    }

    public void addUserToNeAccessGroup(long groupId, long userId) throws IOException, InterruptedException {
        Map<String, Object> b = new HashMap<>(); b.put("user_id", userId);
        postJson("/aa/ne-access-groups/" + groupId + "/users", b);
    }

    public List<Long> listUsersInNeAccessGroup(long groupId) throws IOException, InterruptedException {
        return toLongList(getJson("/aa/ne-access-groups/" + groupId + "/users"));
    }

    public void removeUserFromNeAccessGroup(long groupId, long userId) throws IOException, InterruptedException {
        request("DELETE", "/aa/ne-access-groups/" + groupId + "/users/" + userId, null);
    }

    public void addNeToNeAccessGroup(long groupId, long neId) throws IOException, InterruptedException {
        Map<String, Object> b = new HashMap<>(); b.put("ne_id", neId);
        postJson("/aa/ne-access-groups/" + groupId + "/nes", b);
    }

    public List<Long> listNEsInNeAccessGroup(long groupId) throws IOException, InterruptedException {
        return toLongList(getJson("/aa/ne-access-groups/" + groupId + "/nes"));
    }

    public void removeNEFromNeAccessGroup(long groupId, long neId) throws IOException, InterruptedException {
        request("DELETE", "/aa/ne-access-groups/" + groupId + "/nes/" + neId, null);
    }

    // ── Groups — Cmd Exec Groups ───────────────────────────────────────

    public List<Group> listCmdExecGroups() throws IOException, InterruptedException {
        return toGroups(getJson("/aa/cmd-exec-groups"));
    }

    public Group createCmdExecGroup(String name, String description) throws IOException, InterruptedException {
        return Group.fromMap(postJson("/aa/cmd-exec-groups", groupBody(name, description)));
    }

    public void updateCmdExecGroup(long id, String name, String description) throws IOException, InterruptedException {
        request("PUT", "/aa/cmd-exec-groups/" + id, groupBody(name, description));
    }

    public void deleteCmdExecGroup(long id) throws IOException, InterruptedException {
        request("DELETE", "/aa/cmd-exec-groups/" + id, null);
    }

    public void addUserToCmdExecGroup(long groupId, long userId) throws IOException, InterruptedException {
        Map<String, Object> b = new HashMap<>(); b.put("user_id", userId);
        postJson("/aa/cmd-exec-groups/" + groupId + "/users", b);
    }

    public List<Long> listUsersInCmdExecGroup(long groupId) throws IOException, InterruptedException {
        return toLongList(getJson("/aa/cmd-exec-groups/" + groupId + "/users"));
    }

    public void removeUserFromCmdExecGroup(long groupId, long userId) throws IOException, InterruptedException {
        request("DELETE", "/aa/cmd-exec-groups/" + groupId + "/users/" + userId, null);
    }

    public void addCommandToCmdExecGroup(long groupId, long commandId) throws IOException, InterruptedException {
        Map<String, Object> b = new HashMap<>(); b.put("command_id", commandId);
        postJson("/aa/cmd-exec-groups/" + groupId + "/commands", b);
    }

    public List<Long> listCommandsInCmdExecGroup(long groupId) throws IOException, InterruptedException {
        return toLongList(getJson("/aa/cmd-exec-groups/" + groupId + "/commands"));
    }

    public void removeCommandFromCmdExecGroup(long groupId, long commandId) throws IOException, InterruptedException {
        request("DELETE", "/aa/cmd-exec-groups/" + groupId + "/commands/" + commandId, null);
    }

    private Map<String, Object> groupBody(String name, String description) {
        Map<String, Object> g = new HashMap<>();
        g.put("name", name);
        g.put("description", description);
        return g;
    }

    // ── Policy + Access List + Authorize ────────────────────────────────

    public PasswordPolicy getPasswordPolicy() throws IOException, InterruptedException {
        return PasswordPolicy.fromMap(asMap(getJson("/aa/password-policy")));
    }

    public Map<String, Object> upsertPasswordPolicy(Map<String, Object> policy) throws IOException, InterruptedException {
        return asMap(request("PUT", "/aa/password-policy", policy));
    }

    public List<AccessListEntry> listAccessList(String listType) throws IOException, InterruptedException {
        String path = "/aa/access-list" + (listType != null && !listType.isEmpty() ? ("?list_type=" + listType) : "");
        return toAccessList(getJson(path));
    }

    public AccessListEntry createAccessListEntry(String listType, String matchType, String pattern, String reason) throws IOException, InterruptedException {
        Map<String, Object> body = new HashMap<>();
        body.put("list_type", listType);
        body.put("match_type", matchType);
        body.put("pattern", pattern);
        body.put("reason", reason);
        return AccessListEntry.fromMap(postJson("/aa/access-list", body));
    }

    public void deleteAccessListEntry(long id) throws IOException, InterruptedException {
        request("DELETE", "/aa/access-list/" + id, null);
    }

    /** "Can user X execute command Y on NE Z?" */
    public AuthorizeDecision authorizeCheck(String username, long neId, long commandId) throws IOException, InterruptedException {
        Map<String, Object> body = new HashMap<>();
        body.put("username", username);
        body.put("ne_id", neId);
        body.put("command_id", commandId);
        return AuthorizeDecision.fromMap(postJson("/aa/authorize/check", body));
    }

    // ── History + Config Backup ─────────────────────────────────────────

    public List<HistoryEntry> listHistory(int limit, String scope, String neNamespace, String account) throws IOException, InterruptedException {
        StringBuilder q = new StringBuilder("/aa/history?");
        if (limit > 0) q.append("limit=").append(limit).append("&");
        if (scope != null && !scope.isEmpty()) q.append("scope=").append(scope).append("&");
        if (neNamespace != null && !neNamespace.isEmpty()) q.append("ne_namespace=").append(neNamespace).append("&");
        if (account != null && !account.isEmpty()) q.append("account=").append(account);
        return toHistory(getJson(q.toString()));
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

    public List<ConfigBackup> listConfigBackups(String neName) throws IOException, InterruptedException {
        String path = "/aa/config-backup/list" + (neName != null && !neName.isEmpty() ? ("?ne_name=" + neName) : "");
        return toConfigBackups(getJson(path));
    }

    public ConfigBackup getConfigBackup(long id) throws IOException, InterruptedException {
        return ConfigBackup.fromMap(asMap(getJson("/aa/config-backup/" + id)));
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
        String role = c.authenticate(user, pass);
        System.out.println("logged in as " + user + "  role=" + role);

        // Users
        List<User> users = c.listUsers();
        System.out.println("\nusers: " + users.size());
        for (User u : users) {
            System.out.printf("  #%d  %-15s  role=%-12s  enabled=%s%n",
                    u.id, u.username, u.role, u.isEnabled);
        }

        // NEs
        List<NE> nes = c.listNEs();
        System.out.println("\nNEs: " + nes.size());
        for (NE n : nes) System.out.println("  " + n);

        // Commands
        List<Command> cmds = c.listCommands();
        System.out.println("\ncommands: " + cmds.size());
        for (Command cmd : cmds) System.out.println("  " + cmd);

        // Groups
        List<Group> neGroups = c.listNeAccessGroups();
        System.out.println("\nNE access groups: " + neGroups.size());
        for (Group g : neGroups) System.out.println("  " + g);

        List<Group> cmdGroups = c.listCmdExecGroups();
        System.out.println("cmd exec groups: " + cmdGroups.size());
        for (Group g : cmdGroups) System.out.println("  " + g);

        // Password policy
        PasswordPolicy policy = c.getPasswordPolicy();
        System.out.println("\npassword policy: " + policy);

        // Access list
        List<AccessListEntry> acl = c.listAccessList(null);
        System.out.println("access list entries: " + acl.size());
        for (AccessListEntry e : acl) System.out.println("  " + e);

        // Authorize check
        if (!users.isEmpty() && !nes.isEmpty() && !cmds.isEmpty()) {
            User u = users.get(0);
            NE n = nes.get(0);
            Command cmd = cmds.get(0);
            AuthorizeDecision d = c.authorizeCheck(u.username, n.id, cmd.id);
            System.out.println("\nauthorize(" + u.username + ", ne=" + n.namespace
                    + ", cmd=#" + cmd.id + ") -> "
                    + (d.allowed ? "ALLOW" : "DENY: " + d.reason));
        }

        // History
        List<HistoryEntry> hist = c.listHistory(5, null, null, null);
        System.out.println("\nhistory (last 5): " + hist.size());
        for (HistoryEntry h : hist) System.out.println("  " + h);

        // Config backups
        List<ConfigBackup> backups = c.listConfigBackups(null);
        System.out.println("config backups: " + backups.size());
        for (ConfigBackup b : backups) System.out.println("  " + b);

        System.out.println("\ndone.");
    }
}
