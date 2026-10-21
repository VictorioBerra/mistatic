package helpers

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
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

func TestCleanIndexDocument(t *testing.T) {
	tests := map[string]string{
		"":                   "index.html",
		".":                  "index.html",
		"   ":                "index.html",
		"index.html":         "index.html",
		"home.html":          "home.html",
		"/home.html":         "home.html",
		"docs/index.html":    "docs/index.html",
		"/nested/app.html":   "nested/app.html",
		"../secret.txt":      "index.html",
		"../../etc/passwd":   "index.html",
		"docs/../index.html": "index.html",
	}

	for input, expected := range tests {
		actual := cleanIndexDocument(input)
		if actual != expected {
			t.Errorf("cleanIndexDocument(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestRouteFromRecordIndexDocument(t *testing.T) {
	sitesCollection := core.NewBaseCollection("sites")
	sitesCollection.Fields.Add(&core.TextField{Name: "index_document"})

	// Record with index_document unset
	recordDefault := core.NewRecord(sitesCollection)
	recordDefault.Id = "site-default"
	routeDefault := routeFromRecord(recordDefault)
	if routeDefault.indexDocument != "" {
		t.Fatalf("expected raw indexDocument to be empty, got %q", routeDefault.indexDocument)
	}
	if clean := cleanIndexDocument(routeDefault.indexDocument); clean != "index.html" {
		t.Fatalf("expected cleaned indexDocument to default to index.html, got %q", clean)
	}

	// Record with custom index_document
	recordCustom := core.NewRecord(sitesCollection)
	recordCustom.Id = "site-custom"
	recordCustom.Set("index_document", "home.html")
	routeCustom := routeFromRecord(recordCustom)
	if routeCustom.indexDocument != "home.html" {
		t.Fatalf("expected raw indexDocument to be home.html, got %q", routeCustom.indexDocument)
	}
	if clean := cleanIndexDocument(routeCustom.indexDocument); clean != "home.html" {
		t.Fatalf("expected cleaned indexDocument to be home.html, got %q", clean)
	}
}

func TestResolveFilePath(t *testing.T) {
	tempDir := t.TempDir()

	// Setup files
	_ = os.WriteFile(filepath.Join(tempDir, "index.html"), []byte("index content"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "home.html"), []byte("home content"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "about.html"), []byte("about content"), 0644)
	_ = os.MkdirAll(filepath.Join(tempDir, "docs"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "docs", "index.html"), []byte("docs index"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "docs", "home.html"), []byte("docs home"), 0644)

	tests := []struct {
		name        string
		reqPath     string
		indexDoc    string
		spaFallback bool
		expectedRel string
	}{
		{
			name:        "Root with default index.html",
			reqPath:     "/",
			indexDoc:    "",
			spaFallback: false,
			expectedRel: "index.html",
		},
		{
			name:        "Root with custom index home.html",
			reqPath:     "/",
			indexDoc:    "home.html",
			spaFallback: false,
			expectedRel: "home.html",
		},
		{
			name:        "Direct file request",
			reqPath:     "/about.html",
			indexDoc:    "home.html",
			spaFallback: false,
			expectedRel: "about.html",
		},
		{
			name:        "Directory request with custom index",
			reqPath:     "/docs/",
			indexDoc:    "home.html",
			spaFallback: false,
			expectedRel: filepath.Join("docs", "home.html"),
		},
		{
			name:        "Missing file with SPA fallback to custom index",
			reqPath:     "/unknown/route",
			indexDoc:    "home.html",
			spaFallback: true,
			expectedRel: "home.html",
		},
		{
			name:        "Missing file without SPA fallback remains non-existent target",
			reqPath:     "/unknown/route",
			indexDoc:    "home.html",
			spaFallback: false,
			expectedRel: filepath.Join("unknown", "route"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := resolveFilePath(tempDir, tc.reqPath, tc.indexDoc, tc.spaFallback)
			expected := filepath.Join(tempDir, tc.expectedRel)
			if actual != expected {
				t.Errorf("resolveFilePath() = %q, want %q", actual, expected)
			}
		})
	}
}

