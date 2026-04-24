package com.example.mgt;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpRequest.BodyPublishers;
import java.net.http.HttpResponse;
import java.time.Duration;

/**
 * Minimal client for cli-mgt-svc history endpoint.
 *
 * Endpoint: POST {baseUrl}/aa/history/save  — **unauthenticated** (no JWT required).
 *           Any downstream service (SSH proxy, ne-config, ne-command, etc.) can
 *           log a command without having to carry the user's token.
 *
 * Body fields:
 *   cmd_name  (required) — the command string that was run
 *   ne_name   (required) — NE target
 *   ne_ip     (optional) — NE IP
 *   scope     (optional) — "ne-command" | "ne-config" | "cli-config"
 *   result    (optional) — "success" | "failure" | free text
 *   account   (optional) — username that ran the command. Strongly
 *                          recommended — otherwise the audit trail records
 *                          "unknown". Pass the user identity resolved from
 *                          whatever SSH/auth layer the caller sits behind.
 *
 * Response: 201 on success, 400 if cmd_name or ne_name is empty, 500 on DB error.
 *
 * Java 11+ — uses java.net.http.HttpClient, no external dependencies.
 */
public final class MgtHistoryClient {

    private static final HttpClient HTTP = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(5))
            .build();

    private MgtHistoryClient() {}

    /**
     * Save a history record. Pass {@code account} = username of the actor that
     * ran the command; pass empty/null to let the server log "unknown".
     */
    public static int saveHistory(String baseUrl,
                                  String cmdName,
                                  String neName,
                                  String neIp,
                                  String scope,
                                  String result,
                                  String account) throws Exception {

        String body = "{"
                + "\"cmd_name\":" + jsonStr(cmdName) + ","
                + "\"ne_name\":"  + jsonStr(neName)  + ","
                + "\"ne_ip\":"    + jsonStr(neIp)    + ","
                + "\"scope\":"    + jsonStr(scope)   + ","
                + "\"result\":"   + jsonStr(result)  + ","
                + "\"account\":"  + jsonStr(account)
                + "}";

        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + "/aa/history/save"))
                .timeout(Duration.ofSeconds(10))
                .header("Content-Type", "application/json")
                .POST(BodyPublishers.ofString(body))
                .build();

        HttpResponse<String> resp = HTTP.send(req, HttpResponse.BodyHandlers.ofString());
        if (resp.statusCode() != 201) {
            System.err.printf("save history failed: %d %s%n", resp.statusCode(), resp.body());
        }
        return resp.statusCode();
    }

    /** Backwards-compat overload — drops the account field (logs "unknown"). */
    public static int saveHistory(String baseUrl,
                                  String cmdName,
                                  String neName,
                                  String neIp,
                                  String scope,
                                  String result) throws Exception {
        return saveHistory(baseUrl, cmdName, neName, neIp, scope, result, null);
    }

    /** Minimal JSON string escaper — handles null, quotes, backslash, and control chars. */
    private static String jsonStr(String s) {
        if (s == null) return "\"\"";
        StringBuilder sb = new StringBuilder(s.length() + 2).append('"');
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
        return sb.append('"').toString();
    }

    // Example usage:
    //
    //   // Preferred — include the actor username so the audit trail is meaningful.
    //   int status = MgtHistoryClient.saveHistory(
    //       "http://mgt-svc:3000",
    //       "show running-config",
    //       "HTSMF01",
    //       "10.10.1.1",
    //       "ne-command",
    //       "success",
    //       "alice"
    //   );
}
