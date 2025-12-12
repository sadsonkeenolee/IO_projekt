package database

import (
	"database/sql"
	_ "database/sql"
	"fmt"
	"log"
	"os"

	gsdmysql "github.com/go-sql-driver/mysql"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
)

var DatabaseLogger *log.Logger = log.New(os.Stderr, "", log.LstdFlags|log.Lmsgprefix|log.Llongfile)

const (
	ItemNotFound    string = "no information"
	InvalidPipeline string = "invalid interface (no InsertPipeline)"
)

type DatasetFileMetadata struct {
	Directory string
	Type      string
	IsRead    bool
	CreatedAt *string
	ReadAt    *string
}

// Query defines reusable data structure for making database queries
type Query struct {
	Content string
	Fields  string
}

// InsertPipeline defines pipeline for inserting elements to the given database.
type InsertPipeline struct {
	Query
	Tracker int
	Data    *[]*Insertable
}

func NewInsertPipeline(q Query, data *[]*Insertable) (Insertable, error) {
	if len(q.Content) == 0 {
		return nil, fmt.Errorf("incorrect query (length 0)")
	}
	if len(q.Fields) == 0 {
		return nil, fmt.Errorf("incorrect fields (length 0)")
	}
	if len(*data) == 0 {
		return nil, fmt.Errorf("no data (length 0)")
	}

	return &InsertPipeline{
		Query:   q,
		Tracker: 0,
		Data:    data,
	}, nil
}

// Next checks if there is a next item in a pipeline, if true returns the
// element
func (ip *InsertPipeline) Next() (*Insertable, bool) {
	if ip.Tracker+1 >= len(*ip.Data) {
		return nil, false
	}
	ip.Tracker++
	return (*ip.Data)[ip.Tracker], true
}

func (ip *InsertPipeline) Reset() {
	ip.Tracker = 0
}

// func (ip *InsertPipeline) ChunkData(chunkSize int) [][]*Insertable {
// 	var chunked [][]*Insertable
// 	for idx := 0; idx < len(*ip.Data); idx += chunkSize {
// 		end := min(len(*ip.Data), idx+chunkSize)
// 		chunked = append(chunked, (*ip.Data)[idx:end])
// 	}
// 	return chunked
// }

func (ip *InsertPipeline) IsInsertable() bool {
	// cannot insert pipeline, but its elements
	return false
}

type Insertable interface {
	// Marker interface for a struct that is an Insertable item.
	IsInsertable() bool
}

type Queryable interface {
	// Marker interface for a struct that is a Queryable item.
	IsQueryable() bool
}

func ParseDriverConfig(ci *services.ConnInfo) *gsdmysql.Config {
	cfg := gsdmysql.NewConfig()
	cfg.User = ci.Username
	cfg.Passwd = ci.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%v:%v", ci.Ip, ci.Port)
	cfg.DBName = ci.Name
	return cfg
}

func RebuildTable(db *sql.DB, table, engine string) error {
	DatabaseLogger.Printf("Currently rebuilding table: %v\n", table)
	// FIXME: Dziwny bug, wczesniej dzialalo
	_, err := db.Exec(fmt.Sprintf("ALTER TABLE %v ENGINE = %v", table, engine))
	if err != nil {
		return err
	}
	return nil
}
