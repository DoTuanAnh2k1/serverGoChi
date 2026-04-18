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
 * Endpoint: POST {baseUrl}/aa/history/save
 * Auth:     header "Authorization: Basic <jwt_token>" (token returned by /aa/authenticate
 *           already contains the "Basic " prefix — pass it as-is).
 * Body:     {"cmd_name", "ne_name", "ne_ip", "scope", "result"}
 *           cmd_name + ne_name are required; the rest may be empty/null.
 *           account is taken from JWT context, do NOT send it in the body.
 *
 * Returns HTTP status code (201 on success, 400 = missing required field, 401 = bad token).
 *
 * Java 11+ — uses java.net.http.HttpClient, no external dependencies.
 */
public final class MgtHistoryClient {

    private static final HttpClient HTTP = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(5))
            .build();

    private MgtHistoryClient() {}

    public static int saveHistory(String baseUrl,
                                  String token,
                                  String cmdName,
                                  String neName,
                                  String neIp,
                                  String scope,
                                  String result) throws Exception {

        String body = "{"
                + "\"cmd_name\":" + jsonStr(cmdName) + ","
                + "\"ne_name\":"  + jsonStr(neName)  + ","
                + "\"ne_ip\":"    + jsonStr(neIp)    + ","
                + "\"scope\":"    + jsonStr(scope)   + ","
                + "\"result\":"   + jsonStr(result)
                + "}";

        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + "/aa/history/save"))
                .timeout(Duration.ofSeconds(10))
                .header("Content-Type", "application/json")
                .header("Authorization", token)
                .POST(BodyPublishers.ofString(body))
                .build();

        HttpResponse<String> resp = HTTP.send(req, HttpResponse.BodyHandlers.ofString());
        if (resp.statusCode() != 201) {
            System.err.printf("save history failed: %d %s%n", resp.statusCode(), resp.body());
        }
        return resp.statusCode();
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
    //   int status = MgtHistoryClient.saveHistory(
    //       "http://mgt-svc:3000",
    //       "Basic eyJhbGciOi...",
    //       "show running-config",
    //       "HTSMF01",
    //       "10.10.1.1",
    //       "ne-command",
    //       "success"
    //   );
}
