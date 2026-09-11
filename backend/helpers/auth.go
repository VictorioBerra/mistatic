package helpers

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
)

func RequireRole(role string) *hook.Handler[*core.RequestEvent] {
	return &hook.Handler[*core.RequestEvent]{
		Func: func(e *core.RequestEvent) error {
			if e.Auth == nil {
				return e.UnauthorizedError("", nil)
			}
			if e.Auth.Collection().Name != "users" || e.Auth.GetString("role") != role {
				return e.ForbiddenError("", nil)
			}
			return e.Next()
		},
	}
}

func HasAnyRole(record *core.Record, roles ...string) bool {
	if record == nil || record.Collection().Name != "users" {
		return false
	}
	for _, role := range roles {
		if record.GetString("role") == role {
			return true
		}
	}
	return false
}
