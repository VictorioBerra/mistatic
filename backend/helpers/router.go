package helpers

import (
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const siteRouteCacheTTL = 5 * time.Minute

type siteRoute struct {
	siteID                string
	userID                string
	deploymentID          string
	spaFallback           bool
	isPublic              bool
	disableRequestLogging bool
}

type siteRouteCacheEntry struct {
	route     siteRoute
	expiresAt time.Time
}

var siteRoutes = struct {
	sync.RWMutex
	entries    map[string]siteRouteCacheEntry
	generation uint64
}{
	entries: make(map[string]siteRouteCacheEntry),
}

func RegisterSiteRouteCacheHooks(app core.App) {
	invalidate := func(e *core.RecordEvent) error {
		err := e.Next()
		if err == nil {
			clearSiteRouteCache()
		}
		return err
	}

	app.OnRecordAfterCreateSuccess("sites", "custom_domains").BindFunc(invalidate)
	app.OnRecordAfterUpdateSuccess("sites", "custom_domains").BindFunc(invalidate)
	app.OnRecordAfterDeleteSuccess("sites", "custom_domains").BindFunc(invalidate)
}

func clearSiteRouteCache() {
	siteRoutes.Lock()
	siteRoutes.entries = make(map[string]siteRouteCacheEntry)
	siteRoutes.generation++
	siteRoutes.Unlock()
}

func ServeStaticSite(se *core.ServeEvent, distDirFS fs.FS) {
	appDomain := normalizeHost(os.Getenv("APP_DOMAIN"))
	if appDomain == "" {
		appDomain = "app.mistatic.local"
	}
	rootDomain := normalizeHost(os.Getenv("ROOT_DOMAIN"))
	if rootDomain == "" {
		rootDomain = "mistatic.local"
	}
	cwd, _ := os.Getwd()
	sitesRoot := filepath.Clean(filepath.Join(cwd, "..", "sites"))

	dashboardHandler := apis.Static(distDirFS, true)

	// SSO Callback POST handler
	se.Router.POST("/xapi/sso-callback", func(e *core.RequestEvent) error {
		host := normalizeHost(e.Request.Host)

		token := e.Request.FormValue("token")
		if token == "" {
			return e.String(http.StatusBadRequest, "Missing token")
		}

		var userID string
		var err error
		siteParam := e.Request.URL.Query().Get("site")
		if siteParam != "" {
			var siteRecord *core.Record
			siteRecord, err = e.App.FindRecordById("sites", siteParam)
			if err == nil {
				userID = siteRecord.GetString("user")
			}
		} else {
			var route siteRoute
			route, err = resolveHostToSite(e.App, host, rootDomain)
			if err == nil {
				userID = route.userID
			}
		}

		if err != nil {
			return e.String(http.StatusNotFound, "Site not found")
		}

		authRecord, err := e.App.FindAuthRecordByToken(token, core.TokenTypeAuth)
		if err != nil || authRecord.Id != userID {
			return e.String(http.StatusForbidden, "Invalid token or you do not own this site")
		}

		cookie := new(http.Cookie)
		cookie.Name = "mistatic_auth"
		cookie.Value = token
		cookie.Path = "/"
		cookie.HttpOnly = true

		// In dev we don't force secure cookie unless HTTPS
		if e.Request.TLS != nil || e.Request.Header.Get("X-Forwarded-Proto") == "https" {
			cookie.Secure = true
			cookie.SameSite = http.SameSiteNoneMode
		}

		http.SetCookie(e.Response, cookie)

		returnTo := e.Request.URL.Query().Get("returnTo")
		if returnTo == "" {
			returnTo = "/"
		}
		return e.Redirect(http.StatusFound, returnTo)
	})

	se.Router.GET("/{path...}", func(e *core.RequestEvent) error {
		start := time.Now()

		host := normalizeHost(e.Request.Host)

		var route siteRoute
		var routeFound bool
		var isSubpathSite bool
		var subpathPrefix string

		reqPath := e.Request.URL.Path

		if host == appDomain || host == "localhost" || host == "127.0.0.1" {
			if strings.HasPrefix(reqPath, "/site/") {
				parts := strings.SplitN(reqPath, "/", 4)
				if len(parts) >= 3 {
					subdomain := parts[2]
					site, err := resolveHostToSite(e.App, subdomain+"."+rootDomain, rootDomain)
					if err == nil {
						route = site
						routeFound = true
						isSubpathSite = true
						subpathPrefix = "/site/" + subdomain

						if reqPath == subpathPrefix {
							return e.Redirect(http.StatusMovedPermanently, subpathPrefix+"/")
						}
					}
				}
			}

			if !routeFound {
				return dashboardHandler(e)
			}
		}

		if !routeFound {
			var err error
			route, err = resolveHostToSite(e.App, host, rootDomain)
			if err != nil {
				return e.String(http.StatusNotFound, "Site not found or inactive")
			}
			routeFound = true
		}

		// Fire and forget logging
		defer func() {
			// Check if logging is disabled for this site
			if !route.disableRequestLogging {
				duration := time.Since(start).Milliseconds()

				// Determine real IP
				ip := e.Request.Header.Get("CF-Connecting-IP")
				if ip == "" {
					ip = e.Request.Header.Get("X-Forwarded-For")
					if ip != "" {
						// X-Forwarded-For can contain multiple IPs, the first one is the client
						ip = strings.Split(ip, ",")[0]
						ip = strings.TrimSpace(ip)
					}
				}
				if ip == "" {
					ip = e.Request.Header.Get("X-Real-IP")
				}
				if ip == "" {
					ip = e.Request.RemoteAddr
				}

				payload := LogPayload{
					App:        e.App,
					SiteID:     route.siteID,
					Method:     e.Request.Method,
					Path:       reqPath,
					IP:         ip,
					UserAgent:  e.Request.UserAgent(),
					DurationMs: duration,
				}

				select {
				case LogQueue <- payload:
					// Successfully queued
				default:
					// Queue is full, drop the log to prevent blocking
				}
			}
		}()

		siteID := route.siteID
		spaFallback := route.spaFallback
		isPublic := route.isPublic

		if !isPublic {
			// Check cookie
			cookie, err := e.Request.Cookie("mistatic_auth")

			needsAuth := false
			if err != nil || cookie.Value == "" {
				needsAuth = true
			} else {
				authRecord, authErr := e.App.FindAuthRecordByToken(cookie.Value, core.TokenTypeAuth)
				if authErr != nil || authRecord.Id != route.userID {
					needsAuth = true
				}
			}

			if needsAuth {
				scheme := "http://"
				if e.Request.TLS != nil || e.Request.Header.Get("X-Forwarded-Proto") == "https" {
					scheme = "https://"
				}

				redirectUrl := scheme + e.Request.Host + "/xapi/sso-callback?site=" + siteID + "&returnTo=" + url.QueryEscape(reqPath)

				dashboardURL := os.Getenv("SSO_DASHBOARD_URL")
				if dashboardURL == "" {
					appHost := appDomain
					if strings.Contains(e.Request.Host, ":") {
						appHost = appDomain + ":" + strings.Split(e.Request.Host, ":")[1]
					}
					dashboardURL = scheme + appHost
				}

				return e.Redirect(http.StatusFound, dashboardURL+"/sso?redirect="+url.QueryEscape(redirectUrl))
			}
		}

		deploymentID := route.deploymentID
		if deploymentID == "" {
			return e.HTML(http.StatusOK, defaultSiteHTML(host))
		}

		siteDir := filepath.Join(sitesRoot, siteID, deploymentID)

		if isSubpathSite {
			reqPath = strings.TrimPrefix(reqPath, subpathPrefix)
			if !strings.HasPrefix(reqPath, "/") {
				reqPath = "/" + reqPath
			}
		}

		relativePath := strings.TrimPrefix(path.Clean("/"+reqPath), "/")
		if relativePath == "" || relativePath == "." {
			relativePath = "index.html"
		}

		fullPath := filepath.Join(siteDir, filepath.FromSlash(relativePath))
		if spaFallback {
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				fullPath = filepath.Join(siteDir, "index.html")
			}
		}

		http.ServeFile(e.Response, e.Request, fullPath)
		return nil
	})
}

