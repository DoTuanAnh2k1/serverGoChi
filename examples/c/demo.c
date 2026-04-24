/* demo.c — end-to-end usage of mgt_client against a running cli-mgt v2.
 *
 * Build:
 *   gcc -Wall -O2 -o demo mgt_client.c demo.c -lcurl
 *
 * Run:
 *   MGT_BASE=http://localhost:3000 MGT_USER=admin MGT_PASS=admin ./demo
 *
 * The demo logs in, prints a short summary, then probes authorize against
 * the first (user, ne, command) triple it can find. Read mgt_client.h for
 * the full API surface. */

#include "mgt_client.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static const char *env_or(const char *key, const char *fallback) {
    const char *v = getenv(key);
    return (v && *v) ? v : fallback;
}

int main(void) {
    const char *base = env_or("MGT_BASE", "http://localhost:3000");
    const char *user = env_or("MGT_USER", "admin");
    const char *pass = env_or("MGT_PASS", "admin");

    mgt_client_t *c = mgt_client_new(base);
    if (!c) { fprintf(stderr, "mgt_client_new failed\n"); return 1; }

    if (mgt_authenticate(c, user, pass) != 0) {
        fprintf(stderr, "login failed\n");
        mgt_client_free(c);
        return 1;
    }
    printf("logged in as %s\n", user);

    /* GET /aa/users → print the raw JSON (the demo stays dependency-free —
     * a real caller would parse with cJSON / jansson / json-c). */
    char *body = NULL;
    if (mgt_get(c, "/aa/users", &body) == 0) {
        printf("users: %.300s%s\n", body, strlen(body) > 300 ? " ..." : "");
        free(body); body = NULL;
    }
    if (mgt_get(c, "/aa/nes", &body) == 0) {
        printf("nes:   %.300s%s\n", body, strlen(body) > 300 ? " ..." : "");
        free(body); body = NULL;
    }
    if (mgt_get(c, "/aa/commands", &body) == 0) {
        printf("cmds:  %.300s%s\n", body, strlen(body) > 300 ? " ..." : "");
        free(body); body = NULL;
    }

    /* Authorize probe — use the seed admin user + whatever ne/command happen
     * to exist (IDs 1 are a best-guess; the check will cleanly report the
     * first failure in its trace flags). */
    authorize_decision_t d;
    if (mgt_authorize_check(c, user, 1, 1, &d) == 0) {
        printf("authorize(%s, ne=1, cmd=1) → %s\n",
               user, d.allowed ? "ALLOW" : "DENY");
        if (!d.allowed) printf("  reason: %s\n", d.reason);
        printf("  trace: user_exists=%d user_enabled=%d command_on_ne=%d "
               "ne_reachable=%d command_exec_allowed=%d\n",
               d.user_exists, d.user_enabled, d.command_on_ne,
               d.ne_reachable, d.command_exec_allowed);
    }

    /* Unauth audit push — exactly what cli-gate / the ne-command proxy does
     * after running an operator command. No JWT needed. */
    if (mgt_save_history(c,
            user, "show version", "htsmf01", "10.0.0.1",
            "ne-command", "ok") == 0) {
        printf("audit pushed\n");
    }

    mgt_client_free(c);
    return 0;
}
