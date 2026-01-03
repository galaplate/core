package database

import (
	"fmt"
	"strings"

	config "github.com/galaplate/core/env"
	"github.com/galaplate/core/supports"
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
		dbType: supports.MapPostgres(config.Get("DB_CONNECTION")),
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
	tableName    string
	dbType       string
	mode         string // "create" or "alter"
	columns      []Column
	indexes      []Index
	foreign      []ForeignKey
	dropColumns  []string
	dropIndexes  []string
	dropForeigns []string
}

// NewBlueprint creates a new Blueprint
func NewBlueprint(tableName, dbType string) *Blueprint {
	return &Blueprint{
		tableName:    tableName,
		dbType:       supports.MapPostgres(dbType),
		mode:         "create",
		columns:      []Column{},
		indexes:      []Index{},
		foreign:      []ForeignKey{},
		dropColumns:  []string{},
		dropIndexes:  []string{},
		dropForeigns: []string{},
	}
}

// SetMode sets the blueprint mode (create/alter)
func (b *Blueprint) SetMode(mode string) {
	b.mode = mode
}

// addOrUpdateColumn adds a column or updates the last one if it's a modify with matching name
func (b *Blueprint) addOrUpdateColumn(col Column) {
	// If last column is a modify with matching name, update it instead of appending
	if len(b.columns) > 0 && b.columns[len(b.columns)-1].Modify && b.columns[len(b.columns)-1].Name == col.Name {
		col.Modify = true
		b.columns[len(b.columns)-1] = col
	} else {
		b.columns = append(b.columns, col)
	}
}

// quoteIdentifier quotes an identifier based on database type
func (b *Blueprint) quoteIdentifier(name string) string {
	// Check if identifier needs quoting (reserved words, special characters, etc.)
	if b.needsQuoting(name) {
		switch b.dbType {
		case "mysql":
			return fmt.Sprintf("`%s`", name)
		case "postgres", "sqlite":
			return fmt.Sprintf("\"%s\"", name)
		default:
			return name
		}
	}
	return name
}

// needsQuoting checks if an identifier needs to be quoted
func (b *Blueprint) needsQuoting(name string) bool {
	// Common reserved words that often cause issues
	reservedWords := map[string]bool{
		"key": true, "user": true, "order": true, "group": true, "table": true,
		"column": true, "index": true, "primary": true, "foreign": true, "unique": true,
		"check": true, "default": true, "null": true, "select": true, "insert": true,
		"update": true, "delete": true, "create": true, "drop": true, "alter": true,
		"where": true, "from": true, "join": true, "limit": true, "offset": true,
		"having": true, "union": true, "begin": true, "commit": true, "rollback": true,
		"set": true, "show": true, "type": true, "time": true, "timestamp": true,
		"date": true, "boolean": true, "text": true, "varchar": true, "char": true,
		"integer": true, "bigint": true, "serial": true, "bigserial": true,
	}

	// Check if it's a reserved word
	if reservedWords[strings.ToLower(name)] {
		return true
	}

	// Check if it contains special characters or starts with a number
	if strings.ContainsAny(name, " -+*/=<>!@#$%^&()[]{}|\\:;\"'.,?~`") || (len(name) > 0 && name[0] >= '0' && name[0] <= '9') {
		return true
	}

	return false
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
	Unsigned   bool
	Comment    string
	After      string // For MySQL ALTER TABLE
	EnumValues []string
	Modify     bool // Whether this column is being modified
}

// Index represents a database index
type Index struct {
	Name    string
	Columns []string
	Type    string // "index", "unique", "primary"
}

// ForeignKey represents a foreign key constraint
type ForeignKey struct {
	Column            string
	References        string
	On                string
	OnDelete          string
	OnUpdate          string
	Deferrable        bool
	InitiallyDeferred bool
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
	b.addOrUpdateColumn(col)
	return b
}

