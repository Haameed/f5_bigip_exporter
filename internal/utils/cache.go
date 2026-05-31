package utils

import (
	"sync"
	"time"
)

const tokenSafetyMargin = 30 * time.Second

const fallbackTTL = 5 * time.Minute

type cachedToken struct {
	token     TokenDetails
	expiresAt time.Time
}

type tokenCache struct {
	mu     sync.Mutex
	tokens map[string]cachedToken
	locks  map[string]*sync.Mutex
}

var globalTokenCache = &tokenCache{
	tokens: make(map[string]cachedToken),
	locks:  make(map[string]*sync.Mutex),
}

func (tc *tokenCache) targetLock(target string) *sync.Mutex {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	l, ok := tc.locks[target]
	if !ok {
		l = &sync.Mutex{}
		tc.locks[target] = l
	}
	return l
}

func (tc *tokenCache) get(target string) (TokenDetails, bool) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	c, ok := tc.tokens[target]
	if !ok || time.Now().After(c.expiresAt) {
		return TokenDetails{}, false
	}
	return c.token, true
}

func (tc *tokenCache) set(target string, token TokenDetails) {
	ttl := time.Duration(token.Timeout) * time.Second
	if ttl <= tokenSafetyMargin {
		ttl = fallbackTTL
	}

	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tokens[target] = cachedToken{
		token:     token,
		expiresAt: time.Now().Add(ttl - tokenSafetyMargin),
	}
}

func GetToken(url, username, password string, insecure bool, timeout time.Duration) (TokenDetails, error) {
	if token, ok := globalTokenCache.get(url); ok {
		return token, nil
	}
	lock := globalTokenCache.targetLock(url)
	lock.Lock()
	defer lock.Unlock()

	if token, ok := globalTokenCache.get(url); ok {
		return token, nil
	}

	token, err := GetTokenFromF5(url, username, password, insecure, timeout)
	if err != nil {
		return TokenDetails{}, err
	}

	globalTokenCache.set(url, token)
	return token, nil
}
