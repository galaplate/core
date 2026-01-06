package database

import (
	"strings"
	"testing"
)

func TestUnsignedInteger(t *testing.T) {
	blueprint := NewBlueprint("test_table", "mysql")
	blueprint.Integer("user_id").NotNullable().Unsigned()
	blueprint.BigInteger("count").Unsigned()

	sql := blueprint.ToSQL()

	// Check that UNSIGNED is in SQL for integer
	if !strings.Contains(sql, "INT UNSIGNED") {
		t.Errorf("Expected SQL to contain 'INT UNSIGNED', got: %s", sql)
	}

	// Check that UNSIGNED is in SQL for bigint
	if !strings.Contains(sql, "BIGINT UNSIGNED") {
		t.Errorf("Expected SQL to contain 'BIGINT UNSIGNED', got: %s", sql)
	}
}

func TestUnsignedDecimal(t *testing.T) {
	blueprint := NewBlueprint("test_table", "mysql")
	blueprint.Decimal("amount", 10, 2).Unsigned()

	sql := blueprint.ToSQL()

	// Check that UNSIGNED is in SQL for decimal
	if !strings.Contains(sql, "DECIMAL(10,2) UNSIGNED") {
		t.Errorf("Expected SQL to contain 'DECIMAL(10,2) UNSIGNED', got: %s", sql)
	}
}

func TestUnsignedWithOtherModifiers(t *testing.T) {
	blueprint := NewBlueprint("test_table", "mysql")
	blueprint.Integer("id").NotNullable().Unsigned().Default(0)

	sql := blueprint.ToSQL()

	// Check that all modifiers are present
	if !strings.Contains(sql, "INT UNSIGNED") {
		t.Errorf("Expected SQL to contain 'INT UNSIGNED', got: %s", sql)
	}
	if !strings.Contains(sql, "NOT NULL") {
		t.Errorf("Expected SQL to contain 'NOT NULL', got: %s", sql)
	}
	if !strings.Contains(sql, "DEFAULT 0") {
		t.Errorf("Expected SQL to contain 'DEFAULT 0', got: %s", sql)
	}
}

func TestUnsignedDifferentDatabases(t *testing.T) {
	tests := []struct {
		dbType   string
		contains string
	}{
		{"mysql", "INT UNSIGNED"},
		{"postgres", "INTEGER"}, // PostgreSQL doesn't have UNSIGNED
		{"sqlite", "INTEGER"},   // SQLite doesn't have UNSIGNED
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			blueprint := NewBlueprint("test_table", tt.dbType)
			blueprint.Integer("id").Unsigned()

			sql := blueprint.ToSQL()

			if !strings.Contains(sql, tt.contains) {
				t.Errorf("[%s] Expected SQL to contain '%s', got: %s", tt.dbType, tt.contains, sql)
			}
		})
	}
}

func TestUnsignedAllNumericTypes(t *testing.T) {
	blueprint := NewBlueprint("test_table", "mysql")
	blueprint.TinyInt("tiny").Unsigned()
	blueprint.SmallInt("small").Unsigned()
	blueprint.MediumInt("medium").Unsigned()
	blueprint.Integer("int").Unsigned()
	blueprint.BigInteger("big").Unsigned()
	blueprint.Float("float_col").Unsigned()
	blueprint.Double("double_col").Unsigned()
	blueprint.Decimal("decimal_col", 8, 2).Unsigned()

	sql := blueprint.ToSQL()

	// Check that UNSIGNED is added to all numeric types
	expectedTypes := []string{
		"TINYINT UNSIGNED",
		"SMALLINT UNSIGNED",
		"MEDIUMINT UNSIGNED",
		"INT UNSIGNED",
		"BIGINT UNSIGNED",
		"FLOAT UNSIGNED",
		"DOUBLE UNSIGNED",
		"DECIMAL(8,2) UNSIGNED",
	}

	for _, expected := range expectedTypes {
		if !strings.Contains(sql, expected) {
			t.Errorf("Expected SQL to contain '%s', got: %s", expected, sql)
		}
	}
}

func TestUnsignedNotAppliedToNonNumeric(t *testing.T) {
	blueprint := NewBlueprint("test_table", "mysql")
	blueprint.String("name", 100).Unsigned() // Should not add UNSIGNED to string

	sql := blueprint.ToSQL()

	// UNSIGNED should NOT be in SQL for VARCHAR
	if strings.Contains(sql, "VARCHAR(100) UNSIGNED") {
		t.Errorf("Expected SQL not to contain 'VARCHAR(100) UNSIGNED', got: %s", sql)
	}
}
