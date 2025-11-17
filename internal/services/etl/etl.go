// etl implements Service for Extract Transfrom Load pipeline.
package etl

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"database/sql"

	"github.com/gin-gonic/gin"
	gsdmysql "github.com/go-sql-driver/mysql"
	"github.com/sadsonkeenolee/IO_projekt/pkg/database"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
	"github.com/spf13/viper"
)

type Etl struct {
	services.Service
}

func NewEtl() (services.IService, error) {
	e := Etl{}
	l := log.New(os.Stdout, "Etl: ", log.LstdFlags)
	e.Logger = l
	e.Router = gin.Default()

	v, err := e.SetConfig()
	if err != nil {
		e.Logger.Printf("Error: %v.\n", err)
		e.Logger.Println("Service goes into `Idle mode`.")
		e.Logger.Println("Please note the service configuration is incomplete.")
		e.ConfigReader = nil
		e.State = services.StateIdle
		return &e, err
	}
	e.ConfigReader = v

	var ci services.ConnInfo
	if err := e.ConfigReader.UnmarshalKey("ConnInfo", &ci); err != nil {
		l.Fatalln(err)
	}

	e.ConnInfo = &ci
	cfg := gsdmysql.NewConfig()
	cfg.User = ci.Username
	cfg.Passwd = ci.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%v:%v", ci.Ip, ci.Port)
	cfg.DBName = ci.Name

	db, err := sql.Open(ci.Type, cfg.FormatDSN())
	if err != nil {
		e.Logger.Printf("Error: %v.\n", err)
		e.Logger.Println("Service goes into `Idle mode`.")
		e.Logger.Println("Please note the service configuration is incomplete.")
		e.State = services.StateIdle
		return &e, err
	}
	e.DB = db
	return &e, nil
}

func (e *Etl) IsDataExtracted() ([]*database.DatasetFileMetadata, error) {
	rows, err := e.DB.Query(`SELECT directory, data_type, is_read, created_at, read_at FROM read_table`)
	if err != nil {
		e.Logger.Println(err)
		return nil, err
	}
	defer rows.Close()

	dfms := []*database.DatasetFileMetadata{}
	for rows.Next() {
		dfm := &database.DatasetFileMetadata{}
		if err := rows.Scan(&dfm.Directory, &dfm.Type, &dfm.IsRead,
			&dfm.CreatedAt, &dfm.ReadAt); err != nil {
			e.Logger.Printf("Error occured: %v.\n", err)
			continue
		}
		if !dfm.IsRead {
			dfms = append(dfms, dfm)
		}
	}
	return dfms, nil
}

func CsvStreamer(path *string, l *log.Logger, c chan<- []string) error {
	if l == nil {
		panic("CsvStreamer: No logging instance provided.")
	}

	fd, err := os.Open(*path)
	if err != nil {
		return err
	}
	defer fd.Close()
	csvReader := csv.NewReader(fd)

	// read until EOF
	for {
		record, err := csvReader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			l.Printf("Error while parsing: %v", err)
			continue
		}

		c <- record
	}
	close(c)
	return nil
}

// Extract parses stream and puts data into meaningful structure.
// Function requires full path to a file and defined extractor functions that
// will perform other extraction operations in order provided by the caller.
func (e *Etl) Extract(path *string, extractors ...func(stream *[]string, opts *any)) ([]*any, error) {
	c := make(chan []string)
	go CsvStreamer(path, e.Logger, c)
	var extractedStructs []*any = make([]*any, 1)
	var structsTracker uint = 0

	for stream := range c {
		var data any
		for _, extractor := range extractors {
			extractor(&stream, &data)
		}
		extractedStructs[structsTracker] = &data
		extractedStructs = append(extractedStructs, extractedStructs[structsTracker])
		structsTracker++
	}
	return extractedStructs, nil
}

// Transform applies transformations to the given structs
func (e *Etl) Transform(left, right *[]*any, opts ...func(*[]*any, *[]*any, *[]*any)) ([]*any, error) {
	var transformedStructs []*any
	for _, transform := range opts {
		transform(left, right, &transformedStructs)
	}
	return transformedStructs, nil
}

func (e *Etl) Load(final *[]*any, loaders ...func(*any)) error {
	for _, loader := range loaders {
		for _, el := range *final {
			loader(el)
		}
	}
	return nil
}

func (e *Etl) Start() error {
	dfms, err := e.IsDataExtracted()
	if err != nil {
		e.Logger.Printf("Error: %v.\n", err)
		e.Logger.Println("Service goes into `Idle mode`.")
		e.Logger.Println("Please note the service configuration is incomplete.")
		e.State = services.StateIdle
	}

	if err == nil && dfms != nil {
		// Extract data here
		e.Logger.Println("Extracting data for the first ever setup...")
		path1 := filepath.Join(os.Getenv("DOWNLOAD_DIR"), "tmdb-movies-data/tmdb_5000_movies.csv")
		path2 := filepath.Join(os.Getenv("DOWNLOAD_DIR"), "tmdb-movies-data/tmdb_5000_credits.csv")

		// TODO: Dodac zwracanie errorow
		mm1, _ := e.Extract(&path1, database.TmdbMapFromStream)
		mm2, _ := e.Extract(&path2, database.TmdbMapCreditsFromStream)
		final, _ := e.Transform(&mm1, &mm2, database.TmdbJoinBoth)
		_ = e.Load(&final,
			database.InsertIntoMovies(e.DB),
			database.InsertIntoLanguages(e.DB),
			database.InsertIntoKeywords(e.DB),
			database.InsertIntoGenres(e.DB),
			database.InsertIntoCountries(e.DB),
			database.InsertIntoCompanies(e.DB),
			database.InsertIntoMovie2Companies(e.DB),
			database.InsertIntoMovie2Countries(e.DB),
			database.InsertIntoMovie2Genres(e.DB),
			database.InsertIntoMovie2Keywords(e.DB),
			database.InsertIntoMovie2Languages(e.DB),
		)
	}
	// v1 of api.
	{
		v1 := e.Router.Group("/v1")
		v1.POST("api/sync", func(ctx *gin.Context) { panic("Not implemented") })
		v1.POST("api/", func(ctx *gin.Context) { panic("Not implemented") })
	}

	go func() {
		if err := e.Router.Run(":9998"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalf("Router failed: %v\n", err)
		}
	}()

	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
	sig := <-kill
	e.Logger.Printf("Gracefully shutting down the server: %v\n.", sig)
	e.State = services.StateDown
	e.DB.Close()
	return fmt.Errorf("server closed")
}

func (e *Etl) SetConfig() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName("EtlConfig")
	v.SetConfigType("toml")
	v.AddConfigPath(os.Getenv("ETL_CONFIG_DIR_PATH"))
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}

// ExposeConnInfo exposes configuration.
func (e *Etl) ExposeConnInfo() *services.ConnInfo {
	return e.ConnInfo
}

func (e *Etl) String() string {
	return "ETL"
}
