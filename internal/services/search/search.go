// Package search implements logic for user requests and responses.
package search

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/sadsonkeenolee/IO_projekt/pkg/database"
	"github.com/sadsonkeenolee/IO_projekt/pkg/services"
	"github.com/spf13/viper"
)

type SearchService struct {
	services.Service
}

type UriContent[T uint64 | string] struct {
	Content T `uri:"indentifier"`
}

var GlobalSearchLogger *log.Logger = log.New(os.Stderr, "", log.LstdFlags|log.Lmsgprefix|log.Llongfile)

func WithLogger(l *log.Logger) func(s *SearchService) {
	return func(s *SearchService) {
		s.Logger = l
	}
}

func WithRouter(ge *gin.Engine) func(s *SearchService) {
	return func(s *SearchService) {
		s.Router = ge
	}
}

func WithViper(v *viper.Viper) func(s *SearchService) {
	return func(s *SearchService) {
		s.ConfigReader = v
	}
}

func WithConnectionInfo(c *services.Connection) func(s *SearchService) {
	return func(s *SearchService) {
		s.ConnInfo = c
	}
}

func WithDatabase(db *sql.DB) func(s *SearchService) {
	return func(s *SearchService) {
		s.DB = db
	}
}

// SearchBuilder implements builder constructor for the search service.
func SearchBuilder(opts ...func(*SearchService)) services.IService {
	f := &SearchService{}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (s *SearchService) Start() error {
	if err := s.HealthCheck(); err != nil {
		GlobalSearchLogger.Fatalf("HealthCheck failed, reason: %v\n", err)
	}
	// v1 of api.
	{
		v1 := s.Router.Group("/v1")
		v1.GET("api/tv/id/:identifier/", s.TvById)
		v1.GET("api/book/id/:identifier/", s.BookById)
		v1.GET("api/concert/id/:identifier/", func(ctx *gin.Context) { panic("Not implemented") })

		v1.GET("api/tv/title/:identifier/", s.TvByTitle)
		v1.GET("api/book/title/:identifier/", s.BookByTitle)
		v1.GET("api/concert/title/:identifier/", func(ctx *gin.Context) { panic("Not implemented") })

		v1.GET("api/", func(ctx *gin.Context) { panic("Not implemented") })
	}
	go func() {
		if err := s.Router.Run(":9997"); err != nil && err != http.ErrServerClosed {
			s.Logger.Fatalf("Router failed: %v\n", err)
		}
	}()
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
	sig := <-kill
	s.Logger.Printf("Gracefully shutting down the server: %v\n.", sig)
	s.State = services.StateDown
	s.DB.Close()
	return fmt.Errorf("server closed")
}

func (s *SearchService) HealthCheck() error {
	if s.Logger == nil {
		return fmt.Errorf("No logger setup")
	}

	if s.Router == nil {
		return fmt.Errorf("No router setup")
	}

	if s.DB == nil {
		return fmt.Errorf("No database setup")
	}

	if s.ConnInfo == nil {
		return fmt.Errorf("No connection info setup")
	}

	if s.ConfigReader == nil {
		return fmt.Errorf("No config setup")
	}

	return nil
}

// TvByTitle gets tv by the title
func (s *SearchService) TvByTitle(ctx *gin.Context) {
	var uc UriContent[string]
	uc.Content = ctx.Param("identifier")

	if uc.Content == "" {
		s.Logger.Println("couldn't parse title")
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	var id uint64
	if err := s.DB.QueryRow(`call find_movie_id(?)`, uc.Content).Scan(&id); err != nil {
		s.Logger.Printf("no id for title %v\n", uc.Content)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	var ms database.MovieSelectable
	err := s.DB.QueryRow("CALL get_movie_by_id(?)", id).Scan(
		&ms.Id, &ms.Budget, &ms.MovieId, &ms.OriginalLanguage,
		&ms.Title, &ms.Overview, &ms.Popularity, &ms.ReleaseDate,
		&ms.Revenue, &ms.Runtime, &ms.Status, &ms.Tagline, &ms.AverageScore,
		&ms.TotalScore,
	)

	if err != nil {
		s.Logger.Printf("Lookup failed for ID %v: %v\n", id, err)
		services.NewBadContentRequest(ctx, "movie doesn't exist")
		return
	}

	services.NewGoodContentRequest(ctx, ms)
}

// TvById gets tv by id
func (s *SearchService) TvById(ctx *gin.Context) {
	var uc UriContent[uint64]
	uc.Content, _ = strconv.ParseUint(ctx.Param("identifier"), 10, 64)

	// 1. Fetch the main movie data
	var ms database.MovieSelectable
	err := s.DB.QueryRow("CALL get_movie_by_id(?)", uc.Content).Scan(
		&ms.Id, &ms.Budget, &ms.MovieId, &ms.OriginalLanguage,
		&ms.Title, &ms.Overview, &ms.Popularity, &ms.ReleaseDate,
		&ms.Revenue, &ms.Runtime, &ms.Status, &ms.Tagline, &ms.AverageScore,
		&ms.TotalScore,
	)

	if err != nil {
		s.Logger.Printf("Lookup failed for ID %v: %v\n", uc.Content, err)
		services.NewBadContentRequest(ctx, "movie doesn't exist")
		return
	}

	services.NewGoodContentRequest(ctx, ms)
}

// BookById gets book by id
func (s *SearchService) BookById(ctx *gin.Context) {
	var uc UriContent[uint64]
	uc.Content, _ = strconv.ParseUint(ctx.Param("identifier"), 10, 64)

	var bs database.BookSelectable
	err := s.DB.QueryRow("CALL get_book_by_id(?)", uc.Content).Scan(
		&bs.Id, &bs.Title, &bs.Isbn, &bs.Isbn13, &bs.Language,
		&bs.Pages, &bs.ReleaseDate, &bs.Publisher, &bs.Rating, &bs.TotalRating,
	)

	if err != nil {
		s.Logger.Printf("could not find book ID %v: %v\n", bs.BookId, err)
		services.NewBadContentRequest(ctx, "book doesn't exist")
		return
	}

	services.NewGoodContentRequest(ctx, bs)
}

// BookByTitle gets book by title
func (s *SearchService) BookByTitle(ctx *gin.Context) {
	var uc UriContent[string]
	uc.Content = ctx.Param("identifier")

	var bookID uint64
	err := s.DB.QueryRow("CALL find_book_id(?)", uc.Content).Scan(&bookID)
	if err != nil {
		s.Logger.Printf("could not find book with title %v: %v\n", uc.Content, err)
		services.NewBadContentRequest(ctx, "book not found")
		return
	}

	var bs database.BookSelectable
	err = s.DB.QueryRow("CALL get_book_by_id(?)", bookID).Scan(
		&bs.Id, &bs.Title, &bs.Isbn, &bs.Isbn13, &bs.Language,
		&bs.Pages, &bs.ReleaseDate, &bs.Publisher, &bs.Rating, &bs.TotalRating,
	)
	if err != nil {
		s.Logger.Println(err)
		services.NewBadContentRequest(ctx, "error fetching book details")
		return
	}

	services.NewGoodContentRequest(ctx, bs)
}

// ExposeConnection exposes configuration.
func (s *SearchService) ExposeConnection() *services.Connection {
	return s.ConnInfo
}

func (s *SearchService) String() string {
	return "Fetch"
}
