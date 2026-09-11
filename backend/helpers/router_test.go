package helpers

import (
	"testing"
	"time"
)

func TestNormalizeHost(t *testing.T) {
	tests := map[string]string{
		"Example.COM":         "example.com",
		"example.com.":        "example.com",
		"example.com:8090":    "example.com",
		"[2001:db8::1]:8090":  "2001:db8::1",
		"[2001:db8::1]":       "2001:db8::1",
		"sub.example.com:443": "sub.example.com",
	}

	for input, expected := range tests {
		if actual := normalizeHost(input); actual != expected {
			t.Errorf("normalizeHost(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestSiteRouteCacheInvalidationGeneration(t *testing.T) {
	clearSiteRouteCache()
	siteRoutes.RLock()
	generation := siteRoutes.generation
	siteRoutes.RUnlock()

	route := siteRoute{siteID: "site-a", deploymentID: "deployment-a", isPublic: true}
	cacheSiteRoute("a.example.com", route, generation)
	if cached, ok := getCachedSiteRoute("a.example.com"); !ok || cached != route {
		t.Fatalf("expected cached route, got %#v, %v", cached, ok)
	}

	clearSiteRouteCache()
	cacheSiteRoute("a.example.com", route, generation)
	if _, ok := getCachedSiteRoute("a.example.com"); ok {
		t.Fatal("stale lookup repopulated cache after invalidation")
	}
}

func TestSiteRouteCacheExpiry(t *testing.T) {
	clearSiteRouteCache()
	siteRoutes.Lock()
	siteRoutes.entries["expired.example.com"] = siteRouteCacheEntry{
		route:     siteRoute{siteID: "expired"},
		expiresAt: time.Now().Add(-time.Second),
	}
	siteRoutes.Unlock()

	if _, ok := getCachedSiteRoute("expired.example.com"); ok {
		t.Fatal("expired route returned from cache")
	}
}

func BenchmarkGetCachedSiteRoute(b *testing.B) {
	clearSiteRouteCache()
	siteRoutes.RLock()
	generation := siteRoutes.generation
	siteRoutes.RUnlock()
	cacheSiteRoute("bench.example.com", siteRoute{siteID: "site-a"}, generation)

	b.ReportAllocs()
	for b.Loop() {
		if _, ok := getCachedSiteRoute("bench.example.com"); !ok {
			b.Fatal("cache miss")
		}
	}
}
