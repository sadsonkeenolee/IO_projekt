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

func (s *SearchService) GetSpokenLanguages(mss ...*database.MovieSelectable) {
	for _, ms := range mss {
		rows, err := s.DB.Query(`call get_languages(?)`, ms.MovieId)
		if err != nil {
			s.Logger.Printf("couldn't fetch spoken languages, reason: %v\n", err)
			return
		}
		// make sure the memory is allocated for the keywords
		ms.SpokenLanguages = make([]database.LanguageEncoding, 0, 16)
		for rows.Next() {
			var le database.LanguageEncoding
			var null any
			if err := rows.Scan(&null, &le.Encoding, &le.Name); err != nil {
				s.Logger.Printf("couldn't scan the genre, reason: %v\n", err)
				continue
			}
			ms.SpokenLanguages = append(ms.SpokenLanguages, le)
		}
	}
}

func (s *SearchService) GetProductionCompanies(mss ...*database.MovieSelectable) {
	for _, ms := range mss {
		rows, err := s.DB.Query(`call get_production_companies(?)`, ms.MovieId)
		if err != nil {
			s.Logger.Printf("couldn't fetch companies, reason: %v\n", err)
			return
		}
		// make sure the memory is allocated for the keywords
		ms.ProductionCompanies = make([]database.ProductionCompaniesInsertable, 0, 16)
		for rows.Next() {
			var pci database.ProductionCompaniesInsertable
			if err := rows.Scan(&pci.IdName.Id, &pci.IdName.Name); err != nil {
				s.Logger.Printf("couldn't scan the genre, reason: %v\n", err)
				continue
			}
			ms.ProductionCompanies = append(ms.ProductionCompanies, pci)
		}
	}
}

func (s *SearchService) GetKeywords(mss ...*database.MovieSelectable) {
	for _, ms := range mss {
		rows, err := s.DB.Query(`call get_keywords(?)`, ms.MovieId)
		if err != nil {
			s.Logger.Printf("couldn't fetch keywords, reason: %v\n", err)
			return
		}
		// make sure the memory is allocated for the keywords
		ms.Keywords = make([]database.Keywords, 0, 16)
		for rows.Next() {
			var k database.Keywords
			if err := rows.Scan(&k.IdName.Id, &k.IdName.Name); err != nil {
				s.Logger.Printf("couldn't scan the genre, reason: %v\n", err)
				continue
			}
			ms.Keywords = append(ms.Keywords, k)
		}
	}
}

func (s *SearchService) GetGenres(mss ...*database.MovieSelectable) {
	for _, ms := range mss {
		rows, err := s.DB.Query(`call get_genres(?)`, ms.MovieId)
		if err != nil {
			s.Logger.Printf("couldn't fetch genres, reason: %v\n", err)
			return
		}
		// make sure the memory is allocated for the genres
		ms.Genres = make([]database.Genre, 0, 16)
		for rows.Next() {
			var g database.Genre
			if err := rows.Scan(&g.IdName.Id, &g.IdName.Name); err != nil {
				s.Logger.Printf("couldn't scan the genre, reason: %v\n", err)
				continue
			}
			ms.Genres = append(ms.Genres, g)
		}
	}
}

func (s *SearchService) GetTop100Shows(ctx *gin.Context) {
	rows, err := s.DB.Query(`select * from top_100_shows`)
	if err != nil {
		s.Logger.Printf("cannot query top 100 shows, reason: %v\n", err)
		return
	}
	defer rows.Close()
	shows := make([]*database.MovieSelectable, 0, 100)

	for rows.Next() {
		var ms database.MovieSelectable
		if err := rows.Scan(
			&ms.Id, &ms.Budget, &ms.MovieId, &ms.OriginalLanguage,
			&ms.Title, &ms.Overview, &ms.Popularity, &ms.ReleaseDate,
			&ms.Revenue, &ms.Runtime, &ms.Status, &ms.Tagline, &ms.AverageScore,
			&ms.TotalScore,
		); err != nil {
			s.Logger.Println(err)
			continue
		}
		shows = append(shows, &ms)
	}
	services.NewGoodContentRequest(ctx, shows)
}