func resolveHostToSite(app core.App, host string, rootDomain string) (siteRoute, error) {
	host = normalizeHost(host)
	rootDomain = normalizeHost(rootDomain)

	if route, ok := getCachedSiteRoute(host); ok {
		return route, nil
	}

	siteRoutes.RLock()
	generation := siteRoutes.generation
	siteRoutes.RUnlock()

	var record *core.Record
	var err error
	if strings.HasSuffix(host, "."+rootDomain) {
		subdomain := strings.TrimSuffix(host, "."+rootDomain)
		record, err = app.FindFirstRecordByFilter("sites", "subdomain = {:sub}", dbx.Params{"sub": subdomain})
	} else {
		var domainRecord *core.Record
		domainRecord, err = app.FindFirstRecordByFilter("custom_domains", "domain = {:domain}", dbx.Params{"domain": host})
		if err == nil {
			record, err = app.FindRecordById("sites", domainRecord.GetString("site"))
		}
	}
	if err != nil {
		return siteRoute{}, err
	}

	route := routeFromRecord(record)
	cacheSiteRoute(host, route, generation)
	return route, nil
}

func routeFromRecord(record *core.Record) siteRoute {
	return siteRoute{
		siteID:                record.Id,
		userID:                record.GetString("user"),
		deploymentID:          record.GetString("active_deployment"),
		spaFallback:           record.GetBool("spa_fallback"),
		isPublic:              record.GetBool("is_public"),
		disableRequestLogging: record.GetBool("disable_request_logging"),
	}
}

func getCachedSiteRoute(host string) (siteRoute, bool) {
	siteRoutes.RLock()
	entry, ok := siteRoutes.entries[host]
	siteRoutes.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		return siteRoute{}, false
	}
	return entry.route, true
}

func cacheSiteRoute(host string, route siteRoute, generation uint64) {
	siteRoutes.Lock()
	defer siteRoutes.Unlock()
	if generation != siteRoutes.generation {
		return
	}
	if len(siteRoutes.entries) >= 4096 {
		siteRoutes.entries = make(map[string]siteRouteCacheEntry)
	}
	siteRoutes.entries[host] = siteRouteCacheEntry{
		route:     route,
		expiresAt: time.Now().Add(siteRouteCacheTTL),
	}
}

func normalizeHost(hostPort string) string {
	host := hostPort
	if strings.HasPrefix(hostPort, "[") {
		if bracket := strings.IndexByte(hostPort, ']'); bracket > 0 {
			host = hostPort[1:bracket]
		}
	} else if colon := strings.IndexByte(hostPort, ':'); colon >= 0 && colon == strings.LastIndexByte(hostPort, ':') {
		host = hostPort[:colon]
	}
	return strings.ToLower(strings.TrimSuffix(host, "."))
}
