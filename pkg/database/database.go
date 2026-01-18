package database

import (
	"database/sql"
	_ "database/sql"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
)

var DatabaseLogger *log.Logger = log.New(os.Stderr, "", log.LstdFlags|log.Lmsgprefix|log.Llongfile)

const (
	ItemNotFound    string = "no information"
	InvalidPipeline string = "invalid interface (no InsertPipeline)"
)

type DatasetFileMetadata struct {
	Directory string
	Type      string
	CreatedAt *string
	ReadAt    bool
}

// Table defines tables present in the database
type Table struct {
	// Table name
	Name string
	// Query fields (?, ?, ..., ?)
	QueryField string
	// Table fields (ID, etc.)
	Fields []string
}

func NewTable(name string, fields []string) *Table {
	return &Table{
		Name:       name,
		Fields:     fields,
		QueryField: InitQueryFields(len(fields)),
	}
}

// InsertPipeline defines pipeline for inserting elements to the given database.
type InsertPipeline struct {
	// Track which data was inserted
	Tracker int
	Data    *[]*Insertable
}

func NewInsertPipeline(data *[]*Insertable) (Insertable, error) {
	if len(*data) <= 0 {
		return nil, fmt.Errorf("no data (length %v)", len(*data))
	}
	return &InsertPipeline{
		Tracker: 0,
		Data:    data,
	}, nil
}

// Next checks if there is a next item in a pipeline, if true returns the
// element
func (ip *InsertPipeline) Next() *Insertable {
	if ip.Tracker+1 >= len(*ip.Data) {
		return nil
	}
	ip.Tracker++
	return (*ip.Data)[ip.Tracker]
}

func (ip *InsertPipeline) Reset() {
	ip.Tracker = 0
}

func (ip *InsertPipeline) IsInsertable() (*Table, bool) {
	// Pipeline itself is not insertable, however its element are.
	return nil, true
}

func (ip *InsertPipeline) ConstructInsertQuery() string {
	return ""
}

type Insertable interface {
	// IsInsertable tells if the item is insertable and returns table information
	IsInsertable() (*Table, bool)
	// This should include only statement and table fields, e.g.:
	// INSERT INTO table(ID, revenue, ...) VALUES
	ConstructInsertQuery() string
}

type Selectable interface {
	// IsSelectable tells if the item is selectable and returns a query how to
	// select it.
	IsSelectable() (*Table, bool)
	// Only SELECT is permitted, no statements like INSERT, etc.
	ConstructSelectQuery() string
}

func InitQueryFields(cnt int) string {
	return fmt.Sprintf("(%v)", strings.Join(slices.Repeat([]string{"?"}, cnt), ","))
}

func JoinTableFields(t *Table) string {
	return fmt.Sprintf("(%v)", strings.Join(t.Fields, ","))

}

func RebuildTable(db *sql.DB, table, engine string) error {
	DatabaseLogger.Printf("Currently rebuilding table: %v\n", table)
	_, err := db.Exec(fmt.Sprintf("ALTER TABLE %v ENGINE = %v", table, engine))
	if err != nil {
		return err
	}
	return nil
}
