/* mgt_client.c — libcurl-backed implementation of mgt_client.h.
 *
 * Parses the small amount of JSON we actually need (top-level object lookup
 * of a string or boolean) by hand. If you already link against cJSON or
 * json-c, feel free to swap json_find_string / json_find_bool out. */

#include "mgt_client.h"

#include <curl/curl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

struct mgt_client {
    char *base_url;
    char *token;       /* "Basic <jwt>" returned by /aa/authenticate */
    CURL *curl;
};

typedef struct {
    char   *data;
    size_t  len;
    size_t  cap;
} buf_t;

static size_t buf_write_cb(void *ptr, size_t size, size_t nmemb, void *userdata) {
    buf_t *b = (buf_t *)userdata;
    size_t n = size * nmemb;
    if (b->len + n + 1 > b->cap) {
        size_t new_cap = b->cap ? b->cap * 2 : 1024;
        while (new_cap < b->len + n + 1) new_cap *= 2;
        char *p = realloc(b->data, new_cap);
        if (!p) return 0;
        b->data = p;
        b->cap  = new_cap;
    }
    memcpy(b->data + b->len, ptr, n);
    b->len += n;
    b->data[b->len] = '\0';
    return n;
}

static char *strdup_or_null(const char *s) {
    if (!s) return NULL;
    char *d = malloc(strlen(s) + 1);
    if (d) strcpy(d, s);
    return d;
}

mgt_client_t *mgt_client_new(const char *base_url) {
    if (!base_url) return NULL;
    mgt_client_t *c = calloc(1, sizeof(*c));
    if (!c) return NULL;
    /* trim trailing slash */
    size_t n = strlen(base_url);
    while (n > 0 && base_url[n-1] == '/') n--;
    c->base_url = malloc(n + 1);
    if (!c->base_url) { free(c); return NULL; }
    memcpy(c->base_url, base_url, n);
    c->base_url[n] = '\0';

    curl_global_init(CURL_GLOBAL_DEFAULT);
    c->curl = curl_easy_init();
    if (!c->curl) { free(c->base_url); free(c); return NULL; }
    return c;
}

void mgt_client_free(mgt_client_t *c) {
    if (!c) return;
    if (c->curl) curl_easy_cleanup(c->curl);
    free(c->base_url);
    free(c->token);
    free(c);
}

void mgt_set_token(mgt_client_t *c, const char *token) {
    if (!c) return;
    free(c->token);
    c->token = strdup_or_null(token);
}

/* ── HTTP ───────────────────────────────────────────────────────────── */

int mgt_request(mgt_client_t *c,
                const char *method,
                const char *path,
                const char *body,
                char **out_body,
                long  *out_http_status)
{
    if (!c || !method || !path) return -1;
    buf_t resp = {0};
    char  url[1024];
    snprintf(url, sizeof(url), "%s%s", c->base_url, path);

    curl_easy_reset(c->curl);
    curl_easy_setopt(c->curl, CURLOPT_URL, url);
    curl_easy_setopt(c->curl, CURLOPT_CUSTOMREQUEST, method);
    curl_easy_setopt(c->curl, CURLOPT_WRITEFUNCTION, buf_write_cb);
    curl_easy_setopt(c->curl, CURLOPT_WRITEDATA, &resp);
    curl_easy_setopt(c->curl, CURLOPT_TIMEOUT, 15L);
    curl_easy_setopt(c->curl, CURLOPT_CONNECTTIMEOUT, 10L);

    struct curl_slist *hdrs = NULL;
    hdrs = curl_slist_append(hdrs, "Content-Type: application/json");
    if (c->token && *c->token) {
        char auth[1024];
        snprintf(auth, sizeof(auth), "Authorization: %s", c->token);
        hdrs = curl_slist_append(hdrs, auth);
    }
    curl_easy_setopt(c->curl, CURLOPT_HTTPHEADER, hdrs);

    if (body) {
        curl_easy_setopt(c->curl, CURLOPT_POSTFIELDS, body);
        curl_easy_setopt(c->curl, CURLOPT_POSTFIELDSIZE, (long)strlen(body));
    }

    CURLcode rc = curl_easy_perform(c->curl);
    long status = 0;
    curl_easy_getinfo(c->curl, CURLINFO_RESPONSE_CODE, &status);
    curl_slist_free_all(hdrs);

    if (out_http_status) *out_http_status = status;

    if (rc != CURLE_OK) {
        fprintf(stderr, "mgt: %s %s: %s\n", method, path, curl_easy_strerror(rc));
        free(resp.data);
        return -2;
    }
    if (status < 200 || status >= 300) {
        fprintf(stderr, "mgt: %s %s: HTTP %ld: %s\n", method, path, status,
                resp.data ? resp.data : "");
        if (out_body) { *out_body = resp.data; } else { free(resp.data); }
        return (int)status;
    }
    if (out_body) *out_body = resp.data;
    else          free(resp.data);
    return 0;
}

int mgt_get(mgt_client_t *c, const char *path, char **out) {
    long st;
    return mgt_request(c, "GET", path, NULL, out, &st);
}
int mgt_post(mgt_client_t *c, const char *path, const char *body, char **out) {
    long st;
    return mgt_request(c, "POST", path, body, out, &st);
}
int mgt_put(mgt_client_t *c, const char *path, const char *body, char **out) {
    long st;
    return mgt_request(c, "PUT", path, body, out, &st);
}
int mgt_delete(mgt_client_t *c, const char *path) {
    long st;
    return mgt_request(c, "DELETE", path, NULL, NULL, &st);
}

