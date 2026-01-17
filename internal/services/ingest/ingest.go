// ingest implements Service for Extract Transfrom Load pipeline.
package ingest

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
	"github.com/sadsonkeenolee/IO_projekt/pkg/database"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
	"github.com/sadsonkeenolee/IO_projekt/pkg/utils"
	"github.com/spf13/viper"
)

// Use it as global logger that will log everything that happens globally
var GlobalIngestLogger *log.Logger = log.New(os.Stderr, "",
	log.LstdFlags|log.Lmsgprefix|log.Llongfile)

// Define function signatures for functions.
type ExtractFunc = func(*[]string, *database.Insertable) error
type TransformFunc = func(*[]*database.Insertable, *[]*database.Insertable,
	*[]*database.Insertable) error
type LoadFunc = func(*database.Insertable) error

type Ingest struct {
	services.Service
	MaxBatchSize int
}

func WithLogger(l *log.Logger) func(i *Ingest) {
	return func(i *Ingest) {
		i.Logger = l
	}
}

func WithRouter(ge *gin.Engine) func(i *Ingest) {
	return func(i *Ingest) {
		i.Router = ge
	}
}

func WithViper(v *viper.Viper) func(i *Ingest) {
	return func(i *Ingest) {
		i.ConfigReader = v
	}
}

func WithConnectionInfo(ci *services.Connection) func(i *Ingest) {
	return func(i *Ingest) {
		i.ConnInfo = ci
	}
}

func WithDatabase(db *sql.DB) func(i *Ingest) {
	return func(i *Ingest) {
		i.DB = db
	}
}

func WithBatch(n int) func(i *Ingest) {
	return func(i *Ingest) {
		if n <= 0 {
			i.MaxBatchSize = 128
			return
		}
		i.MaxBatchSize = n
	}
}

func IngestBuilder(opts ...func(*Ingest)) services.IService {
	i := &Ingest{}
	for _, opt := range opts {
		opt(i)
	}
	return i
}

func (i *Ingest) IsDataExtracted() ([]*database.DatasetFileMetadata, error) {
	rows, err := i.DB.Query(`SELECT directory, data_type, created_at, 
		read_at FROM read_table`)
	if err != nil {
		i.Logger.Println(err)
		return nil, err
	}
	defer rows.Close()

	dfms := []*database.DatasetFileMetadata{}
	for rows.Next() {
		dfm := &database.DatasetFileMetadata{}
		if err := rows.Scan(&dfm.Directory, &dfm.Type, &dfm.CreatedAt,
			&dfm.ReadAt); err != nil {
			i.Logger.Printf("Error occured: %v.\n", err)
			continue
		}
		if dfm.ReadAt == nil {
			dfms = append(dfms, dfm)
		}
	}
	return dfms, nil
}

