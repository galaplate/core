package database

import (
	"strings"
	"testing"
)

func TestPrimaryKeySingleColumn(t *testing.T) {
	blueprint := NewBlueprint("test_table", "mysql")
	blueprint.Integer("id").NotNullable()
	blueprint.String("name", 100).NotNullable()
	blueprint.Primary([]string{"id"})

	sql := blueprint.ToSQL()

	// Check that PRIMARY KEY is in SQL
	if !strings.Contains(sql, "PRIMARY KEY") {
		t.Errorf("Expected SQL to contain PRIMARY KEY, got: %s", sql)
	}

	// Check that it references the id column
	if !strings.Contains(sql, "PRIMARY KEY (id)") {
		t.Errorf("Expected SQL to contain 'PRIMARY KEY (id)', got: %s", sql)
	}
}

func TestPrimaryKeyComposite(t *testing.T) {
	blueprint := NewBlueprint("role_user", "mysql")
	blueprint.Integer("user_id").NotNullable()
	blueprint.Integer("role_id").NotNullable()
	blueprint.Char("kode_satker", 6).NotNullable()
	blueprint.Primary([]string{"user_id", "role_id", "kode_satker"})

	sql := blueprint.ToSQL()

	// Check that PRIMARY KEY is in SQL
	if !strings.Contains(sql, "PRIMARY KEY") {
		t.Errorf("Expected SQL to contain PRIMARY KEY, got: %s", sql)
	}

	// Check that it references all three columns
	if !strings.Contains(sql, "PRIMARY KEY (user_id, role_id, kode_satker)") {
		t.Errorf("Expected SQL to contain composite primary key, got: %s", sql)
	}
}

func TestPrimaryKeyWithCustomName(t *testing.T) {
	blueprint := NewBlueprint("test_table", "postgres")
	blueprint.Integer("id").NotNullable()
	blueprint.Primary([]string{"id"}, "custom_pk_name")

	sql := blueprint.ToSQL()

	// Check that PRIMARY KEY is in SQL
	if !strings.Contains(sql, "PRIMARY KEY (id)") {
		t.Errorf("Expected SQL to contain 'PRIMARY KEY (id)', got: %s", sql)
	}
}

func TestPrimaryKeyDifferentDatabases(t *testing.T) {
	databases := []string{"mysql", "postgres", "sqlite"}

	for _, dbType := range databases {
		t.Run(dbType, func(t *testing.T) {
			blueprint := NewBlueprint("test_table", dbType)
			blueprint.Integer("id").NotNullable()
			blueprint.String("code", 10).NotNullable()
			blueprint.Primary([]string{"id"})

			sql := blueprint.ToSQL()

			if !strings.Contains(sql, "PRIMARY KEY") {
				t.Errorf("[%s] Expected SQL to contain PRIMARY KEY, got: %s", dbType, sql)
			}

			if !strings.Contains(sql, "PRIMARY KEY (id)") {
				t.Errorf("[%s] Expected SQL to contain 'PRIMARY KEY (id)', got: %s", dbType, sql)
			}
		})
	}
}
