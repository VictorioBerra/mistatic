package helpers

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestHasAnyRole(t *testing.T) {
	users := core.NewAuthCollection("users")
	record := core.NewRecord(users)

	for _, role := range []string{"Admin", "Reader"} {
		record.Set("role", role)
		if !HasAnyRole(record, "Admin", "Reader") {
			t.Errorf("expected %s role to be accepted", role)
		}
	}

	for _, role := range []string{"", "admin", "reader", "Owner"} {
		record.Set("role", role)
		if HasAnyRole(record, "Admin", "Reader") {
			t.Errorf("expected %q role to be denied", role)
		}
	}

	superuser := core.NewRecord(core.NewAuthCollection("_superusers"))
	superuser.Set("role", "Admin")
	if HasAnyRole(superuser, "Admin") {
		t.Fatal("expected non-users auth collection to be denied")
	}
}
