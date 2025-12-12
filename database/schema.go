package database

import (
	"fmt"
	"strings"

	config "github.com/galaplate/core/env"
	"gorm.io/gorm"
)

// Schema provides database schema operations
type Schema struct {
	db     *gorm.DB
	dbType string
}

// NewSchema creates a new Schema instance
func NewSchema() *Schema {
	return &Schema{
		db:     Connect,
		dbType: config.Get("DB_CONNECTION"),
	}
}

// Create creates a new table
func (s *Schema) Create(tableName string, callback func(table *Blueprint)) error {
	blueprint := NewBlueprint(tableName, s.dbType)
	callback(blueprint)

	sql := blueprint.ToSQL()
	return s.db.Exec(sql).Error
}

// Table modifies an existing table
func (s *Schema) Table(tableName string, callback func(table *Blueprint)) error {
	blueprint := NewBlueprint(tableName, s.dbType)
	blueprint.SetMode("alter")
	callback(blueprint)

	sql := blueprint.ToSQL()
	return s.db.Exec(sql).Error
}

// Drop drops a table
func (s *Schema) Drop(tableName string) error {
	sql := fmt.Sprintf("DROP TABLE %s;", tableName)
	return s.db.Exec(sql).Error
}

// DropIfExists drops a table if it exists
func (s *Schema) DropIfExists(tableName string) error {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)
	return s.db.Exec(sql).Error
}

// HasTable checks if table exists
func (s *Schema) HasTable(tableName string) bool {
	var count int64

	switch s.dbType {
	case "mysql":
		s.db.Raw("SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?", tableName).Scan(&count)
	case "postgres":
		s.db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?", tableName).Scan(&count)
	case "sqlite":
		s.db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?", tableName).Scan(&count)
	}

	return count > 0
}

// HasColumn checks if column exists
func (s *Schema) HasColumn(tableName, columnName string) bool {
	var count int64

	switch s.dbType {
	case "mysql":
		s.db.Raw("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?", tableName, columnName).Scan(&count)
	case "postgres":
		s.db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = ? AND column_name = ?", tableName, columnName).Scan(&count)
	case "sqlite":
		s.db.Raw("SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?", tableName, columnName).Scan(&count)
	}

	return count > 0
}

// Blueprint represents a table blueprint for building schema
type Blueprint struct {
	tableName string
	dbType    string
	mode      string // "create" or "alter"
	columns   []Column
	indexes   []Index
	foreign   []ForeignKey
}

// NewBlueprint creates a new Blueprint
func NewBlueprint(tableName, dbType string) *Blueprint {
	return &Blueprint{
		tableName: tableName,
		dbType:    dbType,
		mode:      "create",
		columns:   []Column{},
		indexes:   []Index{},
		foreign:   []ForeignKey{},
	}
}

// SetMode sets the blueprint mode (create/alter)
func (b *Blueprint) SetMode(mode string) {
	b.mode = mode
}

// Column represents a database column
type Column struct {
	Name       string
	Type       string
	Length     int
	Precision  int
	Scale      int
	Nullable   bool
	Default    interface{}
	Unique     bool
	Primary    bool
	Auto       bool
	Comment    string
	After      string // For MySQL ALTER TABLE
	EnumValues []string
}

// Index represents a database index
type Index struct {
	Name    string
	Columns []string
	Type    string // "index", "unique", "primary"
}

// ForeignKey represents a foreign key constraint
type ForeignKey struct {
	Column     string
	References string
	On         string
	OnDelete   string
	OnUpdate   string
}

// Column type methods

// ID creates an auto-incrementing primary key
func (b *Blueprint) ID() *Blueprint {
	col := Column{
		Name:    "id",
		Type:    "id",
		Primary: true,
		Auto:    true,
	}
	b.columns = append(b.columns, col)
	return b
}

// String creates a VARCHAR column
func (b *Blueprint) String(name string, length ...int) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "string",
		Length:   255,
		Nullable: true,
	}
	if len(length) > 0 {
		col.Length = length[0]
	}
	b.columns = append(b.columns, col)
	return b
}

