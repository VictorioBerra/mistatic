package api

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func RegisterHooks(app core.App) {
	// Validate Sites
	validateSite := func(e *core.RecordEvent) error {
		if e.Record.Collection().Name != "sites" {
			return e.Next()
		}

		subdomain := e.Record.GetString("subdomain")
		
		subdomain = strings.ToLower(subdomain)
		reg := regexp.MustCompile(`[^a-z0-9-]+`)
		subdomain = reg.ReplaceAllString(subdomain, "")
		subdomain = strings.Trim(subdomain, "-")
		
		if len(subdomain) < 2 {
			return fmt.Errorf("subdomain must be at least 2 characters")
		}
		
		e.Record.Set("subdomain", subdomain)

		// 2. Check conflict with custom domains
		fullDomain := subdomain + ".mistatic.local" // Assuming local dev MVP
		
		count, _ := app.CountRecords("custom_domains", dbx.HashExp{"domain": fullDomain})
		if count > 0 {
			return fmt.Errorf("subdomain conflicts with an existing custom domain")
		}
		
		count2, _ := app.CountRecords("custom_domains", dbx.HashExp{"domain": subdomain})
		if count2 > 0 {
			return fmt.Errorf("subdomain conflicts with an existing custom domain")
		}

		return e.Next()
	}

	app.OnRecordCreate("sites").BindFunc(validateSite)
	app.OnRecordUpdate("sites").BindFunc(validateSite)

	// Validate Custom Domains
	validateDomain := func(e *core.RecordEvent) error {
		if e.Record.Collection().Name != "custom_domains" {
			return e.Next()
		}

		domain := strings.ToLower(e.Record.GetString("domain"))
		e.Record.Set("domain", domain)

		stripped := strings.TrimSuffix(domain, ".mistatic.local")
		count, _ := app.CountRecords("sites", dbx.HashExp{"subdomain": stripped})
		if count > 0 {
			return fmt.Errorf("custom domain conflicts with an existing site subdomain")
		}
		
		count2, _ := app.CountRecords("sites", dbx.HashExp{"subdomain": domain})
		if count2 > 0 {
			return fmt.Errorf("custom domain conflicts with an existing site subdomain")
		}

		return e.Next()
	}

	app.OnRecordCreate("custom_domains").BindFunc(validateDomain)
	app.OnRecordUpdate("custom_domains").BindFunc(validateDomain)

	// Delete Hooks for Cleanup
	app.OnRecordAfterDeleteSuccess("deployments").BindFunc(func(e *core.RecordEvent) error {
		siteID := e.Record.GetString("site")
		deploymentID := e.Record.Id
		if siteID != "" && deploymentID != "" {
			cwd, _ := os.Getwd()
			dir := filepath.Join(cwd, "..", "sites", siteID, deploymentID)
			os.RemoveAll(dir)
		}
		return e.Next()
	})

	app.OnRecordAfterDeleteSuccess("sites").BindFunc(func(e *core.RecordEvent) error {
		siteID := e.Record.Id
		if siteID != "" {
			cwd, _ := os.Getwd()
			dir := filepath.Join(cwd, "..", "sites", siteID)
			os.RemoveAll(dir)
		}
		return e.Next()
	})
}
