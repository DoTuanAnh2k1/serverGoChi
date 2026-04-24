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
 * header. */
int  mgt_authenticate(mgt_client_t *c, const char *username, const char *password);
void mgt_set_token(mgt_client_t *c, const char *token);

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

#endif /* MGT_CLIENT_H */
