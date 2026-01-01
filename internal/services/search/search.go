package search

import (
	"fmt"
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

type SearchService struct {
	services.Service
}

type Id struct {
	Content uint64 `uri:"id" binding:"required"`
}

type Title struct {
	Content string `uri:"title" binding:"required"`
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
		// Id z bazy danych, zwraca konkretne id
		v1.GET("api/tv/id/:id/", s.TvById)
		v1.GET("api/book/id/:id/", s.BookById)
		v1.GET("api/concert/id/:id/", func(ctx *gin.Context) { panic("Not implemented") })

		// Zwraca wszystkie rekordy z pasujacym tytulem
		v1.GET("api/tv/title/:title/", s.TvByTitle)
		v1.GET("api/book/title/:title/", s.BookByTitle)
		v1.GET("api/concert/title/:title/", func(ctx *gin.Context) { panic("Not implemented") })

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

func (s *SearchService) TvByTitle(ctx *gin.Context) {
	bind := Title{Content: ""}
	if err := ctx.ShouldBindUri(&bind); err != nil {
		s.Logger.Println(err)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}
	if bind.Content == "" {
		s.Logger.Println("couldn't parse title")
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	var id uint64
	if err := s.DB.QueryRow(`call find_movie_id(?)`, bind.Content).Scan(&id); err != nil {
		s.Logger.Printf("no id for title %v\n", bind.Content)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	// 1. Fetch the main movie data
	var mq database.MovieSelectable
	err := s.DB.QueryRow("CALL get_movie_by_id(?)", id).Scan(
		&mq.Id, &mq.Budget, &mq.MovieId, &mq.OriginalLanguage,
		&mq.Title, &mq.Overview, &mq.Popularity, &mq.ReleaseDate,
		&mq.Revenue, &mq.Runtime, &mq.Status, &mq.Tagline, &mq.AverageScore,
		&mq.TotalScore,
	)

	if err != nil {
		s.Logger.Printf("Lookup failed for ID %v: %v\n", id, err)
		services.NewBadContentRequest(ctx, "movie doesn't exist")
		return
	}

	s.fillMetadata(id, "CALL get_production_companies(?)", &mq.ProductionCompanies)
	s.fillMetadata(id, "CALL get_genres(?)", &mq.Genres)
	s.fillMetadata(id, "CALL get_keywords(?)", &mq.Keywords)
	s.fillMetadata(id, "CALL get_production_countries(?)", &mq.ProductionCountries)
	s.fillMetadata(id, "CALL get_languages(?)", &mq.SpokenLanguages)
	services.NewGoodContentRequest(ctx, mq)
}

func (s *SearchService) TvById(ctx *gin.Context) {
	bind := Id{}
	if err := ctx.ShouldBindUri(&bind); err != nil {
		s.Logger.Println(err)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	// 1. Fetch the main movie data
	var mq database.MovieSelectable
	err := s.DB.QueryRow("CALL get_movie_by_id(?)", bind.Content).Scan(
		&mq.Id, &mq.Budget, &mq.MovieId, &mq.OriginalLanguage,
		&mq.Title, &mq.Overview, &mq.Popularity, &mq.ReleaseDate,
		&mq.Revenue, &mq.Runtime, &mq.Status, &mq.Tagline, &mq.AverageScore,
		&mq.TotalScore,
	)

	if err != nil {
		s.Logger.Printf("Lookup failed for ID %v: %v\n", bind.Content, err)
		services.NewBadContentRequest(ctx, "movie doesn't exist")
		return
	}

	// 2. Populate the complex slices using the procedures
	// We pass the pointer to the slices directly
	s.fillMetadata(bind.Content, "CALL get_production_companies(?)", &mq.ProductionCompanies)
	s.fillMetadata(bind.Content, "CALL get_genres(?)", &mq.Genres)
	s.fillMetadata(bind.Content, "CALL get_keywords(?)", &mq.Keywords)
	s.fillMetadata(bind.Content, "CALL get_production_countries(?)", &mq.ProductionCountries)
	s.fillMetadata(bind.Content, "CALL get_languages(?)", &mq.SpokenLanguages)

	services.NewGoodContentRequest(ctx, mq)
}

func (s *SearchService) fillMetadata(movieID uint64, query string, target interface{}) {
	rows, err := s.DB.Query(query, movieID)
	if err != nil {
		s.Logger.Printf("Procedure error [%s]: %v\n", query, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var dummyID uint64
		var val string
		if err := rows.Scan(&dummyID, &val); err != nil {
			s.Logger.Println("Scan error:", err)
			continue
		}

		switch v := target.(type) {
		case *[]database.Genre:
			item := database.Genre{}
			item.Id = dummyID
			item.Name = val
			*v = append(*v, item)

		case *[]database.Keywords:
			item := database.Keywords{}
			item.Id = dummyID // If your Keyword struct has an ID field
			item.Name = val
			*v = append(*v, item)

		case *[]database.ProductionCompaniesInsertable:
			item := database.ProductionCompaniesInsertable{}
			item.Id = dummyID
			item.Name = val
			*v = append(*v, item)

		case *[]database.CodeCountry:
			item := database.CodeCountry{}
			item.Name = val
			*v = append(*v, item)

		case *[]database.LanguageEncoding:
			item := database.LanguageEncoding{}
			item.Name = val
			*v = append(*v, item)
		}
	}
}

func (s *SearchService) BookById(ctx *gin.Context) {
	bind := Id{}
	if err := ctx.ShouldBindUri(&bind); err != nil {
		s.Logger.Println(err)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	var bq database.BookSelectable
	// Use the procedure for the main book data
	// Assuming your Scan matches the procedure: ID, title, overview, isbn, isbn13, language, pages, release_date, publisher, rating, total_ratings
	err := s.DB.QueryRow("CALL get_book_by_id(?)", bind.Content).Scan(
		&bq.Id, &bq.Title, &bq.Isbn, &bq.Isbn13, &bq.Language,
		&bq.Pages, &bq.ReleaseDate, &bq.Publisher, &bq.Rating, &bq.TotalRating,
	)

	if err != nil {
		s.Logger.Printf("Could not find book ID %v: %v\n", bind.Content, err)
		services.NewBadContentRequest(ctx, "book doesn't exist")
		return
	}

	// Populate Authors using the helper
	s.fillMetadata(bind.Content, "CALL get_book_authors(?)", &bq.Authors)

	services.NewGoodContentRequest(ctx, bq)
}

func (s *SearchService) BookByTitle(ctx *gin.Context) {
	bind := Title{}
	if err := ctx.ShouldBindUri(&bind); err != nil {
		s.Logger.Println(err)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	var bookID uint64
	// 1. Find the best matching ID via procedure
	err := s.DB.QueryRow("CALL find_book_id(?)", bind.Content).Scan(&bookID)
	if err != nil {
		s.Logger.Printf("Could not find book with title %v: %v\n", bind.Content, err)
		services.NewBadContentRequest(ctx, "book not found")
		return
	}

	// 2. Reuse the ID fetch logic to get the full struct + authors
	var bq database.BookSelectable
	err = s.DB.QueryRow("CALL get_book_by_id(?)", bookID).Scan(
		&bq.Id, &bq.Title, &bq.Isbn, &bq.Isbn13, &bq.Language,
		&bq.Pages, &bq.ReleaseDate, &bq.Publisher, &bq.Rating, &bq.TotalRating,
	)
	if err != nil {
		s.Logger.Println(err)
		services.NewBadContentRequest(ctx, "error fetching book details")
		return
	}

	// 3. Fill the Authors slice
	s.fillMetadata(bookID, "CALL get_book_authors(?)", &bq.Authors)

	services.NewGoodContentRequest(ctx, bq)
}

// ExposeConnection exposes configuration.
func (s *SearchService) ExposeConnection() *services.Connection {
	return s.ConnInfo
}

func (s *SearchService) String() string {
	return "Fetch"
}
