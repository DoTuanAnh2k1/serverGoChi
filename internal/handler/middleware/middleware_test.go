package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/internal/config"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/testutil"
	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
)

func TestMain(m *testing.M) {
	testutil.InitTestLogger()
	config.Init(&config_models.Config{})
	os.Exit(m.Run())
}

// ── RateLimit ─────────────────────────────────────────────────────────────────

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	rl := middleware.NewRateLimiterForTest(5, time.Minute)
	handler := middleware.RateLimit(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.2.3.4:1000"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request %d: got %d, want 200", i+1, w.Code)
		}
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	rl := middleware.NewRateLimiterForTest(3, time.Minute)
	handler := middleware.RateLimit(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "5.6.7.8:2000"

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request %d should be allowed, got %d", i+1, w.Code)
		}
	}

	// 4th request should be blocked
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("4th request: got %d, want 429", w.Code)
	}
}

func TestRateLimit_DifferentIPsAreIndependent(t *testing.T) {
	rl := middleware.NewRateLimiterForTest(2, time.Minute)
	handler := middleware.RateLimit(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust limit for IP A
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:1"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	// IP A is now blocked
	reqA := httptest.NewRequest(http.MethodGet, "/", nil)
	reqA.RemoteAddr = "10.0.0.1:1"
	wA := httptest.NewRecorder()
	handler.ServeHTTP(wA, reqA)
	if wA.Code != http.StatusTooManyRequests {
		t.Errorf("IP A should be blocked, got %d", wA.Code)
	}

	// IP B should still be allowed
	reqB := httptest.NewRequest(http.MethodGet, "/", nil)
	reqB.RemoteAddr = "10.0.0.2:1"
	wB := httptest.NewRecorder()
	handler.ServeHTTP(wB, reqB)
	if wB.Code != http.StatusOK {
		t.Errorf("IP B should be allowed, got %d", wB.Code)
	}
}

// ── Authenticate middleware ───────────────────────────────────────────────────

func TestAuthenticate_MissingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", w.Code)
	}
}

func TestAuthenticate_MalformedHeader(t *testing.T) {
	cases := []string{
		"Bearer sometoken",
		"Basictoken",
		"token-only",
	}
	for _, h := range cases {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", h)

		middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("header %q: got %d, want 401", h, w.Code)
		}
	}
}

func TestAuthenticate_InvalidToken(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic not.a.valid.jwt")

	middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Error("invalid token should not pass through")
	}
}

// ── CheckRole middleware ──────────────────────────────────────────────────────

func userCtx(r *http.Request, u *middleware.User) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserContextKey, u)
	return r.WithContext(ctx)
}

func TestCheckRole_AdminAllowed(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = userCtx(req, &middleware.User{Username: "alice", Roles: "admin viewer"})

	middleware.CheckRole(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("admin user: got %d, want 200", w.Code)
	}
}

func TestCheckRole_NonAdminForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = userCtx(req, &middleware.User{Username: "bob", Roles: "viewer operator"})

	middleware.CheckRole(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("non-admin user: got %d, want 403", w.Code)
	}
}

func TestCheckRole_AdminCaseInsensitive(t *testing.T) {
	for _, role := range []string{"Admin", "ADMIN", "admin"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = userCtx(req, &middleware.User{Username: "alice", Roles: role})

		middleware.CheckRole(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("role %q: got %d, want 200", role, w.Code)
		}
	}
}

func TestCheckRole_NoUserInContext(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// no user injected into context

	middleware.CheckRole(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("missing user in context: got %d, want 500", w.Code)
	}
}

func TestCheckRole_EmptyRoles(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = userCtx(req, &middleware.User{Username: "alice", Roles: ""})

	middleware.CheckRole(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("empty roles: got %d, want 403", w.Code)
	}
}
