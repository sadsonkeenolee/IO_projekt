package fetch

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"database/sql"

	"github.com/gin-gonic/gin"
	gsdmysql "github.com/go-sql-driver/mysql"
	"github.com/sadsonkeenolee/IO_projekt/pkg/database"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
	"github.com/spf13/viper"
)

type Fetch struct {
	services.Service
}

type Id struct {
	Content int `uri:"id" binding:"required"`
}

type Title struct {
	Content string `uri:"title" binding:"required"`
}

func NewFetch() (services.IService, error) {
	f := Fetch{}
	l := log.New(os.Stdout, "Fetch: ", log.LstdFlags)
	f.Logger = l
	f.Router = gin.Default()

	v, err := f.SetConfig()
	if err != nil {
		f.Logger.Printf("Error: %v.\n", err)
		f.Logger.Println("Service goes into `Idle mode`.")
		f.Logger.Println("Please note the service configuration is incomplete.")
		f.ConfigReader = nil
		f.State = services.StateIdle
		return &f, err
	}
	f.ConfigReader = v

	var ci services.ConnInfo
	if err := f.ConfigReader.UnmarshalKey("ConnInfo", &ci); err != nil {
		l.Fatalln(err)
	}

	f.ConnInfo = &ci
	cfg := gsdmysql.NewConfig()
	cfg.User = ci.Username
	cfg.Passwd = ci.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%v:%v", ci.Ip, ci.Port)
	cfg.DBName = ci.Name

	db, err := sql.Open(ci.Type, fmt.Sprintf("%v?multiStatements=true&parseTime=true", cfg.FormatDSN()))
	if err != nil {
		f.Logger.Printf("Error: %v.\n", err)
		f.Logger.Println("Service goes into `Idle mode`.")
		f.Logger.Println("Please note the service configuration is incomplete.")
		f.State = services.StateIdle
		return &f, err
	}
	f.DB = db
	return &f, nil
}

func (f *Fetch) Start() error {
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

func (f *Fetch) TvByTitle(ctx *gin.Context) {
	bind := Title{Content: ""}
	if err := ctx.ShouldBindUri(&bind); err != nil {
		f.Logger.Println(err)
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}
	if bind.Content == "" {
		f.Logger.Println("couldn't parse title")
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}

	rows, err := f.DB.Query(`SELECT ID, budget, tmdb_id, original_lang_id,
		title, overview, popularity, release_date, revenue, runtime, status, 
		tagline, vote_average, vote_total FROM movies WHERE title LIKE ?`, "%"+bind.Content+"%")
	if err != nil {
		f.Logger.Println(err)
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}
	defer rows.Close()
	mqs := []database.MovieQueryable{}
	// TODO: Dodac parsowanie reszty pol
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
	ctx.JSON(
		http.StatusFound,
		gin.H{
			"content": mqs,
		})
}

func (f *Fetch) TvById(ctx *gin.Context) {
	bind := Id{-1}
	if err := ctx.ShouldBindUri(&bind); err != nil {
		f.Logger.Println(err)
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}
	if bind.Content == -1 {
		f.Logger.Println("couldn't parse id")
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}

	rows, err := f.DB.Query(`SELECT ID, budget, tmdb_id, original_lang_id,
		title, overview, popularity, release_date, revenue, runtime, status, 
		tagline, vote_average, vote_total  FROM movies WHERE ID=?`, bind.Content)
	if err != nil {
		f.Logger.Println(err)
		services.NewBadCredentialsCoreResponse(ctx, services.InvalidRequestMessage)
		return
	}
	defer rows.Close()
	mqs := []database.MovieQueryable{}
	// TODO: Dodac parsowanie reszty pol
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
	ctx.JSON(
		http.StatusFound,
		gin.H{
			"content": mqs,
		})
}

func (f *Fetch) SetConfig() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName("FetchConfig")
	v.SetConfigType("toml")
	v.AddConfigPath(os.Getenv("FETCH_CONFIG_DIR_PATH"))
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}

// ExposeConnInfo exposes configuration.
func (f *Fetch) ExposeConnInfo() *services.ConnInfo {
	return f.ConnInfo
}

func (f *Fetch) String() string {
	return "Fetch"
}
