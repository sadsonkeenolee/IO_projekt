// ingest implements Service for Extract Transfrom Load pipeline.
package ingest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
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
		if dfm.ReadAt == false {
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

func (i *Ingest) InsertMoviePipeline(movies *[]*database.Insertable) error {
	ip, err := database.NewInsertPipeline(movies)
	if err != nil {
		return fmt.Errorf("Error while creating a pipeline: %v\n", err)
	}

	loaders := []LoadFunc{
		database.InsertIntoMoviesChunked(i.DB, &i.MaxBatchSize),
		database.InsertIntoLanguagesChunked(i.DB, &i.MaxBatchSize),
		database.InsertIntoKeywordsChunked(i.DB, &i.MaxBatchSize),
		database.InsertIntoGenresChunked(i.DB, &i.MaxBatchSize),
		database.InsertIntoCountriesChunked(i.DB, &i.MaxBatchSize),
		database.InsertIntoCompaniesChunked(i.DB, &i.MaxBatchSize),
		database.InsertIntoMovie2CompaniesChunked(i.DB, &i.MaxBatchSize),
		database.InsertIntoMovie2GenresChunked(i.DB, &i.MaxBatchSize),
		database.InsertIntoMovie2KeywordsChunked(i.DB, &i.MaxBatchSize),
		database.InsertIntoMovie2LanguagesChunked(i.DB, &i.MaxBatchSize),
		database.InsertIntoMovie2CountriesChunked(i.DB, &i.MaxBatchSize),
	}

	// Try to load whatever is possible, ignore the errors
	for _, loader := range loaders {
		if err := i.Load(&ip, loader); err != nil {
			i.Logger.Printf("Error while loading the data, reason: %v\n", err)
		}
	}
	return nil
}

func (i *Ingest) RebuildAllTables() {
	tables := []string{"movies", "languages", "keywords", "genres", "countries",
		"movie2companies", "movie2countries", "movie2genres", "movie2keywords",
		"movie2languages", "books", "authors",
	}
	// Try to rebuild table, ignore failures
	for _, table := range tables {
		if err := database.RebuildTable(i.DB, table, "InnoDB"); err != nil {
			i.Logger.Printf("Error while rebuilding table, reason: %v\n", err)
		}
	}
}

func (i *Ingest) FirstMovieLoadPipeline(path1, path2 string) {
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
	i.InsertMoviePipeline(&final)
}

func (i *Ingest) BookInsertPipeline(books *[]*database.Insertable) error {
	bip, err := database.NewInsertPipeline(books)
	if err != nil {
		return fmt.Errorf("cannot create pipeline, %v\n", err)
	}

	loaders := []LoadFunc{
		database.InsertIntoBooksChunked(i.DB, &i.MaxBatchSize),
		database.InsertIntoAuthorsChunked(i.DB, &i.MaxBatchSize),
	}

	for _, loader := range loaders {
		if err := i.Load(&bip, loader); err != nil {
			i.Logger.Printf("Error while loading books, reason: %v\n", err)
		}
	}
	return nil
}

func (i *Ingest) FirstBookLoadPipeline(path3 string) {
	bis, err := i.Extract(&path3, database.BookFromStream)
	if err != nil {
		i.Logger.Fatalf("Third extraction failed: %v\n", err)
	}
	i.Logger.Println("Third extraction completed.")
	i.BookInsertPipeline(&bis)
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

		// Pipeline starts here
		{
			start := time.Now()
			i.FirstMovieLoadPipeline(path1, path2)
			i.FirstBookLoadPipeline(path3)
			tx, err := i.DB.Begin()
			updateStmt, err := tx.Prepare(`update read_table set read_at=1 where read_at=0`)
			if err != nil {
				i.Logger.Printf("couldn't update read_table, reason: %v\n", err)
			}
			defer updateStmt.Close()
			if _, err := updateStmt.Exec(); err != nil {
				i.Logger.Printf("couldn't update read_table, reason: %v\n", err)
			}
			elapsed := time.Since(start)
			i.RebuildAllTables()
			i.Logger.Printf("Loading completed: %v.\n", elapsed)
		}
		// Pipeline ends here
	}
	// v1 of api.
	{
		v1 := i.Router.Group("/v1")
		v1.POST("api/ingest/:identifier", i.NewTvRecord)
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

// FetchTvDataFromWeb fetches basic data about the movies that the title matches
func (i *Ingest) FetchTvDataFromWeb(title string) (*utils.TmdbResponseContentSchema, error) {
	url := utils.SanitizeUrl(os.Getenv("TMDB_FETCH_MOVIE_BASIC"), title)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", os.Getenv("TMDB_API_KEY")))
	body, err := utils.DoRequest(req)
	if err != nil {
		return nil, err
	}
	var respJson utils.TmdbResponseContentSchema
	if err := json.Unmarshal(body, &respJson); err != nil {
		return nil, err
	}
	return &respJson, nil
}

// FetchTvDataFromWebSpecific fetches all data about the movie that the title matches
func (i *Ingest) FetchTvDataFromWebSpecific(id uint64) (*utils.TmdbTvSchema, error) {
	url := utils.SanitizeUrl(os.Getenv("TMDB_FETCH_MOVIE_SPECIFIC"), strconv.FormatUint(id, 10))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", os.Getenv("TMDB_API_KEY")))
	body, err := utils.DoRequest(req)
	if err != nil {
		return nil, err
	}
	var m utils.TmdbTvSchema
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (i *Ingest) NewTvRecord(ctx *gin.Context) {
	var uc services.UriContent[string]
	uc.Content = ctx.Param("identifier")
	res, err := i.FetchTvDataFromWeb(uc.Content)
	if err != nil {
		i.Logger.Printf("couldn't fetch the basic data %v\n", err)
		return
	}
	sort.Slice(
		res.Results,
		func(i, j int) bool {
			return res.Results[i].Popularity > res.Results[j].Popularity
		})
	relevantID := res.Results[0].Id
	tmdbSchema, err := i.FetchTvDataFromWebSpecific(relevantID)
	if err != nil {
		i.Logger.Printf("couldn't fetch the specific data %v\n", err)
		return
	}
	mi := tmdbSchema.IntoMovieInsertable()
	if mi == nil {
		i.Logger.Printf("couldn't transform into MovieInsertable.\n")
		return
	}
	ins, err := database.CastFromMovieInsertableToInsertable(mi)
	if err != nil {
		i.Logger.Println(err)
		return
	}
	mis := []*database.Insertable{&ins}
	i.InsertMoviePipeline(&mis)
}

// ExposeConnection exposes configuration.
func (i *Ingest) ExposeConnection() *services.Connection {
	return i.ConnInfo
}

func (i *Ingest) String() string {
	return "Etl"
}
