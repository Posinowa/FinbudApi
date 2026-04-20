package blacklist

import "sync"

var store sync.Map

// Add token jti'sini blacklist'e ekler.
func Add(jti string) {
	store.Store(jti, struct{}{})
}

// IsBlacklisted token jti'sinin blacklist'te olup olmadığını kontrol eder.
func IsBlacklisted(jti string) bool {
	_, exists := store.Load(jti)
	return exists
}
