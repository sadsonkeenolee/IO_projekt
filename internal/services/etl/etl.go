// etl implements Service for Extract Transfrom Load pipeline.
package etl

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/sadsonkeenolee/IO_projekt/pkg/database"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
	"github.com/sadsonkeenolee/IO_projekt/pkg/utils"
	"github.com/spf13/viper"
)

// Use it as global logger that will log everything that happens globally
var EtlLogger *log.Logger = log.New(os.Stderr, "", log.LstdFlags|log.Lmsgprefix|log.Llongfile)

type ExtractFunc = func(*[]string, *database.Insertable) error
type TransformFunc = func(*[]*database.Insertable, *[]*database.Insertable, *[]*database.Insertable) error
type LoadFunc = func(*database.Insertable) error

type Etl struct {
	services.Service
	MaxBatchSize int
}

func WithLogger(out io.Writer, prefix string, flags int) func(e *Etl) {
	return func(e *Etl) {
		if flags&log.Llongfile == log.Llongfile {
			EtlLogger.Fatalln("`log.Llongfile` flag is not allowed for the service purposes.")
		}
		if out == os.Stderr {
			EtlLogger.Fatalln("`os.Stderr` io.Writer is not allowed for the service purposes.")
		}
		e.Logger = log.New(out, prefix, flags)
	}
}

func WithRouter(opts ...gin.OptionFunc) func(e *Etl) {
	return func(e *Etl) {
		e.Router = gin.Default(opts...)
	}
}

func WithConfig(filename, ext string, cfgPaths ...string) func(e *Etl) {
	return func(e *Etl) {
		v := viper.New()
		v.SetConfigName(filename)
		v.SetConfigType(ext)
		for _, cfgPath := range cfgPaths {
			v.AddConfigPath(cfgPath)
		}

		if err := v.ReadInConfig(); err != nil {
			EtlLogger.Fatalf("Got error while parsing config file: %v\n", err)
		}
		e.ConfigReader = v
	}
}

func WithConnectionInfo(tableName string) func(e *Etl) {
	return func(e *Etl) {
		if e.ConfigReader == nil {
			EtlLogger.Fatalln("Before parsing a connection info, initialize your config reader.")
		}

		var ci services.ConnInfo
		if err := e.ConfigReader.UnmarshalKey(tableName, &ci); err != nil {
			EtlLogger.Fatalf("Got error while unmarshalling: %v\n", err)
		}
		e.ConnInfo = &ci
	}
}

func WithDatabase() func(e *Etl) {
	return func(e *Etl) {
		cfgParsed := database.ParseDriverConfig(e.ConnInfo)
		db, err := sql.Open(e.ConnInfo.Type, cfgParsed.FormatDSN())
		if err != nil {
			EtlLogger.Fatalf("Got error while creating a database driver: %v\n", err)
		}
		e.DB = db
	}
}

func WithBatchSize(n int) func(e *Etl) {
	return func(e *Etl) {
		if n < 0 {
			EtlLogger.Fatalf("Got batch size of size (%v), expected n > 0", n)
		}
		e.MaxBatchSize = n
	}
}

