/* demo.c — end-to-end usage of mgt_client against a running cli-mgt v2.
 *
 * Build:
 *   gcc -Wall -O2 -o demo mgt_client.c demo.c -lcurl
 *
 * Run:
 *   MGT_BASE=http://localhost:3000 MGT_USER=admin MGT_PASS=admin ./demo
 *
 * The demo logs in, creates resources through the full v2 API, wires them
 * together via groups, runs an authorize check, pushes an audit event, and
 * finally tears everything down. Read mgt_client.h for the full API surface.
 */

#include "mgt_client.h"

#include <inttypes.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* ── helpers ─────────────────────────────────────────────────────────── */

static const char *env_or(const char *key, const char *fallback) {
    const char *v = getenv(key);
    return (v && *v) ? v : fallback;
}

/* Tiny helper: find first "id": <number> in a JSON response. Returns -1
 * if not found. Good enough for the single-object responses that create
 * endpoints return. */
static int64_t grab_id(const char *json) {
    if (!json) return -1;
    const char *p = strstr(json, "\"id\"");
    if (!p) return -1;
    p += 4;
    while (*p && *p != ':') p++;
    if (!*p) return -1;
    p++;
    while (*p == ' ' || *p == '\t' || *p == '\n') p++;
    if (*p < '0' || *p > '9') return -1;
    int64_t v = 0;
    while (*p >= '0' && *p <= '9') { v = v * 10 + (*p - '0'); p++; }
    return v;
}

static void print_json(const char *label, const char *json) {
    if (!json) { printf("%s: (null)\n", label); return; }
    printf("%s: %.400s%s\n", label, json, strlen(json) > 400 ? " ..." : "");
}

#define CHECK(expr, label) do { \
    if ((expr) != 0) { fprintf(stderr, "FAIL: %s\n", label); goto cleanup; } \
    printf("OK:   %s\n", label); \
} while (0)

/* ── main ────────────────────────────────────────────────────────────── */