/* ── Tiny JSON helpers ─────────────────────────────────────────────────
 *
 * These only handle the top-level flat-object shape the v2 API returns.
 * For nested traversal the caller should plug in a real JSON library. */

/* json_find_raw: find `"key":` at top level and return a pointer to the first
 * char of the value (after the colon, skipping whitespace). Returns NULL if
 * not found. Does not properly handle nested objects — fine for flat v2
 * responses. */
static const char *json_find_raw(const char *s, const char *key) {
    if (!s || !key) return NULL;
    size_t klen = strlen(key);
    const char *p = s;
    while ((p = strchr(p, '"')) != NULL) {
        if (strncmp(p + 1, key, klen) == 0 && p[1 + klen] == '"') {
            p += 2 + klen;
            while (*p && *p != ':') p++;
            if (!*p) return NULL;
            p++;
            while (*p == ' ' || *p == '\t' || *p == '\n') p++;
            return p;
        }
        p++;
    }
    return NULL;
}

/* json_find_string: returns a malloc'd copy of the string value for `key`.
 * Returns NULL if absent or not a string. Unescapes \" \\ \n \r \t \/ but
 * not \u escapes — fine for our ASCII responses. */
static char *json_find_string(const char *s, const char *key) {
    const char *p = json_find_raw(s, key);
    if (!p || *p != '"') return NULL;
    p++;
    const char *start = p;
    size_t cap = 32, len = 0;
    char *out = malloc(cap);
    if (!out) return NULL;
    while (*p && *p != '"') {
        char c = *p++;
        if (c == '\\' && *p) {
            char esc = *p++;
            switch (esc) {
                case '"':  c = '"';  break;
                case '\\': c = '\\'; break;
                case '/':  c = '/';  break;
                case 'n':  c = '\n'; break;
                case 'r':  c = '\r'; break;
                case 't':  c = '\t'; break;
                case 'b':  c = '\b'; break;
                case 'f':  c = '\f'; break;
                default:   c = esc;  break;
            }
        }
        if (len + 2 > cap) {
            cap *= 2;
            char *np = realloc(out, cap);
            if (!np) { free(out); return NULL; }
            out = np;
        }
        out[len++] = c;
    }
    out[len] = '\0';
    (void)start;
    return out;
}

/* json_find_bool: 1 if value is true, 0 if false, -1 if absent. */
static int json_find_bool(const char *s, const char *key) {
    const char *p = json_find_raw(s, key);
    if (!p) return -1;
    if (strncmp(p, "true",  4) == 0) return 1;
    if (strncmp(p, "false", 5) == 0) return 0;
    return -1;
}

/* ── Higher-level helpers ──────────────────────────────────────────────── */

int mgt_authenticate(mgt_client_t *c, const char *username, const char *password) {
    if (!c || !username || !password) return -1;
    char body[512];
    snprintf(body, sizeof(body),
             "{\"username\":\"%s\",\"password\":\"%s\"}", username, password);
    char *resp = NULL;
    int rc = mgt_post(c, "/aa/authenticate", body, &resp);
    if (rc != 0) { free(resp); return rc; }
    char *tok = json_find_string(resp, "token");
    free(resp);
    if (!tok) return -3;
    mgt_set_token(c, tok);
    free(tok);
    return 0;
}

int mgt_authorize_check(mgt_client_t *c,
                        const char *username,
                        int64_t ne_id,
                        int64_t command_id,
                        authorize_decision_t *out)
{
    if (!c || !username || !out) return -1;
    memset(out, 0, sizeof(*out));
    char body[512];
    snprintf(body, sizeof(body),
             "{\"username\":\"%s\",\"ne_id\":%lld,\"command_id\":%lld}",
             username, (long long)ne_id, (long long)command_id);
    char *resp = NULL;
    int rc = mgt_post(c, "/aa/authorize/check", body, &resp);
    if (rc != 0) { free(resp); return rc; }

    out->allowed              = json_find_bool(resp, "allowed") == 1;
    out->user_exists          = json_find_bool(resp, "user_exists") == 1;
    out->user_enabled         = json_find_bool(resp, "user_enabled") == 1;
    out->ne_reachable         = json_find_bool(resp, "ne_reachable") == 1;
    out->command_on_ne        = json_find_bool(resp, "command_on_ne") == 1;
    out->command_exec_allowed = json_find_bool(resp, "command_exec_allowed") == 1;

    char *reason = json_find_string(resp, "reason");
    if (reason) {
        strncpy(out->reason, reason, sizeof(out->reason) - 1);
        out->reason[sizeof(out->reason) - 1] = '\0';
        free(reason);
    }
    free(resp);
    return 0;
}

int mgt_save_history(mgt_client_t *c,
                     const char *account,
                     const char *cmd_text,
                     const char *ne_namespace,
                     const char *ne_ip,
                     const char *scope,
                     const char *result)
{
    if (!c) return -1;
    /* Build JSON manually — keep it to ~1KB, sufficient for audit payloads. */
    char body[2048];
    int n = snprintf(body, sizeof(body),
        "{\"account\":\"%s\",\"cmd_text\":\"%s\",\"ne_namespace\":\"%s\","
        "\"ne_ip\":\"%s\",\"scope\":\"%s\",\"result\":\"%s\"}",
        account      ? account      : "",
        cmd_text     ? cmd_text     : "",
        ne_namespace ? ne_namespace : "",
        ne_ip        ? ne_ip        : "",
        scope        ? scope        : "",
        result       ? result       : "");
    if (n < 0 || n >= (int)sizeof(body)) return -2;
    char *resp = NULL;
    int rc = mgt_post(c, "/aa/history/save", body, &resp);
    free(resp);
    return rc;
}
