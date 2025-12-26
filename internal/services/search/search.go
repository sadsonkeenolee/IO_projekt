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

	rows, err := s.DB.Query(`SELECT ID, budget, tmdb_id, original_lang_id,
		title, overview, popularity, release_date, revenue, runtime, status, 
		tagline, vote_average, vote_total FROM movies WHERE title LIKE ?`, "%"+bind.Content+"%")
	if err != nil {
		s.Logger.Printf("couldn't find a movie with title=%v. Reason: %v\n", bind.Content, err)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}
	defer rows.Close()
	mqs := []database.MovieSelectable{}
	for rows.Next() {
		var mq database.MovieSelectable
		if err := rows.Scan(
			&mq.Id, &mq.Budget, &mq.MovieId, &mq.OriginalLanguage,
			&mq.Title, &mq.Overview, &mq.Popularity, &mq.ReleaseDate,
			&mq.Revenue, &mq.Runtime, &mq.Status, &mq.Tagline, &mq.AverageScore,
			&mq.TotalScore); err != nil {
			s.Logger.Println(err)
			continue
		}
		mqs = append(mqs, mq)
	}
	services.NewGoodContentRequest(ctx, mqs)
}

func (s *SearchService) TvById(ctx *gin.Context) {
	bind := Id{}
	if err := ctx.ShouldBindUri(&bind); err != nil {
		s.Logger.Println(err)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	var mq database.MovieSelectable
	rows, err := s.DB.Query(mq.ConstructSelectQuery(), bind.Content)
	if err != nil {
		s.Logger.Printf("couldn't find a movie with ID=%v.\n", bind.Content)
		s.Logger.Println(err)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(
			&mq.Id, &mq.Budget, &mq.MovieId, &mq.OriginalLanguage,
			&mq.Title, &mq.Overview, &mq.Popularity, &mq.ReleaseDate,
			&mq.Revenue, &mq.Runtime, &mq.Status, &mq.Tagline, &mq.AverageScore,
			&mq.TotalScore); err != nil {
			s.Logger.Println(err)
			services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
			return
		}
	}

	var companies []database.Movie2CompanySelectable
	var c database.Movie2CompanySelectable
	rows, err = s.DB.Query(c.ConstructSelectQuery(), bind.Content)
	if err != nil {
		s.Logger.Printf("couldn't fetch companies. Reason: %v\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		var company database.Movie2CompanySelectable
		if err := rows.Scan(&company.MovieId, &company.Company); err != nil {
			s.Logger.Println(err)
		} else {
			companies = append(companies, company)
		}
	}
	s.Logger.Println(companies)

	if mq.MovieId != bind.Content {
		s.Logger.Printf("id (%v) mismatches the found id (%v)\n",
			bind.Content, mq.MovieId)
		services.NewBadContentRequest(ctx, "movie doesn't exist")
		return
	}
	services.NewGoodContentRequest(ctx, mq)
}

func (s *SearchService) BookById(ctx *gin.Context) {
	bind := Id{}
	if err := ctx.ShouldBindUri(&bind); err != nil {
		s.Logger.Println(err)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	rows, err := s.DB.Query(`SELECT * FROM books WHERE ID=?`, bind.Content)
	if err != nil {
		s.Logger.Printf("couldn't find a movie with title=%v\n", bind.Content)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	defer rows.Close()
	bqs := []database.BookSelectable{}
	for rows.Next() {
		var bq database.BookSelectable
		if err := rows.Scan(
			&bq.Id, &bq.Title, &bq.Rating, &bq.Isbn, &bq.Isbn13, &bq.Language,
			&bq.Pages, &bq.TotalRating, &bq.ReleaseDate, &bq.Publisher,
		); err != nil {
			s.Logger.Println(err)
			continue
		}
		bqs = append(bqs, bq)
	}
	services.NewGoodContentRequest(ctx, bqs)
}

func (s *SearchService) BookByTitle(ctx *gin.Context) {
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

	rows, err := s.DB.Query(`SELECT * FROM books WHERE title LIKE ?`, "%"+bind.Content+"%")
	if err != nil {
		s.Logger.Printf("couldn't find a movie with title=%v\n", bind.Content)
		services.NewBadContentRequest(ctx, services.InvalidRequestMessage)
		return
	}

	defer rows.Close()
	bqs := []database.BookSelectable{}
	for rows.Next() {
		var bq database.BookSelectable
		if err := rows.Scan(
			&bq.Id, &bq.Title, &bq.Rating, &bq.Isbn, &bq.Isbn13, &bq.Language,
			&bq.Pages, &bq.TotalRating, &bq.ReleaseDate, &bq.Publisher,
		); err != nil {
			s.Logger.Println(err)
			continue
		}
		bqs = append(bqs, bq)
	}
	services.NewGoodContentRequest(ctx, bqs)
}

// ExposeConnection exposes configuration.
func (s *SearchService) ExposeConnection() *services.Connection {
	return s.ConnInfo
}

func (s *SearchService) String() string {
	return "Fetch"
}
