package com.example.mgt;

import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.function.Function;

/**
 * Model classes + zero-dependency JSON parser for cli-mgt-svc responses.
 *
 * Every model has a static {@code from(Object)} factory that accepts the
 * output of {@link #parse(String)}. Combine with
 * {@link MgtServiceClient.Response#as} / {@link MgtServiceClient.Response#asList}:
 *
 * <pre>{@code
 *   List<MgtModels.Ne> nes =
 *       MgtServiceClient.listNe().asList(MgtModels.Ne::from);
 *
 *   MgtModels.AuthResult auth =
 *       MgtServiceClient.authenticate("user", "pass").as(MgtModels.AuthResult::from);
 * }</pre>
 *
 * Java 17+ (records). No external dependencies.
 */
public final class MgtModels {

    private MgtModels() {}

    // ═══════════════════════════════════════════════════════════════════════
    //  Parse entry point
    // ═══════════════════════════════════════════════════════════════════════

    /** Parse JSON text into Map / List / String / Long / Double / Boolean / null. */
    public static Object parse(String json) {
        if (json == null || json.isEmpty()) return null;
        JsonParser p = new JsonParser(json);
        p.skipWs();
        Object v = p.value();
        p.skipWs();
        return v;
    }

    // ═══════════════════════════════════════════════════════════════════════
    //  Envelope (ResSuccess / ResError)
    // ═══════════════════════════════════════════════════════════════════════

    /** {@code {status, code, message, error?}} wrapper returned by most mutations. */
    public record Envelope(Boolean status, Integer code, String message, String error) {
        public static Envelope from(Object o) {
            if (!(o instanceof Map)) {
                return new Envelope(null, null, o == null ? null : String.valueOf(o), null);
            }
            Map<?, ?> m = (Map<?, ?>) o;
            return new Envelope(F.bool(m, "status"), F.i(m, "code"),
                    F.s(m, "message"), F.s(m, "error"));
        }
    }

    // ═══════════════════════════════════════════════════════════════════════
    //  Auth
    // ═══════════════════════════════════════════════════════════════════════

