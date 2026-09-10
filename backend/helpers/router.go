package helpers

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"net/url"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func ServeStaticSite(se *core.ServeEvent, distDirFS fs.FS) {
	appDomain := os.Getenv("APP_DOMAIN")
	if appDomain == "" {
		appDomain = "app.mistatic.local"
	}
	rootDomain := os.Getenv("ROOT_DOMAIN")
	if rootDomain == "" {
		rootDomain = "mistatic.local"
	}

	dashboardHandler := apis.Static(distDirFS, true)

	// SSO Callback POST handler
	se.Router.POST("/xapi/sso-callback", func(e *core.RequestEvent) error {
		host := e.Request.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}
		
		token := e.Request.FormValue("token")
		if token == "" {
			return e.String(http.StatusBadRequest, "Missing token")
		}

		var siteRecord *core.Record
		var err error
		siteParam := e.Request.URL.Query().Get("site")
		if siteParam != "" {
			siteRecord, err = e.App.FindRecordById("sites", siteParam)
		} else {
			siteRecord, err = resolveHostToSite(e.App, host, rootDomain)
		}

		if err != nil {
			return e.String(http.StatusNotFound, "Site not found")
		}

		authRecord, err := e.App.FindAuthRecordByToken(token, core.TokenTypeAuth)
		if err != nil || authRecord.Id != siteRecord.GetString("user") {
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
		
		host := e.Request.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}

		var siteRecord *core.Record
		var isSubpathSite bool
		var subpathPrefix string

		reqPath := e.Request.URL.Path

		if host == appDomain || host == "localhost" || host == "127.0.0.1" {
			if strings.HasPrefix(reqPath, "/site/") {
				parts := strings.SplitN(reqPath, "/", 4)
				if len(parts) >= 3 {
					subdomain := parts[2]
					site, err := e.App.FindFirstRecordByFilter("sites", "subdomain = {:sub}", dbx.Params{"sub": subdomain})
					if err == nil {
						siteRecord = site
						isSubpathSite = true
						subpathPrefix = "/site/" + subdomain

						if reqPath == subpathPrefix {
							return e.Redirect(http.StatusMovedPermanently, subpathPrefix+"/")
						}
					}
				}
			}

			if siteRecord == nil {
				return dashboardHandler(e)
			}
		}

		if siteRecord == nil {
			var err error
			siteRecord, err = resolveHostToSite(e.App, host, rootDomain)
			if err != nil {
				return e.String(http.StatusNotFound, "Site not found or inactive")
			}
		}
		
		// Fire and forget logging
		defer func() {
			if siteRecord != nil {
				// Check if logging is disabled for this site
				if !siteRecord.GetBool("disable_request_logging") {
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
						SiteID:     siteRecord.Id,
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
			}
		}()

		siteID := siteRecord.Id
		spaFallback := siteRecord.GetBool("spa_fallback")
		isPublic := siteRecord.GetBool("is_public")

		if !isPublic {
			// Check cookie
			cookie, err := e.Request.Cookie("mistatic_auth")
			
			needsAuth := false
			if err != nil || cookie.Value == "" {
				needsAuth = true
			} else {
				authRecord, authErr := e.App.FindAuthRecordByToken(cookie.Value, core.TokenTypeAuth)
				if authErr != nil || authRecord.Id != siteRecord.GetString("user") {
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

		deploymentID, err := getActiveDeployment(e.App, siteID)
		if err != nil {
			return e.HTML(http.StatusOK, defaultSiteHTML(host))
		}

		cwd, _ := os.Getwd()
		siteDir := filepath.Join(cwd, "..", "sites", siteID, deploymentID)
		
		if isSubpathSite {
			reqPath = strings.TrimPrefix(reqPath, subpathPrefix)
			if !strings.HasPrefix(reqPath, "/") {
				reqPath = "/" + reqPath
			}
		}

		if reqPath == "/" {
			reqPath = "/index.html"
		}
		
		fullPath := filepath.Join(siteDir, reqPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) && spaFallback {
			fullPath = filepath.Join(siteDir, "index.html")
		}

		http.ServeFile(e.Response, e.Request, fullPath)
		return nil
	})
}

func resolveHostToSite(app core.App, host string, rootDomain string) (*core.Record, error) {
	if strings.HasSuffix(host, "."+rootDomain) {
		subdomain := strings.TrimSuffix(host, "."+rootDomain)
		record, err := app.FindFirstRecordByFilter("sites", "subdomain = {:sub}", dbx.Params{"sub": subdomain})
		if err != nil {
			return nil, err
		}
		return record, nil
	}

	domainRecord, err := app.FindFirstRecordByFilter("custom_domains", "domain = {:domain}", dbx.Params{"domain": host})
	if err != nil {
		return nil, err
	}
	
	siteID := domainRecord.GetString("site")
	siteRecord, err := app.FindRecordById("sites", siteID)
	if err != nil {
		return nil, err
	}

	return siteRecord, nil
}

func getActiveDeployment(app core.App, siteID string) (string, error) {
	siteRecord, err := app.FindRecordById("sites", siteID)
	if err != nil {
		return "", err
	}
	
	deploymentID := siteRecord.GetString("active_deployment")
	if deploymentID == "" {
		return "", fmt.Errorf("no active deployment")
	}
	
	return deploymentID, nil
}
