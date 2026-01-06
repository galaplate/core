package database

import (
	"testing"
)

func TestDeferrableForeignKey(t *testing.T) {
	blueprint := NewBlueprint("orders", "postgres")

	// Test basic deferrable FK
	blueprint.Foreign("user_id").
		References("id").
		On("users").
		OnDelete("CASCADE").
		Deferrable().
		Finish()

	sql := blueprint.foreignKeyToSQL(blueprint.foreign[0])

	if !contains(sql, "DEFERRABLE") {
		t.Errorf("Expected DEFERRABLE in SQL, got: %s", sql)
	}

	if contains(sql, "INITIALLY DEFERRED") {
		t.Errorf("Should not have INITIALLY DEFERRED, got: %s", sql)
	}
}

func TestInitiallyDeferredForeignKey(t *testing.T) {
	blueprint := NewBlueprint("orders", "postgres")

	// Test initially deferred FK
	blueprint.Foreign("user_id").
		References("id").
		On("users").
		OnDelete("CASCADE").
		InitiallyDeferred().
		Finish()

	sql := blueprint.foreignKeyToSQL(blueprint.foreign[0])

	if !contains(sql, "DEFERRABLE INITIALLY DEFERRED") {
		t.Errorf("Expected 'DEFERRABLE INITIALLY DEFERRED' in SQL, got: %s", sql)
	}
}

func TestRegularForeignKeyUnchanged(t *testing.T) {
	blueprint := NewBlueprint("orders", "mysql")

	// Test regular FK (no deferrable)
	blueprint.Foreign("user_id").
		References("id").
		On("users").
		OnDelete("CASCADE").
		Finish()

	sql := blueprint.foreignKeyToSQL(blueprint.foreign[0])

	if contains(sql, "DEFERRABLE") {
		t.Errorf("Regular FK should not have DEFERRABLE, got: %s", sql)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
