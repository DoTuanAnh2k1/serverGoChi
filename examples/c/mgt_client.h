/* mgt_client.h — C client for cli-mgt v2.
 *
 * Thin wrapper over libcurl + a tiny hand-written JSON parser (no cJSON
 * dependency). Link with -lcurl. Designed for the cli-gate / ne-command
 * proxy side of the product, which is plain C.
 *
 * All functions return 0 on success, nonzero on failure. Output strings
 * returned through char** parameters are owned by the caller and must be
 * free()'d. The authorize_decision_t struct has no owned strings — copy
 * .reason out yourself if you need it past the caller's stack. */

#ifndef MGT_CLIENT_H
#define MGT_CLIENT_H

#include <stdint.h>

typedef struct mgt_client mgt_client_t;

/* Allocate a client. base_url is copied. Call mgt_client_free() when done. */
mgt_client_t *mgt_client_new(const char *base_url);
void          mgt_client_free(mgt_client_t *c);

/* The token returned by authenticate is stored on the client and attached
 * automatically to subsequent authenticated calls as the Authorization
 * header. The role is also stored and retrievable via mgt_get_role(). */
int         mgt_authenticate(mgt_client_t *c, const char *username, const char *password);
void        mgt_set_token(mgt_client_t *c, const char *token);
const char *mgt_get_role(mgt_client_t *c);

/* Raw helpers — return the response body in *out_body (caller frees).
 * *out_http_status is populated with the HTTP status code. body may be NULL
 * for GET/DELETE. */
int mgt_request(mgt_client_t *c,
                const char *method,
                const char *path,
                const char *body,
                char **out_body,
                long  *out_http_status);

/* Convenience wrappers — return the raw JSON body. Caller frees *out. */
int mgt_get   (mgt_client_t *c, const char *path, char **out);
int mgt_post  (mgt_client_t *c, const char *path, const char *json_body, char **out);
int mgt_put   (mgt_client_t *c, const char *path, const char *json_body, char **out);
int mgt_delete(mgt_client_t *c, const char *path);

/* ── Domain structs ──────────────────────────────────────────────────── */

/* The one question. command_on_ne / ne_reachable / etc. mirror the trace
 * flags returned by POST /aa/authorize/check. */
typedef struct {
    int  allowed;
    int  user_exists;
    int  user_enabled;
    int  ne_reachable;
    int  command_on_ne;
    int  command_exec_allowed;
    char reason[256];
} authorize_decision_t;

typedef struct {
    int64_t id;
    char    username[64];
    char    email[128];
    char    full_name[128];
    char    phone[32];
    char    role[16];       /* "super_admin", "admin", "user" */
    int     is_enabled;
    char    locked_at[32];  /* ISO timestamp or empty */
    char    last_login_at[32];
} mgt_user_t;

typedef struct {
    int64_t id;
    char    namespace_[64];  /* trailing _ to avoid C keyword */
    char    ne_type[32];
    char    site_name[64];
    char    description[256];
    char    master_ip[48];
    int     master_port;
    char    ssh_username[64];
    char    ssh_password[64];
    char    command_url[256];
    char    conf_mode[16];   /* SSH, TELNET, NETCONF, RESTCONF */
} mgt_ne_t;

typedef struct {
    int64_t id;
    int64_t ne_id;
    char    service[32];     /* ne-config, ne-command */
    char    cmd_text[1024];
    char    description[256];
} mgt_command_t;

typedef struct {
    int64_t id;
    char    name[64];
    char    description[256];
} mgt_group_t;

typedef struct {
    int64_t id;
    char    list_type[16];   /* blacklist, whitelist */
    char    match_type[16];  /* username, ip_cidr, email_domain */
    char    pattern[128];
    char    reason[256];
} mgt_access_entry_t;

/* ── Authorize & history ─────────────────────────────────────────────── */

int mgt_authorize_check(mgt_client_t *c,
                        const char *username,
                        int64_t ne_id,
                        int64_t command_id,
                        authorize_decision_t *out);

/* /aa/history/save — unauthenticated, used by the cli-gate proxy to push
 * audit events without forwarding the operator's JWT. account is supplied
 * by the caller. */
int mgt_save_history(mgt_client_t *c,
                     const char *account,
                     const char *cmd_text,
                     const char *ne_namespace,
                     const char *ne_ip,
                     const char *scope,
                     const char *result);

