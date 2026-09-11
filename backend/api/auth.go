package api

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

func findOwnedSite(e *core.RequestEvent, siteID string) (*core.Record, error) {
	if e.Auth == nil {
		return nil, e.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	siteRecord, err := e.App.FindRecordById("sites", siteID)
	if err != nil || siteRecord.GetString("user") != e.Auth.Id {
		return nil, e.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden"})
	}
	return siteRecord, nil
}
