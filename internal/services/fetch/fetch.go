package fetch

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/sadsonkeenolee/IO_projekt/pkg/database"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
	"github.com/spf13/viper"
)

type Fetch struct {
	services.Service
}

type Id struct {
	Content uint64 `uri:"id" binding:"required"`
}

type Title struct {
	Content string `uri:"title" binding:"required"`
}

var FetchLogger *log.Logger = log.New(os.Stderr, "", log.LstdFlags|log.Lmsgprefix|log.Llongfile)

func WithLogger(out io.Writer, prefix string, flags int) func(f *Fetch) {
	return func(f *Fetch) {
		if flags&log.Llongfile == log.Llongfile {
			FetchLogger.Fatalln("log.Llongfile flag is not allowed for the service purposes.")
		}
		if out == os.Stderr {
			FetchLogger.Fatalln("`os.Stderr` io.Writer is not allowed for the service purposes.")
		}
		f.Logger = log.New(out, prefix, flags)
	}
}

func WithRouter(opts ...gin.OptionFunc) func(f *Fetch) {
	return func(f *Fetch) {
		f.Router = gin.Default(opts...)
	}
}

func WithConfig(filename, ext string, cfgPaths ...string) func(f *Fetch) {
	return func(f *Fetch) {
		v := viper.New()
		v.SetConfigName(filename)
		v.SetConfigType(ext)
		for _, cfgPath := range cfgPaths {
			v.AddConfigPath(cfgPath)
		}

		if err := v.ReadInConfig(); err != nil {
			FetchLogger.Fatalf("Got error while parsing config file: %v\n", err)
		}
		f.ConfigReader = v
	}
}

func WithConnectionInfo(tableName string) func(f *Fetch) {
	return func(f *Fetch) {
		if f.ConfigReader == nil {
			FetchLogger.Fatalln("Before parsing a connection info, initialize your config reader.")
		}

		var ci services.ConnInfo
		if err := f.ConfigReader.UnmarshalKey(tableName, &ci); err != nil {
			FetchLogger.Fatalf("Got error while unmarshalling: %v\n", err)
		}
		f.ConnInfo = &ci
	}
}

// TODO: Add additional parameters
func WithDatabase() func(f *Fetch) {
	return func(f *Fetch) {
		cfgParsed := database.ParseDriverConfig(f.ConnInfo)
		db, err := sql.Open(f.ConnInfo.Type, fmt.Sprintf("%v?parseTime=true", cfgParsed.FormatDSN()))
		if err != nil {
			FetchLogger.Fatalf("Got error while creating a database driver: %v\n", err)
		}
		f.DB = db
	}
}

func FetchBuilder(opts ...func(*Fetch)) services.IService {
	f := &Fetch{}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (f *Fetch) Start() error {
	if err := f.HealthCheck(); err != nil {
		FetchLogger.Fatalf("HealthCheck failed, reason: %v\n", err)
	}
	// v1 of api.
	{
		v1 := f.Router.Group("/v1")
		// Id z bazy danych, zwraca konkretne id
		v1.GET("api/tv/id/:id/", f.TvById)
		v1.GET("api/book/id/:id/", func(ctx *gin.Context) { panic("Not implemented") })
		v1.GET("api/concert/id/:id/", func(ctx *gin.Context) { panic("Not implemented") })

		// Zwraca wszystkie rekordy z pasujacym tytulem
		v1.GET("api/tv/title/:title/", f.TvByTitle)
		v1.GET("api/book/title/:title/", func(ctx *gin.Context) { panic("Not implemented") })
		v1.GET("api/concert/title/:title/", func(ctx *gin.Context) { panic("Not implemented") })

		v1.GET("api/", func(ctx *gin.Context) { panic("Not implemented") })
	}
	go func() {
		if err := f.Router.Run(":9997"); err != nil && err != http.ErrServerClosed {
			f.Logger.Fatalf("Router failed: %v\n", err)
		}
	}()
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
	sig := <-kill
	f.Logger.Printf("Gracefully shutting down the server: %v\n.", sig)
	f.State = services.StateDown
	f.DB.Close()
	return fmt.Errorf("server closed")
}

func (f *Fetch) HealthCheck() error {
	if f.Logger == nil {
		return fmt.Errorf("No logger setup")
	}

	if f.Router == nil {
		return fmt.Errorf("No router setup")
	}

	if f.DB == nil {
		return fmt.Errorf("No database setup")
	}

	if f.ConnInfo == nil {
		return fmt.Errorf("No connection info setup")
	}

	if f.ConfigReader == nil {
		return fmt.Errorf("No config setup")
	}

	return nil
}

func (f *Fetch) TvByTitle(ctx *gin.Context) {
	bind := Title{Content: ""}
	if err := ctx.ShouldBindUri(&bind); err != nil {
		f.Logger.Println(err)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}
	if bind.Content == "" {
		f.Logger.Println("couldn't parse title")
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	rows, err := f.DB.Query(`SELECT ID, budget, tmdb_id, original_lang_id,
		title, overview, popularity, release_date, revenue, runtime, status, 
		tagline, vote_average, vote_total FROM movies WHERE title LIKE ?`, "%"+bind.Content+"%")
	if err != nil {
		f.Logger.Printf("couldn't find a movie with title=%v\n", bind.Content)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}
	defer rows.Close()
	mqs := []database.MovieQueryable{}
	for rows.Next() {
		var mq database.MovieQueryable
		if err := rows.Scan(
			&mq.Id, &mq.Budget, &mq.MovieId, &mq.OriginalLanguage,
			&mq.Title, &mq.Overview, &mq.Popularity, &mq.ReleaseDate,
			&mq.Revenue, &mq.Runtime, &mq.Status, &mq.Tagline, &mq.AverageScore,
			&mq.TotalScore); err != nil {
			f.Logger.Println(err)
			continue
		}
		mqs = append(mqs, mq)
	}
	services.NewGoodContentRequest(ctx, mqs)
}

func (f *Fetch) TvById(ctx *gin.Context) {
	bind := Id{}
	if err := ctx.ShouldBindUri(&bind); err != nil {
		f.Logger.Println(err)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	rows, err := f.DB.Query(`SELECT ID, budget, tmdb_id, original_lang_id,
		title, overview, popularity, release_date, revenue, runtime, status, 
		tagline, vote_average, vote_total  FROM movies WHERE ID=?`, bind.Content)
	if err != nil {
		f.Logger.Printf("couldn't find a movie with ID=%v\n", bind.Content)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}
	var mq database.MovieQueryable
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&mq.Id, &mq.Budget, &mq.MovieId, &mq.OriginalLanguage,
			&mq.Title, &mq.Overview, &mq.Popularity, &mq.ReleaseDate,
			&mq.Revenue, &mq.Runtime, &mq.Status, &mq.Tagline, &mq.AverageScore,
			&mq.TotalScore); err != nil {
			f.Logger.Println(err)
			services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
			return
		}
	}

	if mq.Id != bind.Content {
		f.Logger.Printf("id (%v) mismatches the found id (%v)\n",
			bind.Content, mq.Id)
		services.NewBadContentRequest(ctx, "movie doesn't exist")
		return
	}
	services.NewGoodContentRequest(ctx, mq)
}

// ExposeConnInfo exposes configuration.
func (f *Fetch) ExposeConnInfo() *services.ConnInfo {
	return f.ConnInfo
}

func (f *Fetch) String() string {
	return "Fetch"
}
