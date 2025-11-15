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
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	gsdmysql "github.com/go-sql-driver/mysql"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
	"github.com/spf13/viper"
)

const (
	NoInformation string = "No information"
)

type DatasetFileMetadata struct {
	Directory string
	Type      string
	IsRead    bool
	CreatedAt *string
	ReadAt    *string
}

type MovieMetadataExtracted struct {
	Budget              uint32
	Genres              map[uint32]string
	MovieId             string
	Keywords            map[uint32]string
	OriginalLanguage    string
	Title               string
	Overview            string
	Popularity          float32
	ProductionCompanies map[uint32]string
	ProductionCountries map[string]string
	ReleaseDate         time.Time
	Revenue             int32
	SpokenLanguages     map[string]string
	Status              string
	Tagline             string
	AverageScore        float32
	TotalScore          uint64
	Cast                *map[uint32]string
	Crew                *map[uint32]string
}

func (dfm *DatasetFileMetadata) String() string {
	if dfm.ReadAt == nil {
		ni := NoInformation
		dfm.ReadAt = &ni
	}
	return fmt.Sprintf(`
		DatasetFileMetadata { 
			Directory: %v
			Type: %v
			Read?: %v
			CreatedAt: %v
			ReadAt: %v
		}`,
		dfm.Directory, dfm.Type, dfm.IsRead, *dfm.CreatedAt, *dfm.ReadAt)
}

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

func (e *Etl) IsDataExtracted() ([]*DatasetFileMetadata, error) {
	rows, err := e.DB.Query(`SELECT directory, data_type, is_read, created_at, read_at FROM read_table`)
	if err != nil {
		e.Logger.Println(err)
		return nil, err
	}
	defer rows.Close()

	dfms := []*DatasetFileMetadata{}
	for rows.Next() {
		dfm := &DatasetFileMetadata{}
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
func (e *Etl) Extract(path *string, extractors ...func(stream *[]string, opts ...*any)) ([]*any, error) {
	c := make(chan []string)
	go CsvStreamer(path, e.Logger, c)
	// TODO: powinna byc raczej mapa typu {movie_id: Metadane}
	var dataStructure []*any

	for stream := range c {
		var data any
		for _, extractor := range extractors {
			extractor(&stream, &data)
		}
		dataStructure = append(dataStructure, &data)
	}
	return dataStructure, nil
}

func (e *Etl) Start() error {
	dfms, err := e.IsDataExtracted()
	if err != nil {
		e.Logger.Printf("Error: %v.\n", err)
		e.Logger.Println("Service goes into `Idle mode`.")
		e.Logger.Println("Please note the service configuration is incomplete.")
		e.State = services.StateIdle
	}

	// var idToGenre map[uint]string
	if err == nil && dfms != nil {
		// Extract data here
		e.Logger.Println("Extracting data for the first ever setup...")
		path := filepath.Join(os.Getenv("DOWNLOAD_DIR"), "tmdb-movies-data/tmdb_5000_credits.csv")
		// na razie to nic nie robi
		data, _ := e.Extract(&path)
		e.Logger.Printf("Extracted: %v", len(data))
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
