package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// resetCache clears the global cache so tests don't interfere with each other.
func resetCache() {
	globalTokenCache.mu.Lock()
	defer globalTokenCache.mu.Unlock()
	globalTokenCache.tokens = make(map[string]cachedToken)
	globalTokenCache.locks = make(map[string]*sync.Mutex)
}

func newFakeBigIP(t *testing.T, loginCount *int32) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/mgmt/shared/authn/login", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(loginCount, 1)
		resp := LoginResponse{
			Token: TokenDetails{
				Token:   "fake-token-123",
				Timeout: 1200, // 20 minutes
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	return httptest.NewTLSServer(mux)
}

func TestGetToken_CachesAcrossCalls(t *testing.T) {
	resetCache()

	var loginCount int32
	srv := newFakeBigIP(t, &loginCount)
	defer srv.Close()

	tok, err := GetToken(srv.URL, "user", "pass", true, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.Token != "fake-token-123" {
		t.Fatalf("unexpected token: %q", tok.Token)
	}

	for i := 0; i < 5; i++ {
		if _, err := GetToken(srv.URL, "user", "pass", true, 30); err != nil {
			t.Fatalf("unexpected error on call %d: %v", i, err)
		}
	}

	if got := atomic.LoadInt32(&loginCount); got != 1 {
		t.Errorf("expected exactly 1 login, got %d", got)
	}
}

func TestGetToken_ConcurrentSingleLogin(t *testing.T) {
	resetCache()

	var loginCount int32
	srv := newFakeBigIP(t, &loginCount)
	defer srv.Close()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if _, err := GetToken(srv.URL, "user", "pass", true, 30); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&loginCount); got != 1 {
		t.Errorf("expected exactly 1 login under concurrency, got %d", got)
	}
}

func TestGetToken_RefetchesAfterExpiry(t *testing.T) {
	resetCache()

	var loginCount int32
	srv := newFakeBigIP(t, &loginCount)
	defer srv.Close()

	if _, err := GetToken(srv.URL, "user", "pass", true, 30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	globalTokenCache.mu.Lock()
	c := globalTokenCache.tokens[srv.URL]
	c.expiresAt = time.Now().Add(-1 * time.Minute)
	globalTokenCache.tokens[srv.URL] = c
	globalTokenCache.mu.Unlock()

	if _, err := GetToken(srv.URL, "user", "pass", true, 30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := atomic.LoadInt32(&loginCount); got != 2 {
		t.Errorf("expected 2 logins (initial + after expiry), got %d", got)
	}
}

func TestTokenCache_SetUsesFallbackTTL(t *testing.T) {
	resetCache()

	globalTokenCache.set("https://example.test", TokenDetails{
		Token:   "x",
		Timeout: 0,
	})

	if _, ok := globalTokenCache.get("https://example.test"); !ok {
		t.Error("expected token with zero timeout to be cached via fallback TTL")
	}
}