int main(void) {
    const char *base = env_or("MGT_BASE", "http://localhost:3000");
    const char *user = env_or("MGT_USER", "admin");
    const char *pass = env_or("MGT_PASS", "admin");

    int64_t created_user_id   = -1;
    int64_t created_ne_id     = -1;
    int64_t created_cmd_id    = -1;
    int64_t ne_grp_id         = -1;
    int64_t cmd_grp_id        = -1;
    char   *body              = NULL;

    mgt_client_t *c = mgt_client_new(base);
    if (!c) { fprintf(stderr, "mgt_client_new failed\n"); return 1; }

    /* ── 1. Authenticate ─────────────────────────────────────────────── */

    CHECK(mgt_authenticate(c, user, pass), "authenticate");
    printf("      logged in as %s  role=%s\n", user,
           mgt_get_role(c) ? mgt_get_role(c) : "(unknown)");

    /* ── 2. List users (raw JSON) ────────────────────────────────────── */

    if (mgt_list_users(c, &body) == 0) { print_json("users", body); free(body); body = NULL; }

    /* ── 3. Create a user ────────────────────────────────────────────── */

    CHECK(mgt_create_user(c, "demo_operator", "Demo1234!", "op@example.com",
                          "Demo Operator", "user", &body),
          "create user demo_operator");
    created_user_id = grab_id(body);
    printf("      user id=%" PRId64 "\n", created_user_id);
    free(body); body = NULL;

    /* ── 4. Create an NE ─────────────────────────────────────────────── */

    CHECK(mgt_create_ne(c,
        "{\"namespace\":\"lab01\",\"ne_type\":\"router\",\"site_name\":\"LAB\","
        "\"description\":\"demo NE\",\"master_ip\":\"10.0.0.1\",\"master_port\":22,"
        "\"ssh_username\":\"netops\",\"ssh_password\":\"s3cret\","
        "\"conf_mode\":\"SSH\"}", &body),
          "create NE lab01");
    created_ne_id = grab_id(body);
    printf("      ne id=%" PRId64 "\n", created_ne_id);
    free(body); body = NULL;

    /* ── 5. Create a command ─────────────────────────────────────────── */

    CHECK(mgt_create_command(c, created_ne_id, "ne-command",
                             "show version", "display SW version", &body),
          "create command 'show version'");
    created_cmd_id = grab_id(body);
    printf("      command id=%" PRId64 "\n", created_cmd_id);
    free(body); body = NULL;

    /* ── 6. Create NE-access group and wire members ──────────────────── */

    CHECK(mgt_create_ne_access_group(c, "demo-ne-grp", "demo NE access group", &body),
          "create ne-access-group");
    ne_grp_id = grab_id(body);
    printf("      ne-access-group id=%" PRId64 "\n", ne_grp_id);
    free(body); body = NULL;

    CHECK(mgt_ne_access_group_add_user(c, ne_grp_id, created_user_id),
          "add user to ne-access-group");
    CHECK(mgt_ne_access_group_add_ne(c, ne_grp_id, created_ne_id),
          "add NE to ne-access-group");

    /* ── 7. Create cmd-exec group and wire members ───────────────────── */

    CHECK(mgt_create_cmd_exec_group(c, "demo-cmd-grp", "demo cmd exec group", &body),
          "create cmd-exec-group");
    cmd_grp_id = grab_id(body);
    printf("      cmd-exec-group id=%" PRId64 "\n", cmd_grp_id);
    free(body); body = NULL;

    CHECK(mgt_cmd_exec_group_add_user(c, cmd_grp_id, created_user_id),
          "add user to cmd-exec-group");
    CHECK(mgt_cmd_exec_group_add_command(c, cmd_grp_id, created_cmd_id),
          "add command to cmd-exec-group");

    /* ── 8. Authorize check ──────────────────────────────────────────── */

    {
        authorize_decision_t d;
        CHECK(mgt_authorize_check(c, "demo_operator", created_ne_id, created_cmd_id, &d),
              "authorize check");
        printf("      authorize(demo_operator, ne=%" PRId64 ", cmd=%" PRId64 ") -> %s\n",
               created_ne_id, created_cmd_id, d.allowed ? "ALLOW" : "DENY");
        if (!d.allowed)
            printf("      reason: %s\n", d.reason);
        printf("      trace: user_exists=%d user_enabled=%d command_on_ne=%d "
               "ne_reachable=%d command_exec_allowed=%d\n",
               d.user_exists, d.user_enabled, d.command_on_ne,
               d.ne_reachable, d.command_exec_allowed);
    }

    /* ── 9. Save history (audit) ─────────────────────────────────────── */

    CHECK(mgt_save_history(c, "demo_operator", "show version", "lab01",
                           "10.0.0.1", "ne-command", "ok"),
          "save history");

    /* ── 10. List groups (raw JSON) ──────────────────────────────────── */

    if (mgt_list_ne_access_groups(c, &body) == 0) {
        print_json("ne-access-groups", body); free(body); body = NULL;
    }
    if (mgt_list_cmd_exec_groups(c, &body) == 0) {
        print_json("cmd-exec-groups", body); free(body); body = NULL;
    }

    /* ── Cleanup: delete created resources in reverse order ──────────── */

cleanup:
    printf("\n--- cleanup ---\n");

    if (cmd_grp_id > 0) {
        mgt_cmd_exec_group_remove_command(c, cmd_grp_id, created_cmd_id);
        mgt_cmd_exec_group_remove_user(c, cmd_grp_id, created_user_id);
        if (mgt_delete_cmd_exec_group(c, cmd_grp_id) == 0)
            printf("OK:   deleted cmd-exec-group %" PRId64 "\n", cmd_grp_id);
    }
    if (ne_grp_id > 0) {
        mgt_ne_access_group_remove_ne(c, ne_grp_id, created_ne_id);
        mgt_ne_access_group_remove_user(c, ne_grp_id, created_user_id);
        if (mgt_delete_ne_access_group(c, ne_grp_id) == 0)
            printf("OK:   deleted ne-access-group %" PRId64 "\n", ne_grp_id);
    }
    if (created_cmd_id > 0) {
        if (mgt_delete_command(c, created_cmd_id) == 0)
            printf("OK:   deleted command %" PRId64 "\n", created_cmd_id);
    }
    if (created_ne_id > 0) {
        if (mgt_delete_ne(c, created_ne_id) == 0)
            printf("OK:   deleted NE %" PRId64 "\n", created_ne_id);
    }
    if (created_user_id > 0) {
        if (mgt_delete_user(c, created_user_id) == 0)
            printf("OK:   deleted user %" PRId64 "\n", created_user_id);
    }

    free(body);
    mgt_client_free(c);
    return 0;
}
