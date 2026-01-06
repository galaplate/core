package database

import (
	"strings"
	"testing"
)

func TestModifyColumnMySQL(t *testing.T) {
	blueprint := NewBlueprint("users", "mysql")
	blueprint.SetMode("alter")

	// Modify a string column
	blueprint.Modify("email").String("email", 500).NotNullable().Unique()

	sql := blueprint.ToSQL()

	if !strings.Contains(sql, "ALTER TABLE users MODIFY COLUMN") {
		t.Errorf("Expected MySQL MODIFY COLUMN syntax, got: %s", sql)
	}

	if !strings.Contains(sql, "VARCHAR(500)") {
		t.Errorf("Expected VARCHAR(500) type, got: %s", sql)
	}

	if !strings.Contains(sql, "NOT NULL") {
		t.Errorf("Expected NOT NULL constraint, got: %s", sql)
	}

	t.Log("✅ MySQL modify column test passed")
}

func TestModifyColumnPostgres(t *testing.T) {
	blueprint := NewBlueprint("users", "postgres")
	blueprint.SetMode("alter")

	// Modify a column
	blueprint.Modify("status").String("status", 100).NotNullable().Default("active")

	sql := blueprint.ToSQL()

	// PostgreSQL should generate multiple ALTER statements
	if !strings.Contains(sql, "ALTER TABLE users ALTER COLUMN") {
		t.Errorf("Expected PostgreSQL ALTER COLUMN syntax, got: %s", sql)
	}

	if !strings.Contains(sql, "TYPE VARCHAR(100)") {
		t.Errorf("Expected TYPE VARCHAR(100), got: %s", sql)
	}

	if !strings.Contains(sql, "SET NOT NULL") {
		t.Errorf("Expected SET NOT NULL, got: %s", sql)
	}

	if !strings.Contains(sql, "SET DEFAULT") {
		t.Errorf("Expected SET DEFAULT, got: %s", sql)
	}

	t.Log("✅ PostgreSQL modify column test passed")
}

func TestModifyColumnSQLite(t *testing.T) {
	blueprint := NewBlueprint("users", "sqlite")
	blueprint.SetMode("alter")

	blueprint.Modify("age").Integer("age").Nullable()

	sql := blueprint.ToSQL()

	// SQLite should return a comment about limitations
	if !strings.Contains(sql, "SQLite MODIFY not natively supported") {
		t.Errorf("Expected SQLite limitation comment, got: %s", sql)
	}

	t.Log("✅ SQLite modify column test passed")
}

func TestModifyMultipleColumns(t *testing.T) {
	blueprint := NewBlueprint("products", "mysql")
	blueprint.SetMode("alter")

	// Modify multiple columns
	blueprint.Modify("name").String("name", 500).NotNullable()
	blueprint.Modify("price").Decimal("price", 12, 2).Default(0)
	blueprint.Modify("stock").Integer("stock").Default(0)

	sql := blueprint.ToSQL()

	// Should have three MODIFY statements
	modifyCount := strings.Count(sql, "MODIFY COLUMN")
	if modifyCount != 3 {
		t.Errorf("Expected 3 MODIFY COLUMN statements, got %d", modifyCount)
	}

	t.Log("✅ Multiple modify columns test passed")
}

func TestModifyWithDefaultValue(t *testing.T) {
	blueprint := NewBlueprint("users", "mysql")
	blueprint.SetMode("alter")

	blueprint.Modify("is_active").Boolean("is_active").Default(true).NotNullable()

	sql := blueprint.ToSQL()

	if !strings.Contains(sql, "DEFAULT 1") || !strings.Contains(sql, "DEFAULT true") {
		// Either true or 1 is acceptable for boolean default
		if !strings.Contains(sql, "DEFAULT") {
			t.Errorf("Expected DEFAULT value, got: %s", sql)
		}
	}

	t.Log("✅ Modify with default value test passed")
}

func TestModifyColumnTypeChange(t *testing.T) {
	blueprint := NewBlueprint("items", "mysql")
	blueprint.SetMode("alter")

	// Change from INT to BIGINT
	blueprint.Modify("quantity").BigInteger("quantity").NotNullable()

	sql := blueprint.ToSQL()

	if !strings.Contains(sql, "BIGINT") {
		t.Errorf("Expected BIGINT type, got: %s", sql)
	}

	t.Log("✅ Column type change test passed")
}