// Extract parses stream and puts data into meaningful structure.
// Function requires full path to a file and defined extractor functions that
// will perform other extraction operations in order provided by the caller.
func (i *Ingest) Extract(path *string, funcs ...ExtractFunc) ([]*database.Insertable, error) {
	var c chan []string = make(chan []string)
	var extractedData []*database.Insertable
	go utils.CsvStreamer(path, GlobalIngestLogger, c)
	for stream := range c {
		var data database.Insertable
		var shouldAppend bool = true
		for _, extractor := range funcs {
			err := extractor(&stream, &data)
			if err != nil {
				i.Logger.Printf("Got error while extracting: %v\n", err)
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
func (i *Ingest) Transform(left, right *[]*database.Insertable, funcs ...TransformFunc) ([]*database.Insertable, error) {
	var transformedData []*database.Insertable
	for _, transform := range funcs {
		if err := transform(left, right, &transformedData); err != nil {
			i.Logger.Printf("One of the transformations failed: %v\n", err)
		}
	}
	return transformedData, nil
}

func (i *Ingest) Load(pipeline *database.Insertable, funcs ...LoadFunc) error {
	for _, loader := range funcs {
		loader(pipeline)
	}
	return nil
}

func (i *Ingest) MoviePipeline(path1, path2 string) {
	mme1, err := i.Extract(&path1, database.TmdbMapFromStream)
	if err != nil {
		i.Logger.Fatalf("First extraction failed: %v\n", err)
	}
	i.Logger.Println("First extraction completed.")
	mme2, err := i.Extract(&path2, database.TmdbMapCreditsFromStream)
	if err != nil {
		i.Logger.Fatalf("Second extraction failed: %v\n", err)
	}
	i.Logger.Println("Second extraction completed.")
	final, err := i.Transform(&mme1, &mme2, database.TmdbJoinBoth)
	if err != nil {
		i.Logger.Fatalf("Transforming failed: %v\n", err)
	}
	i.Logger.Println("Transforming completed.")

	ip, err := database.NewInsertPipeline(&final)
	if err != nil {
		i.Logger.Fatalf("Error while creating a pipeline: %v\n", err)
	}

	err = i.Load(&ip, database.InsertIntoMoviesChunked(i.DB, &i.MaxBatchSize))
	if err != nil {
		i.Logger.Fatalf("Error while inserting: %v\n", err)
	}

	err = i.Load(&ip, database.InsertIntoLanguagesChunked(i.DB, &i.MaxBatchSize))
	if err != nil {
		i.Logger.Fatalf("Error while inserting: %v\n", err)
	}

	err = i.Load(&ip, database.InsertIntoKeywordsChunked(i.DB, &i.MaxBatchSize))
	if err != nil {
		i.Logger.Fatalf("Error while inserting: %v\n", err)
	}

	err = i.Load(&ip, database.InsertIntoGenresChunked(i.DB, &i.MaxBatchSize))
	if err != nil {
		i.Logger.Fatalf("Error while inserting: %v\n", err)
	}

	err = i.Load(&ip, database.InsertIntoCountriesChunked(i.DB, &i.MaxBatchSize))
	if err != nil {
		i.Logger.Fatalf("Error while inserting: %v\n", err)
	}

	err = i.Load(&ip, database.InsertIntoCompaniesChunked(i.DB, &i.MaxBatchSize))
	if err != nil {
		i.Logger.Fatalf("Error while inserting: %v\n", err)
	}

	err = i.Load(&ip, database.InsertIntoMovie2CompaniesChunked(i.DB, &i.MaxBatchSize))
	if err != nil {
		i.Logger.Fatalf("Error while inserting: %v\n", err)
	}

	err = i.Load(&ip, database.InsertIntoCompaniesChunked(i.DB, &i.MaxBatchSize))
	if err != nil {
		i.Logger.Fatalf("Error while inserting: %v\n", err)
	}

	err = i.Load(&ip, database.InsertIntoMovie2GenresChunked(i.DB, &i.MaxBatchSize))
	if err != nil {
		i.Logger.Fatalf("Error while inserting: %v\n", err)
	}

	err = i.Load(&ip, database.InsertIntoMovie2KeywordsChunked(i.DB, &i.MaxBatchSize))
	if err != nil {
		i.Logger.Fatalf("Error while inserting: %v\n", err)
	}

	err = i.Load(&ip, database.InsertIntoMovie2LanguagesChunked(i.DB, &i.MaxBatchSize))
	if err != nil {
		i.Logger.Fatalf("Error while inserting: %v\n", err)
	}

	err = i.Load(&ip, database.InsertIntoMovie2CountriesChunked(i.DB, &i.MaxBatchSize))
	if err != nil {
		i.Logger.Fatalf("Error while inserting: %v\n", err)
	}

	if err = database.RebuildTable(i.DB, "movies", "InnoDB"); err != nil {
		i.Logger.Fatalf(services.TableRebuildMessage, "movies", err)
	}
	if err = database.RebuildTable(i.DB, "languages", "InnoDB"); err != nil {
		i.Logger.Fatalf(services.TableRebuildMessage, "languages", err)
	}
	if err = database.RebuildTable(i.DB, "keywords", "InnoDB"); err != nil {
		i.Logger.Fatalf(services.TableRebuildMessage, "keywords", err)
	}
	if err = database.RebuildTable(i.DB, "genres", "InnoDB"); err != nil {
		i.Logger.Fatalf(services.TableRebuildMessage, "genres", err)
	}
	if err = database.RebuildTable(i.DB, "countries", "InnoDB"); err != nil {
		i.Logger.Fatalf(services.TableRebuildMessage, "countries", err)
	}
	if err = database.RebuildTable(i.DB, "movie2companies", "InnoDB"); err != nil {
		i.Logger.Fatalf(services.TableRebuildMessage, "movie2companies", err)
	}
	if err = database.RebuildTable(i.DB, "movie2countries", "InnoDB"); err != nil {
		i.Logger.Fatalf(services.TableRebuildMessage, "movie2countries", err)
	}
	if err = database.RebuildTable(i.DB, "movie2genres", "InnoDB"); err != nil {
		i.Logger.Fatalf(services.TableRebuildMessage, "movie2genres", err)
	}
	if err = database.RebuildTable(i.DB, "movie2keywords", "InnoDB"); err != nil {
		i.Logger.Fatalf(services.TableRebuildMessage, "movie2keywords", err)
	}
	if err = database.RebuildTable(i.DB, "movie2languages", "InnoDB"); err != nil {
		i.Logger.Fatalf(services.TableRebuildMessage, "movie2languages", err)
	}
}

func (i *Ingest) BookPipeline(path3 string) {
	bis, err := i.Extract(&path3, database.BookFromStream)
	if err != nil {
		i.Logger.Fatalf("First extraction failed: %v\n", err)
	}
	i.Logger.Println("Third extraction completed.")

	bip, err := database.NewInsertPipeline(&bis)
	if err != nil {
		i.Logger.Printf("Cannot create pipeline, %v\n", err)
	}
	i.Load(&bip, database.InsertIntoBooksChunked(i.DB, &i.MaxBatchSize))
	i.Load(&bip, database.InsertIntoAuthorsChunked(i.DB, &i.MaxBatchSize))
	if err = database.RebuildTable(i.DB, "books", "InnoDB"); err != nil {
		i.Logger.Fatalf(services.TableRebuildMessage, "books", err)
	}
	if err = database.RebuildTable(i.DB, "authors", "InnoDB"); err != nil {
		i.Logger.Fatalf(services.TableRebuildMessage, "authors", err)
	}
}

func (i *Ingest) Start() error {
	if err := i.HealthCheck(); err != nil {
		GlobalIngestLogger.Fatalf("HealthCheck failed, reason: %v\n", err)
	}

	dfms, err := i.IsDataExtracted()
	if err != nil {
		i.Logger.Printf("Error: %v.\n", err)
		i.Logger.Println("Service goes into `Idle mode`.")
		i.Logger.Println("Please note the service configuration is incomplete.")
		i.State = services.StateIdle
	}

	if err == nil && len(dfms) != 0 {
		// Extract data here
		i.Logger.Println("Extracting data for the first ever setup...")
		path1 := filepath.Join(os.Getenv("DOWNLOAD_DIR"), "tmdb-movies-data/tmdb_5000_movies.csv")
		path2 := filepath.Join(os.Getenv("DOWNLOAD_DIR"), "tmdb-movies-data/tmdb_5000_credits.csv")
		path3 := filepath.Join(os.Getenv("DOWNLOAD_DIR"), "goodreads-books-data/books.csv")
		start := time.Now()
		i.MoviePipeline(path1, path2)
		i.BookPipeline(path3)
		tx, err := i.DB.Begin()
		updateStmt, err := tx.Prepare(`update read_table set read_at=current_timestamp`)
		if err != nil {
			i.Logger.Println(err)
		}
		defer updateStmt.Close()
		if _, err := updateStmt.Exec(); err != nil {
			i.Logger.Println(err)
		}
		elapsed := time.Since(start)
		i.Logger.Printf("Loading completed: %v.\n", elapsed)
	}
	// v1 of api.
	{
		v1 := i.Router.Group("/v1")
		v1.POST("api/sync", func(ctx *gin.Context) { panic("Not implemented") })
		v1.POST("api/", func(ctx *gin.Context) { panic("Not implemented") })
	}

	go func() {
		if err := i.Router.Run(":9998"); err != nil && err != http.ErrServerClosed {
			i.Logger.Fatalf("Router failed: %v\n", err)
		}
	}()

	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
	sig := <-kill
	i.Logger.Printf("Gracefully shutting down the server: %v\n.", sig)
	i.State = services.StateDown
	i.DB.Close()
	return fmt.Errorf("server closed")
}

func (i *Ingest) HealthCheck() error {
	if i.Logger == nil {
		return fmt.Errorf("No logger setup")
	}

	if i.Router == nil {
		return fmt.Errorf("No router setup")
	}

	if i.DB == nil {
		return fmt.Errorf("No database setup")
	}

	if i.ConnInfo == nil {
		return fmt.Errorf("No connection info setup")
	}

	if i.ConfigReader == nil {
		return fmt.Errorf("No config setup")
	}

	if i.MaxBatchSize <= 0 {
		return fmt.Errorf("Incorrect batch size")
	}
	return nil
}

// ExposeConnection exposes configuration.
func (i *Ingest) ExposeConnection() *services.Connection {
	return i.ConnInfo
}

func (i *Ingest) String() string {
	return "Etl"
}