/* ── User management ─────────────────────────────────────────────────── */

int mgt_list_users(mgt_client_t *c, char **out_json);
int mgt_create_user(mgt_client_t *c, const char *username, const char *password,
                    const char *email, const char *full_name, const char *role,
                    char **out_json);
int mgt_update_user(mgt_client_t *c, int64_t id, const char *json_patch);
int mgt_delete_user(mgt_client_t *c, int64_t id);
int mgt_reset_password(mgt_client_t *c, int64_t user_id, const char *new_password);

/* ── NE management ───────────────────────────────────────────────────── */

int mgt_list_nes(mgt_client_t *c, char **out_json);
int mgt_create_ne(mgt_client_t *c, const char *json_body, char **out_json);
int mgt_update_ne(mgt_client_t *c, int64_t id, const char *json_patch);
int mgt_delete_ne(mgt_client_t *c, int64_t id);

/* ── Command management ──────────────────────────────────────────────── */

int mgt_list_commands(mgt_client_t *c, char **out_json);
int mgt_create_command(mgt_client_t *c, int64_t ne_id, const char *service,
                       const char *cmd_text, const char *description, char **out_json);
int mgt_update_command(mgt_client_t *c, int64_t id, const char *json_patch);
int mgt_delete_command(mgt_client_t *c, int64_t id);

/* ── User permission lookups ─────────────────────────────────────────── */

int mgt_list_user_executable_commands(mgt_client_t *c, int64_t user_id, char **out_json);
int mgt_list_user_accessible_nes(mgt_client_t *c, int64_t user_id, char **out_json);
int mgt_list_ne_authorized_users(mgt_client_t *c, int64_t ne_id, char **out_json);
int mgt_list_command_authorized_users(mgt_client_t *c, int64_t command_id, char **out_json);

/* ── NE access groups ────────────────────────────────────────────────── */

int mgt_list_ne_access_groups(mgt_client_t *c, char **out_json);
int mgt_create_ne_access_group(mgt_client_t *c, const char *name, const char *desc, char **out_json);
int mgt_update_ne_access_group(mgt_client_t *c, int64_t id, const char *name, const char *desc);
int mgt_delete_ne_access_group(mgt_client_t *c, int64_t id);
int mgt_ne_access_group_add_user(mgt_client_t *c, int64_t group_id, int64_t user_id);
int mgt_ne_access_group_remove_user(mgt_client_t *c, int64_t group_id, int64_t user_id);
int mgt_ne_access_group_add_ne(mgt_client_t *c, int64_t group_id, int64_t ne_id);
int mgt_ne_access_group_remove_ne(mgt_client_t *c, int64_t group_id, int64_t ne_id);

/* ── Command exec groups ─────────────────────────────────────────────── */

int mgt_list_cmd_exec_groups(mgt_client_t *c, char **out_json);
int mgt_create_cmd_exec_group(mgt_client_t *c, const char *name, const char *desc, char **out_json);
int mgt_update_cmd_exec_group(mgt_client_t *c, int64_t id, const char *name, const char *desc);
int mgt_delete_cmd_exec_group(mgt_client_t *c, int64_t id);
int mgt_cmd_exec_group_add_user(mgt_client_t *c, int64_t group_id, int64_t user_id);
int mgt_cmd_exec_group_remove_user(mgt_client_t *c, int64_t group_id, int64_t user_id);
int mgt_cmd_exec_group_add_command(mgt_client_t *c, int64_t group_id, int64_t command_id);
int mgt_cmd_exec_group_remove_command(mgt_client_t *c, int64_t group_id, int64_t command_id);

/* ── Policy & access lists ───────────────────────────────────────────── */

int mgt_get_password_policy(mgt_client_t *c, char **out_json);
int mgt_upsert_password_policy(mgt_client_t *c, const char *json_body);
int mgt_list_access_list(mgt_client_t *c, const char *list_type, char **out_json);
int mgt_create_access_entry(mgt_client_t *c, const char *list_type, const char *match_type,
                            const char *pattern, const char *reason, char **out_json);
int mgt_delete_access_entry(mgt_client_t *c, int64_t id);

#endif /* MGT_CLIENT_H */