func EtlBuilder(opts ...func(*Etl)) services.IService {
	e := &Etl{}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Etl) IsDataExtracted() ([]*database.DatasetFileMetadata, error) {
	rows, err := e.DB.Query(`SELECT directory, data_type, is_read, created_at, 
		read_at FROM read_table`)
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

	go utils.CsvStreamer(path, EtlLogger, c)
	for stream := range c {
		var data database.Insertable
		var shouldAppend bool = true
		for _, extractor := range funcs {
			err := extractor(&stream, &data)
			if err != nil {
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
			e.Logger.Printf("One of the transformations failed: %v\n", err)
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
	if err := e.HealthCheck(); err != nil {
		EtlLogger.Fatalf("HealthCheck failed, reason: %v\n", err)
	}

	dfms, err := e.IsDataExtracted()
	if err != nil {
		e.Logger.Printf("Error: %v.\n", err)
		e.Logger.Println("Service goes into `Idle mode`.")
		e.Logger.Println("Please note the service configuration is incomplete.")
		e.State = services.StateIdle
	}

	if err == nil && len(dfms) != 0 {
		// Extract data here
		e.Logger.Println("Extracting data for the first ever setup...")
		path1 := filepath.Join(os.Getenv("DOWNLOAD_DIR"), "tmdb-movies-data/tmdb_5000_movies.csv")
		path2 := filepath.Join(os.Getenv("DOWNLOAD_DIR"), "tmdb-movies-data/tmdb_5000_credits.csv")

		start := time.Now()
		// MOVIE PIPELINE START
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

		mip, err := database.NewInsertPipeline(database.MovieInsertQuery, &final)
		err = e.Load(&mip, database.InsertIntoMoviesChunked(e.DB, &e.MaxBatchSize))
		if err != nil {
			e.Logger.Fatalf("Error while inserting: %v\n", err)
		}

		lip, err := database.NewInsertPipeline(database.LanguagesInsertQuery, &final)
		err = e.Load(&lip, database.InsertIntoLanguagesChunked(e.DB, &e.MaxBatchSize))
		if err != nil {
			e.Logger.Fatalf("Error while inserting: %v\n", err)
		}

		kip, err := database.NewInsertPipeline(database.KeywordsInsertQuery, &final)
		err = e.Load(&kip, database.InsertIntoKeywordsChunked(e.DB, &e.MaxBatchSize))
		if err != nil {
			e.Logger.Fatalf("Error while inserting: %v\n", err)
		}

		gip, err := database.NewInsertPipeline(database.GenreInsertQuery, &final)
		err = e.Load(&gip, database.InsertIntoGenresChunked(e.DB, &e.MaxBatchSize))
		if err != nil {
			e.Logger.Fatalf("Error while inserting: %v\n", err)
		}

		couip, err := database.NewInsertPipeline(database.CountryInsertQuery, &final)
		err = e.Load(&couip, database.InsertIntoCountriesChunked(e.DB, &e.MaxBatchSize))
		if err != nil {
			e.Logger.Fatalf("Error while inserting: %v\n", err)
		}

		comip, err := database.NewInsertPipeline(database.CompananyInsertQuery, &final)
		err = e.Load(&comip, database.InsertIntoCompaniesChunked(e.DB, &e.MaxBatchSize))
		if err != nil {
			e.Logger.Fatalf("Error while inserting: %v\n", err)
		}

		m2cip, err := database.NewInsertPipeline(database.Movie2CompaniesInsertQuery, &final)
		err = e.Load(&m2cip, database.InsertIntoMovie2CompaniesChunked(e.DB, &e.MaxBatchSize))
		if err != nil {
			e.Logger.Fatalf("Error while inserting: %v\n", err)
		}

		m2coip, err := database.NewInsertPipeline(database.Movie2CountriesInsertQuery, &final)
		err = e.Load(&m2coip, database.InsertIntoCompaniesChunked(e.DB, &e.MaxBatchSize))
		if err != nil {
			e.Logger.Fatalf("Error while inserting: %v\n", err)
		}

		m2gip, err := database.NewInsertPipeline(database.Movie2GenresInsertQuery, &final)
		err = e.Load(&m2gip, database.InsertIntoMovie2GenresChunked(e.DB, &e.MaxBatchSize))
		if err != nil {
			e.Logger.Fatalf("Error while inserting: %v\n", err)
		}

		m2kip, err := database.NewInsertPipeline(database.Movie2KeywordsInsertQuery, &final)
		err = e.Load(&m2kip, database.InsertIntoMovie2KeywordsChunked(e.DB, &e.MaxBatchSize))
		if err != nil {
			e.Logger.Fatalf("Error while inserting: %v\n", err)
		}

		m2lip, err := database.NewInsertPipeline(database.Movie2LanguagesInsertQuery, &final)
		err = e.Load(&m2lip, database.InsertIntoMovie2LanguagesChunked(e.DB, &e.MaxBatchSize))
		if err != nil {
			e.Logger.Fatalf("Error while inserting: %v\n", err)
		}

		// MOVIE PIPELINE END
		// BOOK PIPELINE START

		path3 := filepath.Join(os.Getenv("DOWNLOAD_DIR"), "goodreads-books-data/books.csv")
		bis, err := e.Extract(&path3, database.BookFromStream)
		if err != nil {
			e.Logger.Fatalf("First extraction failed: %v\n", err)
		}
		e.Logger.Println("Third extraction completed.")

		bisp, err := database.NewInsertPipeline(database.BookInsertQuery, &bis)
		if err != nil {
			e.Logger.Printf("Cannot create pipeline, %v\n", err)
		}

		basp, err := database.NewInsertPipeline(database.BookAuthorsQuery, &bis)
		if err != nil {
			e.Logger.Printf("Cannot create pipeline, %v\n", err)
		}

		e.Load(&bisp, database.InsertIntoBooksChunked(e.DB, &e.MaxBatchSize))
		e.Load(&basp, database.InsertIntoAuthorsChunked(e.DB, &e.MaxBatchSize))

		if err = database.RebuildTable(e.DB, "movies", "InnoDB"); err != nil {
			e.Logger.Fatalf("Error while rebuilding movies: %v\n", err)
		}
		if err = database.RebuildTable(e.DB, "languages", "InnoDB"); err != nil {
			e.Logger.Fatalf("Error while rebuilding languages: %v\n", err)
		}
		if err = database.RebuildTable(e.DB, "keywords", "InnoDB"); err != nil {
			e.Logger.Fatalf("Error while rebuilding keywords: %v\n", err)
		}
		if err = database.RebuildTable(e.DB, "genres", "InnoDB"); err != nil {
			e.Logger.Fatalf("Error while rebuilding genres: %v\n", err)
		}
		if err = database.RebuildTable(e.DB, "countries", "InnoDB"); err != nil {
			e.Logger.Fatalf("Error while rebuilding countries: %v\n", err)
		}
		if err = database.RebuildTable(e.DB, "movie2companies", "InnoDB"); err != nil {
			e.Logger.Fatalf("Error while rebuilding movie2companies: %v\n", err)
		}
		if err = database.RebuildTable(e.DB, "movie2countries", "InnoDB"); err != nil {
			e.Logger.Fatalf("Error while rebuilding movie2countries: %v\n", err)
		}
		if err = database.RebuildTable(e.DB, "movie2genres", "InnoDB"); err != nil {
			e.Logger.Fatalf("Error while rebuilding movie2genres: %v\n", err)
		}
		if err = database.RebuildTable(e.DB, "movie2keywords", "InnoDB"); err != nil {
			e.Logger.Fatalf("Error while rebuilding movie2keywords: %v\n", err)
		}
		if err = database.RebuildTable(e.DB, "movie2languages", "InnoDB"); err != nil {
			e.Logger.Fatalf("Error while rebuilding movie2languages: %v\n", err)
		}

		if err = database.RebuildTable(e.DB, "books", "InnoDB"); err != nil {
			e.Logger.Fatalf("Error while rebuilding movies: %v\n", err)
		}
		if err = database.RebuildTable(e.DB, "authors", "InnoDB"); err != nil {
			e.Logger.Fatalf("Error while rebuilding movies: %v\n", err)
		}

		tx, err := e.DB.Begin()
		updateStmt, err := tx.Prepare(`update read_table set is_read=1,
		 read_at=current_timestamp where is_read=0`)
		if err != nil {
			e.Logger.Println(err)
		}
		defer updateStmt.Close()
		if _, err := updateStmt.Exec(); err != nil {
			e.Logger.Println(err)
		}
		elapsed := time.Since(start)
		e.Logger.Printf("Loading completed: %v.\n", elapsed)
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

func (e *Etl) HealthCheck() error {
	if e.Logger == nil {
		return fmt.Errorf("No logger setup")
	}

	if e.Router == nil {
		return fmt.Errorf("No router setup")
	}

	if e.DB == nil {
		return fmt.Errorf("No database setup")
	}

	if e.ConnInfo == nil {
		return fmt.Errorf("No connection info setup")
	}

	if e.ConfigReader == nil {
		return fmt.Errorf("No config setup")
	}

	if e.MaxBatchSize <= 0 {
		return fmt.Errorf("Incorrect batch size")
	}
	return nil
}

// ExposeConnInfo exposes configuration.
func (e *Etl) ExposeConnInfo() *services.ConnInfo {
	return e.ConnInfo
}

func (e *Etl) String() string {
	return "Etl"
}
