// etl implements Service for Extract Transfrom Load pipeline.
package etl

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	gsdmysql "github.com/go-sql-driver/mysql"
	"github.com/sadsonkeenolee/IO_projekt/pkg/database"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
	"github.com/sadsonkeenolee/IO_projekt/pkg/utils"
	"github.com/spf13/viper"
)

type ExtractFunc = func(*[]string, *database.Insertable) error
type TransformFunc = func(*[]*database.Insertable, *[]*database.Insertable, *[]*database.Insertable) error
type LoadFunc = func(*database.Insertable) error

type Etl struct {
	services.Service
	MaxBatchSize int
}

func NewEtl(maxBatchSize int) (services.IService, error) {
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
	e.MaxBatchSize = maxBatchSize
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

// Extract parses stream and puts data into meaningful structure.
// Function requires full path to a file and defined extractor functions that
// will perform other extraction operations in order provided by the caller.
func (e *Etl) Extract(path *string, funcs ...ExtractFunc) ([]*database.Insertable, error) {
	// https://dev.mysql.com/doc/refman/8.4/en/optimizing-innodb-bulk-data-loading.html
	// https://dev.mysql.com/doc/refman/8.4/en/server-system-variables.html#sysvar_bulk_insert_buffer_size
	// https://dev.mysql.com/doc/refman/8.4/en/load-data.html
	// https://dev.mysql.com/doc/refman/8.4/en/insert-optimization.html
	var c chan []string = make(chan []string)
	var extractedData []*database.Insertable

	go utils.CsvStreamer(path, e.Logger, c)
	for stream := range c {
		var data database.Insertable
		var shouldAppend bool = true
		for _, extractor := range funcs {
			err := extractor(&stream, &data)
			if err != nil {
				// TODO: dodac tworzenie errorow
				e.Logger.Printf("Got error while extracting: %v\n", err)
				shouldAppend = false
				break
			}
		}

		if shouldAppend {
			extractedData = append(extractedData, &data)
		}
	}
	return extractedData, nil
}

// Transform applies transformations to the given data. This function will fail
// and return nil, because data might be malformed. fns variable should embed
// the database connection.
func (e *Etl) Transform(left, right *[]*database.Insertable, funcs ...TransformFunc) ([]*database.Insertable, error) {
	var transformedData []*database.Insertable
	for _, transform := range funcs {
		if err := transform(left, right, &transformedData); err != nil {
			return nil, fmt.Errorf("transformations failed: %v\n", err)
		}
	}
	return transformedData, nil
}

func (e *Etl) Load(pipeline *database.Insertable, funcs ...LoadFunc) error {
	for _, loader := range funcs {
		loader(pipeline)
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

		start := time.Now()
		mme1, err := e.Extract(&path1, database.TmdbMapFromStream)
		if err != nil {
			e.Logger.Fatalf("First extraction failed: %v\n", err)
		}
		e.Logger.Println("First extraction completed.")
		mme2, err := e.Extract(&path2, database.TmdbMapCreditsFromStream)
		if err != nil {
			e.Logger.Fatalf("Second extraction failed: %v\n", err)
		}
		e.Logger.Println("Second extraction completed.")
		final, err := e.Transform(&mme1, &mme2, database.TmdbJoinBoth)
		if err != nil {
			e.Logger.Fatalf("Transforming failed: %v\n", err)
		}
		e.Logger.Println("Transforming completed.")
		// define the general pipeline, but use optimized functions
		pipeline, err := database.NewInsertPipeline("IGNORE", "IGNORE", &final)
		err = e.Load(&pipeline,
			database.InsertIntoMoviesChunked(e.DB, &e.MaxBatchSize),
			database.InsertIntoLanguagesChunked(e.DB, &e.MaxBatchSize),
			database.InsertIntoKeywordsChunked(e.DB, &e.MaxBatchSize),
			database.InsertIntoGenresChunked(e.DB, &e.MaxBatchSize),
			database.InsertIntoCountriesChunked(e.DB, &e.MaxBatchSize),
			database.InsertIntoCompaniesChunked(e.DB, &e.MaxBatchSize),
			database.InsertIntoMovie2CompaniesChunked(e.DB, &e.MaxBatchSize),
			database.InsertIntoMovie2CountriesChunked(e.DB, &e.MaxBatchSize),
			database.InsertIntoMovie2GenresChunked(e.DB, &e.MaxBatchSize),
			database.InsertIntoMovie2KeywordsChunked(e.DB, &e.MaxBatchSize),
			database.InsertIntoMovie2LanguagesChunked(e.DB, &e.MaxBatchSize),
		)
		if err != nil {
			e.Logger.Fatalf("Loading failed: %v\n", err)
		}
		elapsed := time.Since(start)
		e.Logger.Printf("Loading completed: %vs.\n", elapsed)
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