// Text creates a TEXT column
func (b *Blueprint) Text(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "text",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// Integer creates an INT column
func (b *Blueprint) Integer(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "integer",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// BigInteger creates a BIGINT column
func (b *Blueprint) BigInteger(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "bigint",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// Boolean creates a BOOLEAN column
func (b *Blueprint) Boolean(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "boolean",
		Nullable: true,
		Default:  false,
	}
	b.columns = append(b.columns, col)
	return b
}

// JSON creates a JSON column
func (b *Blueprint) JSON(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "json",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// Timestamp creates a TIMESTAMP column
func (b *Blueprint) Timestamp(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "timestamp",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// Timestamps creates created_at and updated_at columns
func (b *Blueprint) Timestamps() *Blueprint {
	b.Timestamp("created_at").NotNullable().Default("CURRENT_TIMESTAMP")
	b.Timestamp("updated_at").Nullable()
	return b
}

// Decimal creates a DECIMAL column
func (b *Blueprint) Decimal(name string, precision, scale int) *Blueprint {
	col := Column{
		Name:      name,
		Type:      "decimal",
		Precision: precision,
		Scale:     scale,
		Nullable:  true,
	}
	b.columns = append(b.columns, col)
	return b
}

// UUID creates a UUID column
func (b *Blueprint) UUID(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "uuid",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// Enum creates an ENUM column
func (b *Blueprint) Enum(name string, values []string) *Blueprint {
	col := Column{
		Name:       name,
		Type:       "enum",
		EnumValues: values,
		Nullable:   true,
	}
	b.columns = append(b.columns, col)
	return b
}

// Date creates a DATE column
func (b *Blueprint) Date(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "date",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// DateTime creates a DATETIME column
func (b *Blueprint) DateTime(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "datetime",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// Time creates a TIME column
func (b *Blueprint) Time(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "time",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// Year creates a YEAR column
func (b *Blueprint) Year(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "year",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// Float creates a FLOAT column
func (b *Blueprint) Float(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "float",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// Double creates a DOUBLE column
func (b *Blueprint) Double(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "double",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// Char creates a CHAR column
func (b *Blueprint) Char(name string, length ...int) *Blueprint {
	columnLength := 255
	if len(length) > 0 {
		columnLength = length[0]
	}
	col := Column{
		Name:     name,
		Type:     "char",
		Length:   columnLength,
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// Binary creates a BINARY column
func (b *Blueprint) Binary(name string, length ...int) *Blueprint {
	columnLength := 255
	if len(length) > 0 {
		columnLength = length[0]
	}
	col := Column{
		Name:     name,
		Type:     "binary",
		Length:   columnLength,
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// VarBinary creates a VARBINARY column
func (b *Blueprint) VarBinary(name string, length ...int) *Blueprint {
	columnLength := 255
	if len(length) > 0 {
		columnLength = length[0]
	}
	col := Column{
		Name:     name,
		Type:     "varbinary",
		Length:   columnLength,
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// Blob creates a BLOB column
func (b *Blueprint) Blob(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "blob",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// MediumBlob creates a MEDIUMBLOB column
func (b *Blueprint) MediumBlob(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "mediumblob",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// LongBlob creates a LONGBLOB column
func (b *Blueprint) LongBlob(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "longblob",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// TinyInt creates a TINYINT column
func (b *Blueprint) TinyInt(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "tinyint",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// SmallInt creates a SMALLINT column
func (b *Blueprint) SmallInt(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "smallint",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// MediumInt creates a MEDIUMINT column
func (b *Blueprint) MediumInt(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "mediumint",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// TinyText creates a TINYTEXT column
func (b *Blueprint) TinyText(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "tinytext",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// MediumText creates a MEDIUMTEXT column
func (b *Blueprint) MediumText(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "mediumtext",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// LongText creates a LONGTEXT column
func (b *Blueprint) LongText(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "longtext",
		Nullable: true,
	}
	b.columns = append(b.columns, col)
	return b
}

// Column modifier methods - chainable

// NotNullable makes the column NOT NULL
func (b *Blueprint) NotNullable() *Blueprint {
	if len(b.columns) > 0 {
		b.columns[len(b.columns)-1].Nullable = false
	}
	return b
}

// Nullable makes the column nullable
func (b *Blueprint) Nullable() *Blueprint {
	if len(b.columns) > 0 {
		b.columns[len(b.columns)-1].Nullable = true
	}
	return b
}

// Default sets a default value
func (b *Blueprint) Default(value interface{}) *Blueprint {
	if len(b.columns) > 0 {
		b.columns[len(b.columns)-1].Default = value
	}
	return b
}

// Unique makes the column unique
func (b *Blueprint) Unique() *Blueprint {
	if len(b.columns) > 0 {
		b.columns[len(b.columns)-1].Unique = true
	}
	return b
}

// Comment adds a comment
func (b *Blueprint) Comment(comment string) *Blueprint {
	if len(b.columns) > 0 {
		b.columns[len(b.columns)-1].Comment = comment
	}
	return b
}

// Index methods

// Index creates an index
func (b *Blueprint) Index(columns []string, name ...string) *Blueprint {
	indexName := fmt.Sprintf("idx_%s_%s", b.tableName, strings.Join(columns, "_"))
	if len(name) > 0 {
		indexName = name[0]
	}

	idx := Index{
		Name:    indexName,
		Columns: columns,
		Type:    "index",
	}
	b.indexes = append(b.indexes, idx)
	return b
}

// UniqueIndex creates a unique index
func (b *Blueprint) UniqueIndex(columns []string, name ...string) *Blueprint {
	indexName := fmt.Sprintf("unique_%s_%s", b.tableName, strings.Join(columns, "_"))
	if len(name) > 0 {
		indexName = name[0]
	}

	idx := Index{
		Name:    indexName,
		Columns: columns,
		Type:    "unique",
	}
	b.indexes = append(b.indexes, idx)
	return b
}

// Foreign creates a foreign key
func (b *Blueprint) Foreign(column string) *ForeignKeyBuilder {
	return &ForeignKeyBuilder{
		blueprint: b,
		column:    column,
	}
}

// ForeignKeyBuilder helps build foreign key constraints
type ForeignKeyBuilder struct {
	blueprint *Blueprint
	column    string
	fk        ForeignKey
}

// References sets the referenced column
func (fkb *ForeignKeyBuilder) References(column string) *ForeignKeyBuilder {
	fkb.fk.References = column
	return fkb
}

// On sets the referenced table
func (fkb *ForeignKeyBuilder) On(table string) *ForeignKeyBuilder {
	fkb.fk.On = table
	return fkb
}

// OnDelete sets the ON DELETE action
func (fkb *ForeignKeyBuilder) OnDelete(action string) *ForeignKeyBuilder {
	fkb.fk.OnDelete = action
	return fkb
}

// OnUpdate sets the ON UPDATE action
func (fkb *ForeignKeyBuilder) OnUpdate(action string) *ForeignKeyBuilder {
	fkb.fk.OnUpdate = action
	return fkb
}

// Finish completes the foreign key definition
func (fkb *ForeignKeyBuilder) Finish() *Blueprint {
	fkb.fk.Column = fkb.column
	fkb.blueprint.foreign = append(fkb.blueprint.foreign, fkb.fk)
	return fkb.blueprint
}

// ToSQL converts the blueprint to SQL - handles database differences
func (b *Blueprint) ToSQL() string {
	switch b.mode {
	case "alter":
		return b.toAlterSQL()
	default:
		return b.toCreateSQL()
	}
}

// toCreateSQL generates CREATE TABLE SQL
func (b *Blueprint) toCreateSQL() string {
	var sql strings.Builder

	sql.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", b.tableName))

	// Add columns
	columnSQLs := make([]string, 0, len(b.columns))
	for _, col := range b.columns {
		columnSQLs = append(columnSQLs, b.columnToSQL(col))
	}

	sql.WriteString("  " + strings.Join(columnSQLs, ",\n  "))

	// Add indexes
	for _, idx := range b.indexes {
		sql.WriteString(",\n  " + b.indexToSQL(idx))
	}

	// Add foreign keys
	for _, fk := range b.foreign {
		sql.WriteString(",\n  " + b.foreignKeyToSQL(fk))
	}

	sql.WriteString("\n);")

	return sql.String()
}

// toAlterSQL generates ALTER TABLE SQL
func (b *Blueprint) toAlterSQL() string {
	var sqls []string

	// Add columns
	for _, col := range b.columns {
		sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;", b.tableName, b.columnToSQL(col))
		sqls = append(sqls, sql)
	}

	// Add indexes
	for _, idx := range b.indexes {
		var sql string
		switch idx.Type {
		case "unique":
			sql = fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s);",
				idx.Name, b.tableName, strings.Join(idx.Columns, ", "))
		default:
			sql = fmt.Sprintf("CREATE INDEX %s ON %s (%s);",
				idx.Name, b.tableName, strings.Join(idx.Columns, ", "))
		}
		sqls = append(sqls, sql)
	}

	return strings.Join(sqls, "\n")
}

// columnToSQL converts a column to SQL
func (b *Blueprint) columnToSQL(col Column) string {
	parts := []string{col.Name}

	// Get database-specific type
	dbType := b.getColumnType(col)
	parts = append(parts, dbType)

	// Nullable
	if !col.Nullable {
		parts = append(parts, "NOT NULL")
	}

	// Default value
	if col.Default != nil {
		if str, ok := col.Default.(string); ok {
			// Special case for CURRENT_TIMESTAMP
			if strings.ToUpper(str) == "CURRENT_TIMESTAMP" {
				parts = append(parts, "DEFAULT CURRENT_TIMESTAMP")
			} else {
				// Quote string values (for ENUM, VARCHAR, etc.)
				parts = append(parts, fmt.Sprintf("DEFAULT '%s'", strings.ReplaceAll(str, "'", "''")))
			}
		} else {
			// Numeric or boolean values - no quotes needed
			parts = append(parts, fmt.Sprintf("DEFAULT %v", col.Default))
		}
	}

	// Unique
	if col.Unique {
		parts = append(parts, "UNIQUE")
	}

	// Auto increment (MySQL only)
	if col.Auto && b.dbType == "mysql" {
		parts = append(parts, "AUTO_INCREMENT")
	}

	// Comment (MySQL only)
	if col.Comment != "" && b.dbType == "mysql" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", col.Comment))
	}

	// Add CHECK constraint for enum in PostgreSQL and SQLite
	if col.Type == "enum" && len(col.EnumValues) > 0 && (b.dbType == "postgres" || b.dbType == "sqlite") {
		quotedValues := make([]string, len(col.EnumValues))
		for i, v := range col.EnumValues {
			quotedValues[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
		}
		checkConstraint := fmt.Sprintf("CHECK (%s IN (%s))", col.Name, strings.Join(quotedValues, ", "))
		parts = append(parts, checkConstraint)
	}

	return strings.Join(parts, " ")
}

// getColumnType returns database-specific column type
func (b *Blueprint) getColumnType(col Column) string {
	typeMap := map[string]map[string]string{
		"id": {
			"mysql":    "BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY",
			"postgres": "BIGSERIAL PRIMARY KEY",
			"sqlite":   "INTEGER PRIMARY KEY AUTOINCREMENT",
		},
		"string": {
			"mysql":    fmt.Sprintf("VARCHAR(%d)", col.Length),
			"postgres": fmt.Sprintf("VARCHAR(%d)", col.Length),
			"sqlite":   fmt.Sprintf("VARCHAR(%d)", col.Length),
		},
		"text": {
			"mysql":    "TEXT",
			"postgres": "TEXT",
			"sqlite":   "TEXT",
		},
		"integer": {
			"mysql":    "INT",
			"postgres": "INTEGER",
			"sqlite":   "INTEGER",
		},
		"bigint": {
			"mysql":    "BIGINT",
			"postgres": "BIGINT",
			"sqlite":   "INTEGER",
		},
		"boolean": {
			"mysql":    "TINYINT(1)",
			"postgres": "BOOLEAN",
			"sqlite":   "INTEGER",
		},
		"json": {
			"mysql":    "JSON",
			"postgres": "JSONB",
			"sqlite":   "TEXT",
		},
		"timestamp": {
			"mysql":    "TIMESTAMP",
			"postgres": "TIMESTAMP",
			"sqlite":   "DATETIME",
		},
		"date": {
			"mysql":    "DATE",
			"postgres": "DATE",
			"sqlite":   "DATE",
		},
		"datetime": {
			"mysql":    "DATETIME",
			"postgres": "TIMESTAMP",
			"sqlite":   "DATETIME",
		},
		"decimal": {
			"mysql":    fmt.Sprintf("DECIMAL(%d,%d)", col.Precision, col.Scale),
			"postgres": fmt.Sprintf("DECIMAL(%d,%d)", col.Precision, col.Scale),
			"sqlite":   "REAL",
		},
		"uuid": {
			"mysql":    "CHAR(36)",
			"postgres": "UUID",
			"sqlite":   "CHAR(36)",
		},
		"float": {
			"mysql":    "FLOAT",
			"postgres": "REAL",
			"sqlite":   "REAL",
		},
		"double": {
			"mysql":    "DOUBLE",
			"postgres": "DOUBLE PRECISION",
			"sqlite":   "REAL",
		},
		"char": {
			"mysql":    fmt.Sprintf("CHAR(%d)", col.Length),
			"postgres": fmt.Sprintf("CHAR(%d)", col.Length),
			"sqlite":   fmt.Sprintf("CHAR(%d)", col.Length),
		},
		"binary": {
			"mysql":    fmt.Sprintf("BINARY(%d)", col.Length),
			"postgres": "BYTEA",
			"sqlite":   "BLOB",
		},
		"varbinary": {
			"mysql":    fmt.Sprintf("VARBINARY(%d)", col.Length),
			"postgres": "BYTEA",
			"sqlite":   "BLOB",
		},
		"blob": {
			"mysql":    "BLOB",
			"postgres": "BYTEA",
			"sqlite":   "BLOB",
		},
		"mediumblob": {
			"mysql":    "MEDIUMBLOB",
			"postgres": "BYTEA",
			"sqlite":   "BLOB",
		},
		"longblob": {
			"mysql":    "LONGBLOB",
			"postgres": "BYTEA",
			"sqlite":   "BLOB",
		},
		"tinyint": {
			"mysql":    "TINYINT",
			"postgres": "SMALLINT",
			"sqlite":   "INTEGER",
		},
		"smallint": {
			"mysql":    "SMALLINT",
			"postgres": "SMALLINT",
			"sqlite":   "INTEGER",
		},
		"mediumint": {
			"mysql":    "MEDIUMINT",
			"postgres": "INTEGER",
			"sqlite":   "INTEGER",
		},
		"tinytext": {
			"mysql":    "TINYTEXT",
			"postgres": "TEXT",
			"sqlite":   "TEXT",
		},
		"mediumtext": {
			"mysql":    "MEDIUMTEXT",
			"postgres": "TEXT",
			"sqlite":   "TEXT",
		},
		"longtext": {
			"mysql":    "LONGTEXT",
			"postgres": "TEXT",
			"sqlite":   "TEXT",
		},
		"time": {
			"mysql":    "TIME",
			"postgres": "TIME",
			"sqlite":   "TEXT",
		},
		"year": {
			"mysql":    "YEAR",
			"postgres": "SMALLINT",
			"sqlite":   "INTEGER",
		},
	}

	// Handle enum type specially
	if col.Type == "enum" {
		return b.getEnumType(col)
	}

	if dbTypes, exists := typeMap[col.Type]; exists {
		if dbType, exists := dbTypes[b.dbType]; exists {
			return dbType
		}
	}

	return col.Type // fallback to original type
}

// getEnumType returns database-specific ENUM type
func (b *Blueprint) getEnumType(col Column) string {
	if len(col.EnumValues) == 0 {
		return "VARCHAR(255)" // fallback if no values provided
	}

	switch b.dbType {
	case "mysql":
		// MySQL native ENUM: ENUM('value1', 'value2', 'value3')
		quotedValues := make([]string, len(col.EnumValues))
		for i, v := range col.EnumValues {
			quotedValues[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
		}
		return fmt.Sprintf("ENUM(%s)", strings.Join(quotedValues, ", "))
	case "postgres":
		// PostgreSQL uses CHECK constraint for enum-like behavior
		// Return VARCHAR and we'll add CHECK constraint separately
		return "VARCHAR(255)"
	case "sqlite":
		// SQLite uses CHECK constraint for enum-like behavior
		return "TEXT"
	default:
		return "VARCHAR(255)"
	}
}

// indexToSQL converts an index to SQL
func (b *Blueprint) indexToSQL(idx Index) string {
	switch idx.Type {
	case "unique":
		return fmt.Sprintf("UNIQUE INDEX %s (%s)", idx.Name, strings.Join(idx.Columns, ", "))
	default:
		return fmt.Sprintf("INDEX %s (%s)", idx.Name, strings.Join(idx.Columns, ", "))
	}
}

// foreignKeyToSQL converts a foreign key to SQL
func (b *Blueprint) foreignKeyToSQL(fk ForeignKey) string {
	sql := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)", fk.Column, fk.On, fk.References)

	if fk.OnDelete != "" {
		sql += " ON DELETE " + fk.OnDelete
	}
	if fk.OnUpdate != "" {
		sql += " ON UPDATE " + fk.OnUpdate
	}

	return sql
}