    /** Response body for {@code POST /aa/authenticate}. token already carries the {@code "Basic "} prefix. */
    public record AuthResult(String status, String token, String responseCode, String systemType) {
        public static AuthResult from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new AuthResult(F.s(m, "status"), F.s(m, "response_data"),
                    F.s(m, "response_code"), F.s(m, "system_type"));
        }
    }

    /** Response body for {@code POST /aa/validate-token}. */
    public record ValidateTokenResult(String username, String roles) {
        public static ValidateTokenResult from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new ValidateTokenResult(F.s(m, "username"), F.s(m, "roles"));
        }
    }

    // ═══════════════════════════════════════════════════════════════════════
    //  User
    // ═══════════════════════════════════════════════════════════════════════

    /** Item from {@code GET /aa/authenticate/user/show}. */
    public record UserShow(String username, String role, List<NeRef> nes) {
        public static UserShow from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new UserShow(F.s(m, "username"), F.s(m, "role"),
                    F.listOf(m, "tblNes", NeRef::from));
        }
    }

    public record NeRef(String ne, String site) {
        public static NeRef from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new NeRef(F.s(m, "ne"), F.s(m, "site"));
        }
    }

    /** Item from {@code GET /aa/authorize/user/show}. */
    public record UserPermission(String username, String permission) {
        public static UserPermission from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new UserPermission(F.s(m, "username"), F.s(m, "permission"));
        }
    }

    /** Item from {@code GET /aa/admin/user/list} — full account row without password. */
    public record AdminUser(
            Long accountId, String accountName, String fullName, String email,
            String address, String phoneNumber, Integer accountType, String description,
            Boolean isEnable, Boolean status, String createdBy, Integer loginFailureCount) {
        public static AdminUser from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new AdminUser(
                    F.l(m, "account_id"), F.s(m, "account_name"), F.s(m, "full_name"),
                    F.s(m, "email"), F.s(m, "address"), F.s(m, "phone_number"),
                    F.i(m, "account_type"), F.s(m, "description"),
                    F.bool(m, "is_enable"), F.bool(m, "status"),
                    F.s(m, "created_by"), F.i(m, "login_failure_count"));
        }
    }

    // ═══════════════════════════════════════════════════════════════════════
    //  Network Element
    // ═══════════════════════════════════════════════════════════════════════

    /** Item from {@code GET /aa/authorize/ne/show} — basic NE summary. */
    public record NeShow(Long id, String name, String siteName, String ipAddress,
                         Integer port, String namespace, String description) {
        public static NeShow from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new NeShow(F.l(m, "id"), F.s(m, "name"), F.s(m, "site_name"),
                    F.s(m, "ip_address"), F.i(m, "port"),
                    F.s(m, "namespace"), F.s(m, "description"));
        }
    }

    /** Item from {@code GET /aa/list/ne} — NE the caller is authorized for, full fields. */
    public record Ne(
            String site, String ne, String ip, String description, String namespace,
            Integer port, String commandUrl, String confMode,
            String confMasterIp, String confSlaveIp,
            Integer confPortMasterSsh, Integer confPortSlaveSsh,
            Integer confPortMasterTcp, Integer confPortSlaveTcp,
            String confUsername, String confPassword,
            List<UrlEndpoint> urlList) {
        public static Ne from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new Ne(
                    F.s(m, "site"), F.s(m, "ne"), F.s(m, "ip"),
                    F.s(m, "description"), F.s(m, "namespace"), F.i(m, "port"),
                    F.s(m, "command_url"), F.s(m, "conf_mode"),
                    F.s(m, "conf_master_ip"), F.s(m, "conf_slave_ip"),
                    F.i(m, "conf_port_master_ssh"), F.i(m, "conf_port_slave_ssh"),
                    F.i(m, "conf_port_master_tcp"), F.i(m, "conf_port_slave_tcp"),
                    F.s(m, "conf_username"), F.s(m, "conf_password"),
                    F.listOf(m, "urlList", UrlEndpoint::from));
        }
    }

    /** Item from {@code GET /aa/list/ne/monitor}. */
    public record NeMonitor(String site, String ne, String ip, String description,
                            String namespace, Integer port, String neMonitorUrl) {
        public static NeMonitor from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new NeMonitor(
                    F.s(m, "site"), F.s(m, "ne"), F.s(m, "ip"),
                    F.s(m, "description"), F.s(m, "namespace"),
                    F.i(m, "port"), F.s(m, "ne_monitor_url"));
        }
    }

    /** Item from {@code GET /aa/admin/ne/list} — full CliNe row. */
    public record CliNe(
            Long id, String neName, String namespace, String siteName, String systemType,
            String description, String commandUrl, String confMode,
            String confMasterIp, String confSlaveIp,
            Integer confPortMasterSsh, Integer confPortSlaveSsh,
            Integer confPortMasterTcp, Integer confPortSlaveTcp,
            String confUsername, String confPassword) {
        public static CliNe from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new CliNe(
                    F.l(m, "id"), F.s(m, "ne_name"), F.s(m, "namespace"),
                    F.s(m, "site_name"), F.s(m, "system_type"),
                    F.s(m, "description"), F.s(m, "command_url"), F.s(m, "conf_mode"),
                    F.s(m, "conf_master_ip"), F.s(m, "conf_slave_ip"),
                    F.i(m, "conf_port_master_ssh"), F.i(m, "conf_port_slave_ssh"),
                    F.i(m, "conf_port_master_tcp"), F.i(m, "conf_port_slave_tcp"),
                    F.s(m, "conf_username"), F.s(m, "conf_password"));
        }
    }

    /** Entry in Ne.urlList. */
    public record UrlEndpoint(String ipAddress, Integer port) {
        public static UrlEndpoint from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new UrlEndpoint(F.s(m, "ipAddress"), F.i(m, "port"));
        }
    }

    // ═══════════════════════════════════════════════════════════════════════
    //  NE Config
    // ═══════════════════════════════════════════════════════════════════════

    /** Item from {@code GET /aa/authorize/ne/config/list}. */
    public record NeConfig(
            Long id, Long neId, String neName, String ipAddress, Integer port,
            String username, String password, String protocol, String description) {
        public static NeConfig from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new NeConfig(
                    F.l(m, "id"), F.l(m, "ne_id"), F.s(m, "ne_name"),
                    F.s(m, "ip_address"), F.i(m, "port"),
                    F.s(m, "username"), F.s(m, "password"),
                    F.s(m, "protocol"), F.s(m, "description"));
        }
    }

    // ═══════════════════════════════════════════════════════════════════════
    //  Config Backup
    // ═══════════════════════════════════════════════════════════════════════

    /** Response body for {@code POST /aa/config-backup/save}. */
    public record ConfigBackupSaveResult(String status, Long id) {
        public static ConfigBackupSaveResult from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new ConfigBackupSaveResult(F.s(m, "status"), F.l(m, "id"));
        }
    }

    /** Response body for {@code GET /aa/config-backup/list}. */
    public record ConfigBackupListResult(String status, List<ConfigBackup> backups) {
        public static ConfigBackupListResult from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new ConfigBackupListResult(
                    F.s(m, "status"),
                    F.listOf(m, "backups", ConfigBackup::from));
        }
    }

    public record ConfigBackup(Long id, String neName, String neIp, Long size, String createdAt) {
        public static ConfigBackup from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new ConfigBackup(
                    F.l(m, "id"), F.s(m, "ne_name"), F.s(m, "ne_ip"),
                    F.l(m, "size"), F.s(m, "created_at"));
        }
    }

    /** Response body for {@code GET /aa/config-backup/{id}} — metadata + full XML. */
    public record ConfigBackupDetail(String status, Long id, String neName, String neIp,
                                     Long size, String createdAt, String configXml) {
        public static ConfigBackupDetail from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new ConfigBackupDetail(
                    F.s(m, "status"), F.l(m, "id"),
                    F.s(m, "ne_name"), F.s(m, "ne_ip"),
                    F.l(m, "size"), F.s(m, "created_at"),
                    F.s(m, "config_xml"));
        }
    }

    // ═══════════════════════════════════════════════════════════════════════
    //  History
    // ═══════════════════════════════════════════════════════════════════════

    /** Item from {@code GET /aa/history/list}. */
    public record History(
            Integer id, String account, String cmdName, String neName, String neIp,
            String ipAddress, String scope, String result,
            String createdDate, String executedTime) {
        public static History from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new History(
                    F.i(m, "id"), F.s(m, "account"), F.s(m, "cmd_name"),
                    F.s(m, "ne_name"), F.s(m, "ne_ip"), F.s(m, "ip_address"),
                    F.s(m, "scope"), F.s(m, "result"),
                    F.s(m, "created_date"), F.s(m, "executed_time"));
        }
    }

    // ═══════════════════════════════════════════════════════════════════════
    //  Import
    // ═══════════════════════════════════════════════════════════════════════

    /** Item from {@code POST /aa/import/} response. */
    public record ImportResult(String type, String name, String status, String detail) {
        public static ImportResult from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new ImportResult(F.s(m, "type"), F.s(m, "name"),
                    F.s(m, "status"), F.s(m, "detail"));
        }
    }

    // ═══════════════════════════════════════════════════════════════════════
    //  Subscribers
    // ═══════════════════════════════════════════════════════════════════════

    public record SubscriberFile(String name, Integer index, Long sizeBytes) {
        public static SubscriberFile from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new SubscriberFile(F.s(m, "name"), F.i(m, "index"), F.l(m, "size_bytes"));
        }
    }

    public record SubscriberFileContent(String name, Integer index, List<String> lines, Integer total) {
        public static SubscriberFileContent from(Object o) {
            if (!(o instanceof Map)) return null;
            Map<?, ?> m = (Map<?, ?>) o;
            return new SubscriberFileContent(
                    F.s(m, "name"), F.i(m, "index"),
                    F.stringList(m, "lines"), F.i(m, "total"));
        }
    }

    // ═══════════════════════════════════════════════════════════════════════
    //  Field accessors (null-tolerant, accepting parsed JSON Map values)
    // ═══════════════════════════════════════════════════════════════════════

    static final class F {
        private F() {}

        static String s(Map<?, ?> m, String k) {
            Object v = m.get(k);
            return v == null ? null : String.valueOf(v);
        }

        static Integer i(Map<?, ?> m, String k) {
            Object v = m.get(k);
            if (v == null) return null;
            if (v instanceof Number) return ((Number) v).intValue();
            try { return Integer.parseInt(String.valueOf(v)); }
            catch (NumberFormatException e) { return null; }
        }

        static Long l(Map<?, ?> m, String k) {
            Object v = m.get(k);
            if (v == null) return null;
            if (v instanceof Number) return ((Number) v).longValue();
            try { return Long.parseLong(String.valueOf(v)); }
            catch (NumberFormatException e) { return null; }
        }

        static Boolean bool(Map<?, ?> m, String k) {
            Object v = m.get(k);
            if (v == null) return null;
            if (v instanceof Boolean) return (Boolean) v;
            return Boolean.parseBoolean(String.valueOf(v));
        }

        static <T> List<T> listOf(Map<?, ?> m, String k, Function<Object, T> factory) {
            Object v = m.get(k);
            if (!(v instanceof List)) return Collections.emptyList();
            List<T> out = new ArrayList<>();
            for (Object item : (List<?>) v) {
                T parsed = factory.apply(item);
                if (parsed != null) out.add(parsed);
            }
            return out;
        }

        static List<String> stringList(Map<?, ?> m, String k) {
            Object v = m.get(k);
            if (!(v instanceof List)) return Collections.emptyList();
            List<String> out = new ArrayList<>();
            for (Object item : (List<?>) v) {
                out.add(item == null ? null : String.valueOf(item));
            }
            return out;
        }
    }

    // ═══════════════════════════════════════════════════════════════════════
    //  JSON parser (minimal, JSON-RFC-8259 subset — no comments, no trailing commas)
    // ═══════════════════════════════════════════════════════════════════════

    private static final class JsonParser {
        final String s;
        int i;

        JsonParser(String s) { this.s = s; }

        void skipWs() {
            while (i < s.length()) {
                char c = s.charAt(i);
                if (c == ' ' || c == '\t' || c == '\n' || c == '\r') i++;
                else break;
            }
        }

        char peek() { return i < s.length() ? s.charAt(i) : 0; }

        Object value() {
            skipWs();
            char c = peek();
            if (c == '{') return parseObject();
            if (c == '[') return parseArray();
            if (c == '"') return parseString();
            if (c == 't' || c == 'f') return parseBool();
            if (c == 'n') { expect("null"); return null; }
            return parseNumber();
        }

        Map<String, Object> parseObject() {
            Map<String, Object> m = new LinkedHashMap<>();
            i++; // {
            skipWs();
            if (peek() == '}') { i++; return m; }
            while (true) {
                skipWs();
                String key = parseString();
                skipWs();
                if (peek() != ':') throw new RuntimeException("expected ':' at " + i);
                i++;
                Object v = value();
                m.put(key, v);
                skipWs();
                char c = peek();
                if (c == ',') { i++; continue; }
                if (c == '}') { i++; return m; }
                throw new RuntimeException("expected ',' or '}' at " + i);
            }
        }

        List<Object> parseArray() {
            List<Object> list = new ArrayList<>();
            i++; // [
            skipWs();
            if (peek() == ']') { i++; return list; }
            while (true) {
                list.add(value());
                skipWs();
                char c = peek();
                if (c == ',') { i++; continue; }
                if (c == ']') { i++; return list; }
                throw new RuntimeException("expected ',' or ']' at " + i);
            }
        }

        String parseString() {
            if (peek() != '"') throw new RuntimeException("expected '\"' at " + i);
            i++;
            StringBuilder sb = new StringBuilder();
            while (i < s.length()) {
                char c = s.charAt(i++);
                if (c == '"') return sb.toString();
                if (c == '\\') {
                    if (i >= s.length()) throw new RuntimeException("bad escape at end");
                    char n = s.charAt(i++);
                    switch (n) {
                        case '"':  sb.append('"');  break;
                        case '\\': sb.append('\\'); break;
                        case '/':  sb.append('/');  break;
                        case 'b':  sb.append('\b'); break;
                        case 'f':  sb.append('\f'); break;
                        case 'n':  sb.append('\n'); break;
                        case 'r':  sb.append('\r'); break;
                        case 't':  sb.append('\t'); break;
                        case 'u':
                            if (i + 4 > s.length()) throw new RuntimeException("bad \\u escape");
                            sb.append((char) Integer.parseInt(s.substring(i, i + 4), 16));
                            i += 4;
                            break;
                        default: sb.append(n);
                    }
                } else {
                    sb.append(c);
                }
            }
            throw new RuntimeException("unterminated string");
        }

        Boolean parseBool() {
            if (s.startsWith("true", i))  { i += 4; return Boolean.TRUE; }
            if (s.startsWith("false", i)) { i += 5; return Boolean.FALSE; }
            throw new RuntimeException("expected bool at " + i);
        }

        void expect(String lit) {
            if (!s.startsWith(lit, i)) throw new RuntimeException("expected '" + lit + "' at " + i);
            i += lit.length();
        }

        Object parseNumber() {
            int start = i;
            if (peek() == '-') i++;
            while (i < s.length()) {
                char c = s.charAt(i);
                if ((c >= '0' && c <= '9') || c == '.' || c == 'e' || c == 'E' || c == '+' || c == '-') {
                    i++;
                } else {
                    break;
                }
            }
            String t = s.substring(start, i);
            if (t.contains(".") || t.contains("e") || t.contains("E")) {
                return Double.parseDouble(t);
            }
            try { return Long.parseLong(t); }
            catch (NumberFormatException e) { return Double.parseDouble(t); }
        }
    }
}