func TestModifyNullableConstraint(t *testing.T) {
	tests := []struct {
		name        string
		dbType      string
		nullable    bool
		expectedSQL []string
		notExpected string
	}{
		{
			name:        "MySQL make nullable",
			dbType:      "mysql",
			nullable:    true,
			expectedSQL: []string{"ALTER TABLE", "MODIFY COLUMN"},
		},
		{
			name:        "MySQL make not null",
			dbType:      "mysql",
			nullable:    false,
			expectedSQL: []string{"NOT NULL"},
		},
		{
			name:        "PostgreSQL make not null",
			dbType:      "postgres",
			nullable:    false,
			expectedSQL: []string{"SET NOT NULL"},
		},
		{
			name:        "PostgreSQL make nullable",
			dbType:      "postgres",
			nullable:    true,
			expectedSQL: []string{"DROP NOT NULL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blueprint := NewBlueprint("users", tt.dbType)
			blueprint.SetMode("alter")

			if tt.nullable {
				blueprint.Modify("bio").Text("bio").Nullable()
			} else {
				blueprint.Modify("bio").Text("bio").NotNullable()
			}

			sql := blueprint.ToSQL()

			for _, expected := range tt.expectedSQL {
				if !strings.Contains(sql, expected) {
					t.Errorf("Expected '%s' in SQL, got: %s", expected, sql)
				}
			}

			if tt.notExpected != "" && strings.Contains(sql, tt.notExpected) {
				t.Errorf("Did not expect '%s' in SQL, got: %s", tt.notExpected, sql)
			}
		})
	}
}

func TestModifyWithIndex(t *testing.T) {
	blueprint := NewBlueprint("users", "mysql")
	blueprint.SetMode("alter")

	blueprint.Modify("email").String("email", 500).NotNullable().Unique()
	blueprint.Index([]string{"email"})

	sql := blueprint.ToSQL()

	if !strings.Contains(sql, "MODIFY COLUMN") {
		t.Errorf("Expected MODIFY COLUMN, got: %s", sql)
	}

	if !strings.Contains(sql, "CREATE INDEX") {
		t.Errorf("Expected CREATE INDEX, got: %s", sql)
	}

	t.Log("✅ Modify with index test passed")
}

func TestModifyMixedAddAndModify(t *testing.T) {
	blueprint := NewBlueprint("users", "mysql")
	blueprint.SetMode("alter")

	// Add a new column
	blueprint.String("phone", 20).Nullable()

	// Modify existing column
	blueprint.Modify("email").String("email", 500).NotNullable()

	sql := blueprint.ToSQL()

	// Should have both ADD and MODIFY
	if !strings.Contains(sql, "ADD COLUMN") {
		t.Errorf("Expected ADD COLUMN, got: %s", sql)
	}

	if !strings.Contains(sql, "MODIFY COLUMN") {
		t.Errorf("Expected MODIFY COLUMN, got: %s", sql)
	}

	t.Log("✅ Mixed add and modify test passed")
}

func TestModifyChaining(t *testing.T) {
	blueprint := NewBlueprint("users", "mysql")
	blueprint.SetMode("alter")

	// Test method chaining works correctly for modify
	result := blueprint.
		Modify("name").String("name", 300).NotNullable().
		Modify("bio").Text("bio").Nullable().
		Modify("age").Integer("age")

	sql := result.ToSQL()

	modifyCount := strings.Count(sql, "MODIFY COLUMN")
	if modifyCount != 3 {
		t.Errorf("Expected 3 MODIFY COLUMN statements, got %d in: %s", modifyCount, sql)
	}

	t.Log("✅ Modify chaining test passed")
}

func TestModifyStringToLongText(t *testing.T) {
	blueprint := NewBlueprint("articles", "mysql")
	blueprint.SetMode("alter")

	// Change from String to LongText
	blueprint.Modify("content").LongText("content").Nullable()

	sql := blueprint.ToSQL()

	if !strings.Contains(sql, "LONGTEXT") {
		t.Errorf("Expected LONGTEXT type, got: %s", sql)
	}

	t.Log("✅ String to LongText conversion test passed")
}

func TestModifyDecimalPrecision(t *testing.T) {
	blueprint := NewBlueprint("products", "mysql")
	blueprint.SetMode("alter")

	// Increase decimal precision
	blueprint.Modify("price").Decimal("price", 15, 4).NotNullable()

	sql := blueprint.ToSQL()

	if !strings.Contains(sql, "DECIMAL(15,4)") {
		t.Errorf("Expected DECIMAL(15,4), got: %s", sql)
	}

	t.Log("✅ Decimal precision modification test passed")
}