func (s *SearchService) GetTop100Books(ctx *gin.Context) {
	rows, err := s.DB.Query(`select * from top_100_books`)
	if err != nil {
		s.Logger.Printf("cannot query top 100 books, reason: %v\n", err)
		return
	}
	defer rows.Close()
	books := make([]*database.BookSelectable, 0, 100)

	for rows.Next() {
		var bs database.BookSelectable
		if err := rows.Scan(
			&bs.Id, &bs.Title, &bs.Isbn, &bs.Isbn13, &bs.Language,
			&bs.Pages, &bs.ReleaseDate, &bs.Publisher, &bs.Rating, &bs.TotalRating,
		); err != nil {
			fmt.Println(err)
			continue
		}
		books = append(books, &bs)
	}
	services.NewGoodContentRequest(ctx, books)
}

func (s *SearchService) GetDefaultShowsRecommendations() []*database.MovieSelectable {
	rows, err := s.DB.Query(`select * from default_shows_recommendation`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	shows := make([]*database.MovieSelectable, 0, 25)

	for rows.Next() {
		var ms database.MovieSelectable
		if err := rows.Scan(
			&ms.Id, &ms.Budget, &ms.MovieId, &ms.OriginalLanguage,
			&ms.Title, &ms.Overview, &ms.Popularity, &ms.ReleaseDate,
			&ms.Revenue, &ms.Runtime, &ms.Status, &ms.Tagline, &ms.AverageScore,
			&ms.TotalScore,
		); err != nil {
			s.Logger.Println(err)
			continue
		}
		s.GetGenres(&ms)
		s.GetKeywords(&ms)
		s.GetProductionCompanies(&ms)
		s.GetSpokenLanguages(&ms)
		shows = append(shows, &ms)
	}
	return shows
}

func (s *SearchService) GetDefaultBooksRecommendations() []*database.BookSelectable {
	rows, err := s.DB.Query(`select * from default_books_recommendation`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	books := make([]*database.BookSelectable, 0, 25)

	for rows.Next() {
		var bs database.BookSelectable
		if err := rows.Scan(
			&bs.Id, &bs.Title, &bs.Isbn, &bs.Isbn13, &bs.Language,
			&bs.Pages, &bs.ReleaseDate, &bs.Publisher, &bs.Rating, &bs.TotalRating,
		); err != nil {
			fmt.Println(err)
			continue
		}
		books = append(books, &bs)
	}
	return books
}

func (s *SearchService) HomePage(ctx *gin.Context) {
	shows := s.GetDefaultShowsRecommendations()
	books := s.GetDefaultBooksRecommendations()
	content := map[string]any{
		"shows": shows,
		"books": books,
	}
	services.NewGoodContentRequest(ctx, content)
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
		v1.GET("api/home/", s.HomePage)
		v1.GET("api/tv/id/:identifier/", s.TvById)
		v1.GET("api/tv/title/:identifier/", s.TvByTitle)
		v1.GET("api/tv/top100/", s.GetTop100Shows)
		v1.GET("api/book/id/:identifier/", s.BookById)
		v1.GET("api/book/title/:identifier/", s.BookByTitle)
		v1.GET("api/book/top100/", s.GetTop100Books)
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
	var uc services.UriContent[string]
	uc.Content = ctx.Param("identifier")

	if uc.Content == "" {
		s.Logger.Println("couldn't parse title")
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	var id uint64
	if err := s.DB.QueryRow(`call find_movie_id(?)`, uc.Content).Scan(&id); err != nil {
		s.Logger.Printf("no id for title %v\n", uc.Content)
		// call Ingest and then try to select from database
		req, _ := http.NewRequest("POST",
			fmt.Sprintf("http://localhost:9998/v1/api/ingest/%s", uc.Content), nil)
		http.DefaultClient.Do(req)
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
	s.GetGenres(&ms)
	s.GetKeywords(&ms)
	s.GetProductionCompanies(&ms)
	s.GetSpokenLanguages(&ms)
	services.NewGoodContentRequest(ctx, ms)
}

// TvById gets tv by id
func (s *SearchService) TvById(ctx *gin.Context) {
	var uc services.UriContent[uint64]
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

	s.GetGenres(&ms)
	s.GetKeywords(&ms)
	s.GetProductionCompanies(&ms)
	s.GetSpokenLanguages(&ms)
	services.NewGoodContentRequest(ctx, ms)
}

// BookById gets book by id
func (s *SearchService) BookById(ctx *gin.Context) {
	var uc services.UriContent[uint64]
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
	var uc services.UriContent[string]
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