// Text creates a TEXT column
func (b *Blueprint) Text(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "text",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// Integer creates an INT column
func (b *Blueprint) Integer(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "integer",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// BigInteger creates a BIGINT column
func (b *Blueprint) BigInteger(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "bigint",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
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
	b.addOrUpdateColumn(col)
	return b
}

// JSON creates a JSON column
func (b *Blueprint) JSON(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "json",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// Timestamp creates a TIMESTAMP column
func (b *Blueprint) Timestamp(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "timestamp",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
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
	b.addOrUpdateColumn(col)
	return b
}

// UUID creates a UUID column
func (b *Blueprint) UUID(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "uuid",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
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
	b.addOrUpdateColumn(col)
	return b
}

// Date creates a DATE column
func (b *Blueprint) Date(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "date",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// DateTime creates a DATETIME column
func (b *Blueprint) DateTime(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "datetime",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// Time creates a TIME column
func (b *Blueprint) Time(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "time",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// Year creates a YEAR column
func (b *Blueprint) Year(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "year",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// Float creates a FLOAT column
func (b *Blueprint) Float(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "float",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// Double creates a DOUBLE column
func (b *Blueprint) Double(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "double",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
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
	b.addOrUpdateColumn(col)
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
	b.addOrUpdateColumn(col)
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
	b.addOrUpdateColumn(col)
	return b
}

// Blob creates a BLOB column
func (b *Blueprint) Blob(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "blob",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// MediumBlob creates a MEDIUMBLOB column
func (b *Blueprint) MediumBlob(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "mediumblob",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// LongBlob creates a LONGBLOB column
func (b *Blueprint) LongBlob(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "longblob",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// TinyInt creates a TINYINT column
func (b *Blueprint) TinyInt(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "tinyint",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// SmallInt creates a SMALLINT column
func (b *Blueprint) SmallInt(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "smallint",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// MediumInt creates a MEDIUMINT column
func (b *Blueprint) MediumInt(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "mediumint",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// TinyText creates a TINYTEXT column
func (b *Blueprint) TinyText(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "tinytext",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// MediumText creates a MEDIUMTEXT column
func (b *Blueprint) MediumText(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "mediumtext",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
	return b
}

// LongText creates a LONGTEXT column
func (b *Blueprint) LongText(name string) *Blueprint {
	col := Column{
		Name:     name,
		Type:     "longtext",
		Nullable: true,
	}
	b.addOrUpdateColumn(col)
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

// Unsigned makes the column unsigned (for numeric types)
func (b *Blueprint) Unsigned() *Blueprint {
	if len(b.columns) > 0 {
		b.columns[len(b.columns)-1].Unsigned = true
	}
	return b
}

// Signed makes the column signed (for numeric types) - this is the default
func (b *Blueprint) Signed() *Blueprint {
	if len(b.columns) > 0 {
		b.columns[len(b.columns)-1].Unsigned = false
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

// Primary creates a primary key index
func (b *Blueprint) Primary(columns []string, name ...string) *Blueprint {
	indexName := fmt.Sprintf("pk_%s", b.tableName)
	if len(name) > 0 {
		indexName = name[0]
	}

	idx := Index{
		Name:    indexName,
		Columns: columns,
		Type:    "primary",
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

// Deferrable marks the foreign key as deferrable
func (fkb *ForeignKeyBuilder) Deferrable() *ForeignKeyBuilder {
	fkb.fk.Deferrable = true
	return fkb
}

// InitiallyDeferred marks the foreign key as initially deferred
// (only applies if Deferrable() is also called)
func (fkb *ForeignKeyBuilder) InitiallyDeferred() *ForeignKeyBuilder {
	fkb.fk.Deferrable = true
	fkb.fk.InitiallyDeferred = true
	return fkb
}

// Finish completes the foreign key definition
func (fkb *ForeignKeyBuilder) Finish() *Blueprint {
	fkb.fk.Column = fkb.column
	fkb.blueprint.foreign = append(fkb.blueprint.foreign, fkb.fk)
	return fkb.blueprint
}

// Modify methods

// Modify modifies an existing column - returns a special blueprint for chaining
func (b *Blueprint) Modify(name string) *Blueprint {
	col := Column{
		Name:     name,
		Modify:   true,
		Nullable: true, // Default nullable
	}
	b.columns = append(b.columns, col)
	return b
}

// Drop methods

// DropColumn drops a column
func (b *Blueprint) DropColumn(column string) *Blueprint {
	b.dropColumns = append(b.dropColumns, column)
	return b
}

// DropIndex drops an index
func (b *Blueprint) DropIndex(indexName string) *Blueprint {
	b.dropIndexes = append(b.dropIndexes, indexName)
	return b
}

// DropUniqueIndex drops a unique index (alias for DropIndex)
func (b *Blueprint) DropUniqueIndex(indexName string) *Blueprint {
	return b.DropIndex(indexName)
}

// DropForeign drops a foreign key constraint
func (b *Blueprint) DropForeign(constraintName string) *Blueprint {
	b.dropForeigns = append(b.dropForeigns, constraintName)
	return b
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

	// Add indexes - primary keys are always inline, regular indexes are separate for PostgreSQL
	for _, idx := range b.indexes {
		if idx.Type == "primary" || b.dbType != "postgres" {
			sql.WriteString(",\n  " + b.indexToSQL(idx))
		}
	}

	// Add foreign keys
	for _, fk := range b.foreign {
		sql.WriteString(",\n  " + b.foreignKeyToSQL(fk))
	}

	sql.WriteString("\n);")

	// For PostgreSQL, add separate CREATE INDEX statements for non-primary indexes
	if b.dbType == "postgres" {
		for _, idx := range b.indexes {
			if idx.Type != "primary" {
				var indexSQL string
				switch idx.Type {
				case "unique":
					indexSQL = fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s);",
						idx.Name, b.tableName, strings.Join(idx.Columns, ", "))
				default:
					indexSQL = fmt.Sprintf("CREATE INDEX %s ON %s (%s);",
						idx.Name, b.tableName, strings.Join(idx.Columns, ", "))
				}
				sql.WriteString("\n" + indexSQL)
			}
		}
	}

	return sql.String()
}

// toAlterSQL generates ALTER TABLE SQL
func (b *Blueprint) toAlterSQL() string {
	var sqls []string

	// Drop foreign keys
	for _, constraint := range b.dropForeigns {
		var sql string
		switch b.dbType {
		case "mysql":
			sql = fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s;", b.tableName, b.quoteIdentifier(constraint))
		case "postgres":
			sql = fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", b.tableName, b.quoteIdentifier(constraint))
		case "sqlite":
			// SQLite doesn't have named constraints, skip
			continue
		}
		sqls = append(sqls, sql)
	}

	// Drop indexes
	for _, indexName := range b.dropIndexes {
		var sql string
		switch b.dbType {
		case "mysql":
			sql = fmt.Sprintf("DROP INDEX %s ON %s;", b.quoteIdentifier(indexName), b.tableName)
		case "postgres", "sqlite":
			sql = fmt.Sprintf("DROP INDEX %s;", b.quoteIdentifier(indexName))
		}
		sqls = append(sqls, sql)
	}

	// Drop columns
	for _, column := range b.dropColumns {
		sql := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", b.tableName, b.quoteIdentifier(column))
		sqls = append(sqls, sql)
	}

	// Add or modify columns
	for _, col := range b.columns {
		var sql string
		if col.Modify {
			sql = b.getModifyColumnSQL(col)
		} else {
			sql = fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;", b.tableName, b.columnToSQL(col))
		}
		sqls = append(sqls, sql)
	}

	// Add indexes
	for _, idx := range b.indexes {
		// Quote column names
		quotedColumns := make([]string, len(idx.Columns))
		for i, col := range idx.Columns {
			quotedColumns[i] = b.quoteIdentifier(col)
		}

		var sql string
		switch idx.Type {
		case "unique":
			sql = fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s);",
				b.quoteIdentifier(idx.Name), b.tableName, strings.Join(quotedColumns, ", "))
		default:
			sql = fmt.Sprintf("CREATE INDEX %s ON %s (%s);",
				b.quoteIdentifier(idx.Name), b.tableName, strings.Join(quotedColumns, ", "))
		}
		sqls = append(sqls, sql)
	}

	return strings.Join(sqls, "\n")
}

// getModifyColumnSQL generates the appropriate MODIFY/CHANGE column SQL for the database
func (b *Blueprint) getModifyColumnSQL(col Column) string {
	switch b.dbType {
	case "mysql":
		// MySQL uses MODIFY COLUMN
		columnDef := b.columnToSQL(col)
		return fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s;", b.tableName, columnDef)
	case "postgres":
		// PostgreSQL requires multiple statements for different aspects
		var stmts []string

		// Type change
		if col.Type != "" {
			typeSQL := b.getColumnType(col)
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s;",
				b.tableName, b.quoteIdentifier(col.Name), typeSQL))
		}

		// Nullable constraint - Skip for primary key columns (they must be NOT NULL)
		// Primary keys in PostgreSQL cannot have NOT NULL dropped
		if col.Name != "id" { // Common primary key pattern
			if !col.Nullable {
				stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL;",
					b.tableName, b.quoteIdentifier(col.Name)))
			} else {
				stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL;",
					b.tableName, b.quoteIdentifier(col.Name)))
			}
		}

		// Default value
		if col.Default != nil {
			var defaultVal string
			if str, ok := col.Default.(string); ok {
				if strings.ToUpper(str) == "CURRENT_TIMESTAMP" {
					defaultVal = "CURRENT_TIMESTAMP"
				} else {
					defaultVal = fmt.Sprintf("'%s'", strings.ReplaceAll(str, "'", "''"))
				}
			} else {
				defaultVal = fmt.Sprintf("%v", col.Default)
			}
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;",
				b.tableName, b.quoteIdentifier(col.Name), defaultVal))
		} else {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;",
				b.tableName, b.quoteIdentifier(col.Name)))
		}

		return strings.Join(stmts, "\n")
	case "sqlite":
		// SQLite doesn't support ALTER COLUMN in earlier versions
		// Return a comment noting that a manual migration might be needed
		// For newer SQLite (3.26.0+), we can use GENERATED ALWAYS but for compatibility,
		// we'll document this limitation
		return fmt.Sprintf("-- SQLite MODIFY not natively supported. Recommendation: Use raw SQL or recreate table.\n-- To modify %s: Rename table, create new one, copy data, drop old table",
			b.quoteIdentifier(col.Name))
	default:
		columnDef := b.columnToSQL(col)
		return fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s;", b.tableName, columnDef)
	}
}

// columnToSQL converts a column to SQL
func (b *Blueprint) columnToSQL(col Column) string {
	parts := []string{b.quoteIdentifier(col.Name)}

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
		checkConstraint := fmt.Sprintf("CHECK (%s IN (%s))", b.quoteIdentifier(col.Name), strings.Join(quotedValues, ", "))
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
			// Add UNSIGNED modifier for numeric types in MySQL
			if col.Unsigned && b.dbType == "mysql" && b.isNumericType(col.Type) {
				dbType += " UNSIGNED"
			}
			return dbType
		}
	}

	return col.Type // fallback to original type
}

// isNumericType checks if a type is numeric and can be unsigned
func (b *Blueprint) isNumericType(colType string) bool {
	numericTypes := map[string]bool{
		"integer":   true,
		"bigint":    true,
		"tinyint":   true,
		"smallint":  true,
		"mediumint": true,
		"float":     true,
		"double":    true,
		"decimal":   true,
	}
	return numericTypes[colType]
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
	// Quote column names in the index
	quotedColumns := make([]string, len(idx.Columns))
	for i, col := range idx.Columns {
		quotedColumns[i] = b.quoteIdentifier(col)
	}

	switch idx.Type {
	case "primary":
		return fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(quotedColumns, ", "))
	case "unique":
		return fmt.Sprintf("UNIQUE INDEX %s (%s)", b.quoteIdentifier(idx.Name), strings.Join(quotedColumns, ", "))
	default:
		return fmt.Sprintf("INDEX %s (%s)", b.quoteIdentifier(idx.Name), strings.Join(quotedColumns, ", "))
	}
}

// foreignKeyToSQL converts a foreign key to SQL
func (b *Blueprint) foreignKeyToSQL(fk ForeignKey) string {
	sql := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)",
		b.quoteIdentifier(fk.Column), b.quoteIdentifier(fk.On), b.quoteIdentifier(fk.References))

	if fk.OnDelete != "" {
		sql += " ON DELETE " + fk.OnDelete
	}
	if fk.OnUpdate != "" {
		sql += " ON UPDATE " + fk.OnUpdate
	}

	// Add DEFERRABLE clause (PostgreSQL and SQLite support this)
	if fk.Deferrable {
		sql += " DEFERRABLE"
		if fk.InitiallyDeferred {
			sql += " INITIALLY DEFERRED"
		}
	}

	return sql
}
