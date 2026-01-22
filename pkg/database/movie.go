package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"
)

const (
	TmdbDataLength         int = 20
	TmdbCreditsDataLength  int = 4
	NMovieFields           int = 13
	NLanguageFields        int = 2
	NKeywordFields         int = 2
	NGenreFields           int = 2
	NCountryFields         int = 2
	NCompanyFields         int = 2
	NMovie2CompaniesFields int = 2
	NMovie2LanguagesFields int = 2
	NMovie2KeywordsFields  int = 2
	NMovie2GenresFields    int = 2
	NMovie2CountriesFields int = 2
)

var StreamTmdbIndices map[string]int = map[string]int{
	"Budget":              0,
	"Genre":               1,
	"MovieId":             3,
	"Keywords":            4,
	"OriginalLanguage":    5,
	"Title":               6,
	"Overview":            7,
	"Popularity":          8,
	"ProductionCompanies": 9,
	"ProductionCountries": 10,
	"ReleaseDate":         11,
	"Revenue":             12,
	"Runtime":             13,
	"SpokenLanguages":     14,
	"Status":              15,
	"Tagline":             17,
	"AverageScore":        18,
	"TotalScore":          19,
}

var TmdbCreditsIndices map[string]int = map[string]int{
	"MovieId": 0,
	"Cast":    2,
	"Crew":    3,
}

type IdName struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

type CodeCountry struct {
	Encoding string `json:"iso_3166_1"`
	Name     string `json:"name"`
}

func (cc *CodeCountry) IsInsertable() (*Table, bool) {
	return NewTable(
		"countries",
		[]string{"encoding", "country"},
	), true
}

func (cc *CodeCountry) ConstructInsertQuery() string {
	t, ok := cc.IsInsertable()
	if !ok {
		return ""
	}
	return fmt.Sprintf("INSERT IGNORE INTO %v%v VALUES ", t.Name,
		JoinTableFields(t))
}

type LanguageEncoding struct {
	Encoding string `json:"iso_639_1"`
	Name     string `json:"name"`
}

func (le *LanguageEncoding) IsInsertable() (*Table, bool) {
	return NewTable(
		"languages",
		[]string{"encoding", "language"},
	), true
}

func (le *LanguageEncoding) ConstructInsertQuery() string {
	t, ok := le.IsInsertable()
	if !ok {
		return ""
	}
	return fmt.Sprintf("INSERT IGNORE INTO %v%v VALUES ", t.Name,
		JoinTableFields(t))
}

type Genre struct {
	IdName
}

func (g *Genre) IsInsertable() (*Table, bool) {
	return NewTable(
		"genres",
		[]string{"ID", "genre"},
	), true
}

func (g *Genre) ConstructInsertQuery() string {
	t, ok := g.IsInsertable()
	if !ok {
		return ""
	}
	return fmt.Sprintf("INSERT IGNORE INTO %v%v VALUES ", t.Name,
		JoinTableFields(t))
}

type Keywords struct {
	IdName
}

func (k *Keywords) IsInsertable() (*Table, bool) {
	return NewTable(
		"keywords",
		[]string{"ID", "keyword"},
	), true
}

func (k *Keywords) ConstructInsertQuery() string {
	t, ok := k.IsInsertable()
	if !ok {
		return ""
	}
	return fmt.Sprintf("INSERT IGNORE INTO %v%v VALUES ", t.Name,
		JoinTableFields(t))
}

type ProductionCompaniesInsertable struct {
	IdName
}

func (pc *ProductionCompaniesInsertable) IsInsertable() (*Table, bool) {
	return NewTable(
		"companies",
		[]string{"ID", "company"},
	), true
}

func (pc *ProductionCompaniesInsertable) ConstructInsertQuery() string {
	t, ok := pc.IsInsertable()
	if !ok {
		return ""
	}
	return fmt.Sprintf("INSERT IGNORE INTO %v%v VALUES ", t.Name,
		JoinTableFields(t))
}

type Movie2GenresInsertable struct {
	MovieId uint64
	GenreId uint64
}

func (m2g *Movie2GenresInsertable) IsInsertable() (*Table, bool) {
	return NewTable(
		"movie2genres",
		[]string{"movie_id", "genre_id"},
	), true
}

func (m2g *Movie2GenresInsertable) ConstructInsertQuery() string {
	t, ok := m2g.IsInsertable()
	if !ok {
		return ""
	}
	return fmt.Sprintf("INSERT IGNORE INTO %v%v VALUES ", t.Name,
		JoinTableFields(t))
}

type Movie2KeywordsInsertable struct {
	MovieId   uint64
	KeywordId uint64
}

func (m2k *Movie2KeywordsInsertable) IsInsertable() (*Table, bool) {
	return NewTable(
		"movie2keywords",
		[]string{"movie_id", "keyword_id"},
	), true
}

func (m2k *Movie2KeywordsInsertable) ConstructInsertQuery() string {
	t, ok := m2k.IsInsertable()
	if !ok {
		return ""
	}
	return fmt.Sprintf("INSERT IGNORE INTO %v%v VALUES ", t.Name,
		JoinTableFields(t))
}

type Movie2LanguagesInsertable struct {
	MovieId  uint64
	Language string
}

func (m2l *Movie2LanguagesInsertable) IsInsertable() (*Table, bool) {
	return NewTable(
		"movie2languages",
		[]string{"movie_id", "language_encoding"},
	), true
}

func (m2l *Movie2LanguagesInsertable) ConstructInsertQuery() string {
	t, ok := m2l.IsInsertable()
	if !ok {
		return ""
	}
	return fmt.Sprintf("INSERT IGNORE INTO %v%v VALUES ", t.Name,
		JoinTableFields(t))
}

type Movie2CountriesInsertable struct {
	MovieId uint64
	Country string
}

func (m2c *Movie2CountriesInsertable) IsInsertable() (*Table, bool) {
	return NewTable(
		"movie2countries",
		[]string{"movie_id", "country_encoding"},
	), true
}

func (m2c *Movie2CountriesInsertable) ConstructInsertQuery() string {
	t, ok := m2c.IsInsertable()
	if !ok {
		return ""
	}
	return fmt.Sprintf("INSERT IGNORE INTO %v%v VALUES ", t.Name,
		JoinTableFields(t))
}

type Movie2CompaniesInsertable struct {
	MovieId   uint64
	CompanyId string
}

func (m2c *Movie2CompaniesInsertable) IsInsertable() (*Table, bool) {
	return NewTable(
		"movie2companies",
		[]string{"movie_id", "company_id"},
	), true
}

func (m2c *Movie2CompaniesInsertable) ConstructInsertQuery() string {
	t, ok := m2c.IsInsertable()
	if !ok {
		return ""
	}
	return fmt.Sprintf("INSERT IGNORE INTO %v%v VALUES ", t.Name,
		JoinTableFields(t))
}

type CastCrewMetadata struct {
	MovieId uint64
	Cast    []CastMember
	Crew    []CrewMember
}

func (ccm *CastCrewMetadata) IsInsertable() (*Table, bool) {
	// assume the struct is insertable, while batching inserts, let it fail on the
	// inserting level
	return nil, false
}

func (ccm *CastCrewMetadata) ConstructInsertQuery() string {
	return ""
}

type CastMember struct {
	CastId    uint64 `json:"cast_id"`
	Character string `json:"character"`
	CreditId  string `json:"credit_id"`
	Gender    uint64 `json:"gender"`
	Id        uint64 `json:"id"`
	Name      string `json:"name"`
	Order     uint64 `json:"order"`
}

type CrewMember struct {
	CreditId   string `json:"credit_id"`
	Department string `json:"department"`
	Gender     uint64 `json:"gender"`
	Id         uint64 `json:"id"`
	Job        string `json:"job"`
	Name       string `json:"name"`
}

type MovieInsertable struct {
	Budget              uint64                          `json:"budget"`
	Genres              []Genre                         `json:"genres"`
	MovieId             uint64                          `json:"movie_id"`
	Keywords            []Keywords                      `json:"keywords"`
	OriginalLanguage    string                          `json:"original_language"`
	Title               string                          `json:"title"`
	Overview            string                          `json:"overview"`
	Popularity          float64                         `json:"popularity"`
	ProductionCompanies []ProductionCompaniesInsertable `json:"production_companies"`
	ProductionCountries []CodeCountry                   `json:"production_countries"`
	ReleaseDate         time.Time                       `json:"release_date"`
	Revenue             int64                           `json:"revenue"`
	Runtime             int64                           `json:"runtime"`
	SpokenLanguages     []LanguageEncoding              `json:"spoken_languages"`
	Status              string                          `json:"status"`
	Tagline             string                          `json:"tagline"`
	AverageScore        float64                         `json:"rating"`
	TotalScore          uint64                          `json:"total_ratings"`
	Cast                []CastMember                    `json:"cast"`
	Crew                []CrewMember                    `json:"crew"`
}

func (mi *MovieInsertable) IsInsertable() (*Table, bool) {
	return NewTable(
		"movies",
		[]string{"budget", "tmdb_id", "language", "title", "overview",
			"popularity", "release_date", "revenue", "runtime", "status", "tagline",
			"rating", "total_ratings"},
	), true
}

func (mi *MovieInsertable) ConstructInsertQuery() string {
	t, ok := mi.IsInsertable()
	if !ok {
		return ""
	}
	return fmt.Sprintf("INSERT IGNORE INTO %v%v VALUES ", t.Name,
		JoinTableFields(t))
}

type MovieSelectable struct {
	Id uint64 `json:"id"`
	MovieInsertable
}

func (mq *MovieSelectable) IsSelectable() (*Table, bool) {
	return NewTable(
		"movies",
		[]string{"ID", "budget", "tmdb_id", "language", "title", "overview",
			"popularity", "release_date", "revenue", "runtime", "status", "tagline",
			"rating", "total_ratings"},
	), true
}

func (mq *MovieSelectable) ConstructSelectQuery() string {
	_, ok := mq.IsSelectable()
	if !ok {
		return ""
	}
	return "call get_movie_by_id(?)"
}

type Movie2CompanySelectable struct {
	MovieId uint64
	Company string
}

func (m2c *Movie2CompanySelectable) IsSelectable() (*Table, bool) {
	return NewTable(
		"JoinedTable",
		[]string{"movie_id", "company"},
	), true
}

func (m2c *Movie2CompanySelectable) ConstructSelectQuery() string {
	_, ok := m2c.IsSelectable()
	if !ok {
		return ""
	}
	return "call get_production_companies(?)"
}

func CastFromAnyToInsertable(item *any) (Insertable, error) {
	mme, ok := (*item).(*MovieInsertable)
	if !ok {
		return nil, fmt.Errorf("cast from any to insertable failed")
	}
	return mme, nil
}

func CastFromMovieInsertableToInsertable(item *MovieInsertable) (Insertable, error) {
	return item, nil
}

func CastFromInsertableToMovieInsertable(item *Insertable) (*MovieInsertable, error) {
	mme, ok := (*item).(*MovieInsertable)
	if !ok {
		return nil, fmt.Errorf("cast from insertable to movie insertable failed")
	}
	return mme, nil
}

func InsertStmt(db *sql.DB, stmt *string, argFields *[]any) error {
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}
	if _, err = tx.Exec("set autocommit=0"); err != nil {
		return err
	}
	if _, err = tx.Exec("set unique_checks=0"); err != nil {
		return err
	}
	if _, err = tx.Exec("set foreign_key_checks=0"); err != nil {
		return err
	}
	if _, err = tx.Exec(*stmt, *argFields...); err != nil {
		return err
	}
	if _, err = tx.Exec("set foreign_key_checks=1"); err != nil {
		return err
	}
	if _, err = tx.Exec("set unique_checks=0"); err != nil {
		return err
	}
	if _, err = tx.Exec("commit"); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return nil
	}

	return nil
}

func InsertIntoMoviesChunked(db *sql.DB, chunkSize *int) func(data *Insertable) error {
	idTracker := struct {
		mu      sync.Mutex
		idsSeen map[uint64]bool
	}{idsSeen: make(map[uint64]bool)}

	return func(i *Insertable) error {
		ip, ok := (*i).(*InsertPipeline)
		if !ok {
			return fmt.Errorf("invalid interface (not a InsertPipeline)")
		}
		// prevent deadlocks
		c := make(chan bool, 1)
		defer close(c)
		c <- true

		var wg sync.WaitGroup
		for chunk := range slices.Chunk(*ip.Data, *chunkSize) {
			wg.Go(
				func() {
					i := chunk[0]
					t, _ := (*i).IsInsertable()
					query := (*i).ConstructInsertQuery()
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NMovieFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						idTracker.mu.Lock()
						if _, ok := idTracker.idsSeen[mi.MovieId]; ok {
							idTracker.mu.Unlock()
							continue
						}
						idTracker.idsSeen[mi.MovieId] = true
						idTracker.mu.Unlock()
						queryFields = append(queryFields, t.QueryField)
						argFields = append(argFields, mi.Budget)
						argFields = append(argFields, mi.MovieId)
						argFields = append(argFields, mi.OriginalLanguage)
						argFields = append(argFields, mi.Title)
						argFields = append(argFields, mi.Overview)
						argFields = append(argFields, mi.Popularity)
						argFields = append(argFields, mi.ReleaseDate)
						argFields = append(argFields, mi.Revenue)
						argFields = append(argFields, mi.Runtime)
						argFields = append(argFields, mi.Status)
						argFields = append(argFields, mi.Tagline)
						argFields = append(argFields, mi.AverageScore)
						argFields = append(argFields, mi.TotalScore)
					}

					// here put insert statements
					<-c
					stmt := fmt.Sprintf("%v%v", query, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						DatabaseLogger.Println(err)
					}
					c <- true
				})
		}
		wg.Wait()
		return nil
	}
}

func InsertIntoLanguagesChunked(db *sql.DB, chunkSize *int) func(data *Insertable) error {
	encodingTracker := struct {
		mu            sync.Mutex
		encodingsSeen map[string]bool
	}{encodingsSeen: make(map[string]bool)}
	return func(i *Insertable) error {
		ip, ok := (*i).(*InsertPipeline)
		if !ok {
			return fmt.Errorf("invalid interface (not a InsertPipeline)")
		}

		// prevent deadlocks
		c := make(chan bool, 1)
		defer close(c)
		c <- true
		var wg sync.WaitGroup
		for chunk := range slices.Chunk(*ip.Data, *chunkSize) {
			wg.Go(
				func() {
					template := LanguageEncoding{}
					t, _ := template.IsInsertable()
					query := template.ConstructInsertQuery()
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NLanguageFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						for _, sl := range mi.SpokenLanguages {
							encodingTracker.mu.Lock()
							if _, ok := encodingTracker.encodingsSeen[sl.Name]; ok {
								encodingTracker.mu.Unlock()
								continue
							}
							encodingTracker.encodingsSeen[sl.Encoding] = true
							encodingTracker.mu.Unlock()
							queryFields = append(queryFields, t.QueryField)
							argFields = append(argFields, sl.Encoding)
							argFields = append(argFields, sl.Name)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v%v", query, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						DatabaseLogger.Println(err)
					}
					c <- true
				})
		}
		wg.Wait()
		return nil
	}
}

func InsertIntoKeywordsChunked(db *sql.DB, chunkSize *int) func(data *Insertable) error {
	encodingTracker := struct {
		mu           sync.Mutex
		keywordsSeen map[uint64]bool
	}{keywordsSeen: make(map[uint64]bool)}
	return func(i *Insertable) error {
		ip, ok := (*i).(*InsertPipeline)
		if !ok {
			return fmt.Errorf("invalid interface (not a InsertPipeline)")
		}

		// prevent deadlocks
		c := make(chan bool, 1)
		defer close(c)
		c <- true
		var wg sync.WaitGroup
		for chunk := range slices.Chunk(*ip.Data, *chunkSize) {
			wg.Go(
				func() {
					template := Keywords{}
					t, _ := template.IsInsertable()
					query := template.ConstructInsertQuery()
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NKeywordFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}

						for _, k := range mi.Keywords {
							encodingTracker.mu.Lock()
							if _, ok := encodingTracker.keywordsSeen[k.Id]; ok {
								encodingTracker.mu.Unlock()
								continue
							}
							encodingTracker.keywordsSeen[k.Id] = true
							encodingTracker.mu.Unlock()
							queryFields = append(queryFields, t.QueryField)
							argFields = append(argFields, k.Id)
							argFields = append(argFields, k.Name)
						}
					}

					<-c
					stmt := fmt.Sprintf("%v%v", query, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						DatabaseLogger.Println(err)
					}
					c <- true
				})
		}
		wg.Wait()
		return nil
	}
}

func InsertIntoGenresChunked(db *sql.DB, chunkSize *int) func(data *Insertable) error {
	encodingTracker := struct {
		mu         sync.Mutex
		genresSeen map[uint64]bool
	}{genresSeen: make(map[uint64]bool)}
	return func(i *Insertable) error {
		ip, ok := (*i).(*InsertPipeline)
		if !ok {
			return fmt.Errorf("invalid interface (not a InsertPipeline)")
		}
		// prevent deadlocks
		c := make(chan bool, 1)
		defer close(c)
		c <- true
		var wg sync.WaitGroup

		for chunk := range slices.Chunk(*ip.Data, *chunkSize) {
			wg.Go(
				func() {
					template := Genre{}
					t, _ := template.IsInsertable()
					query := template.ConstructInsertQuery()
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NGenreFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}

						for _, g := range mi.Genres {
							encodingTracker.mu.Lock()
							ok := encodingTracker.genresSeen[g.Id]
							if ok || g.Name == "" {
								encodingTracker.mu.Unlock()
								continue
							}
							encodingTracker.genresSeen[g.Id] = true
							encodingTracker.mu.Unlock()
							queryFields = append(queryFields, t.QueryField)
							argFields = append(argFields, g.Id)
							argFields = append(argFields, g.Name)
						}
					}

					<-c
					stmt := fmt.Sprintf("%v%v", query, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						DatabaseLogger.Println(err)
					}
					c <- true

				})
		}
		wg.Wait()
		return nil
	}
}

func InsertIntoCountriesChunked(db *sql.DB, chunkSize *int) func(data *Insertable) error {
	encodingTracker := struct {
		mu            sync.Mutex
		countriesSeen map[string]bool
	}{countriesSeen: make(map[string]bool)}
	return func(i *Insertable) error {
		ip, ok := (*i).(*InsertPipeline)
		if !ok {
			return fmt.Errorf("invalid interface (not a InsertPipeline)")
		}
		// prevent deadlocks
		c := make(chan bool, 1)
		defer close(c)
		c <- true
		var wg sync.WaitGroup
		for chunk := range slices.Chunk(*ip.Data, *chunkSize) {
			wg.Go(
				func() {
					template := CodeCountry{}
					t, _ := template.IsInsertable()
					query := template.ConstructInsertQuery()
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NCountryFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						for _, pc := range mi.ProductionCountries {
							encodingTracker.mu.Lock()
							ok := encodingTracker.countriesSeen[pc.Name]
							if ok || pc.Name == "" {
								encodingTracker.mu.Unlock()
								continue
							}
							encodingTracker.countriesSeen[pc.Encoding] = true
							encodingTracker.mu.Unlock()
							queryFields = append(queryFields, t.QueryField)
							argFields = append(argFields, pc.Encoding)
							argFields = append(argFields, pc.Name)
						}
					}

					<-c
					stmt := fmt.Sprintf("%v%v", query, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						DatabaseLogger.Println(err)
					}
					c <- true
				})
		}
		wg.Wait()
		return nil
	}
}

func InsertIntoCompaniesChunked(db *sql.DB, chunkSize *int) func(data *Insertable) error {
	encodingTracker := struct {
		mu            sync.Mutex
		companiesSeen map[uint64]bool
	}{companiesSeen: make(map[uint64]bool)}
	return func(i *Insertable) error {
		ip, ok := (*i).(*InsertPipeline)
		if !ok {
			return fmt.Errorf("invalid interface (not a InsertPipeline)")
		}

		// prevent deadlocks
		c := make(chan bool, 1)
		defer close(c)
		c <- true
		var wg sync.WaitGroup
		for chunk := range slices.Chunk(*ip.Data, *chunkSize) {
			wg.Go(
				func() {
					template := ProductionCompaniesInsertable{}
					t, _ := template.IsInsertable()
					query := template.ConstructInsertQuery()
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NCompanyFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						for _, pc := range mi.ProductionCompanies {
							encodingTracker.mu.Lock()
							ok := encodingTracker.companiesSeen[pc.Id]
							if ok || pc.Name == "" {
								encodingTracker.mu.Unlock()
								continue
							}
							encodingTracker.companiesSeen[pc.Id] = true
							encodingTracker.mu.Unlock()
							queryFields = append(queryFields, t.QueryField)
							argFields = append(argFields, pc.Id)
							argFields = append(argFields, pc.Name)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v%v", query, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						DatabaseLogger.Println(err)
					}
					c <- true

				})
		}
		wg.Wait()
		return nil
	}
}

func InsertIntoMovie2LanguagesChunked(db *sql.DB, chunkSize *int) func(data *Insertable) error {
	return func(i *Insertable) error {
		ip, ok := (*i).(*InsertPipeline)
		if !ok {
			return fmt.Errorf("invalid interface (not a InsertPipeline)")
		}
		// prevent deadlocks
		c := make(chan bool, 1)
		defer close(c)
		c <- true
		var wg sync.WaitGroup
		for chunk := range slices.Chunk(*ip.Data, *chunkSize) {
			wg.Go(
				func() {
					template := Movie2LanguagesInsertable{}
					t, _ := template.IsInsertable()
					query := template.ConstructInsertQuery()
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NMovie2LanguagesFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						for _, sl := range mi.SpokenLanguages {
							queryFields = append(queryFields, t.QueryField)
							argFields = append(argFields, mi.MovieId)
							argFields = append(argFields, sl.Encoding)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v%v", query, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						DatabaseLogger.Println(err)
					}
					c <- true
				})
		}
		wg.Wait()
		return nil
	}
}

func InsertIntoMovie2KeywordsChunked(db *sql.DB, chunkSize *int) func(data *Insertable) error {
	return func(i *Insertable) error {
		ip, ok := (*i).(*InsertPipeline)
		if !ok {
			return fmt.Errorf("invalid interface (not a InsertPipeline)")
		}
		// prevent deadlocks
		c := make(chan bool, 1)
		defer close(c)
		c <- true

		var wg sync.WaitGroup
		for chunk := range slices.Chunk(*ip.Data, *chunkSize) {
			wg.Go(
				func() {
					template := Movie2KeywordsInsertable{}
					t, _ := template.IsInsertable()
					query := template.ConstructInsertQuery()
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NMovie2KeywordsFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						for _, k := range mi.Keywords {
							queryFields = append(queryFields, t.QueryField)
							argFields = append(argFields, mi.MovieId)
							argFields = append(argFields, k.Id)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v%v", query, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						DatabaseLogger.Println(err)
					}
					c <- true
				})
		}
		wg.Wait()
		return nil
	}
}

func InsertIntoMovie2GenresChunked(db *sql.DB, chunkSize *int) func(data *Insertable) error {
	return func(i *Insertable) error {
		ip, ok := (*i).(*InsertPipeline)
		if !ok {
			return fmt.Errorf("invalid interface (not a InsertPipeline)")
		}
		// prevent deadlocks
		c := make(chan bool, 1)
		defer close(c)
		c <- true
		var wg sync.WaitGroup
		for chunk := range slices.Chunk(*ip.Data, *chunkSize) {
			wg.Go(
				func() {
					template := Movie2GenresInsertable{}
					t, _ := template.IsInsertable()
					query := template.ConstructInsertQuery()
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NMovie2GenresFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						for _, g := range mi.Genres {
							queryFields = append(queryFields, t.QueryField)
							argFields = append(argFields, mi.MovieId)
							argFields = append(argFields, g.Id)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v%v", query, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						DatabaseLogger.Println(err)
					}
					c <- true
				})
		}
		wg.Wait()
		return nil
	}
}

func InsertIntoMovie2CountriesChunked(db *sql.DB, chunkSize *int) func(data *Insertable) error {
	return func(i *Insertable) error {
		ip, ok := (*i).(*InsertPipeline)
		if !ok {
			return fmt.Errorf("invalid interface (not a InsertPipeline)")
		}

		// prevent deadlocks
		c := make(chan bool, 1)
		defer close(c)
		c <- true
		var wg sync.WaitGroup
		for chunk := range slices.Chunk(*ip.Data, *chunkSize) {
			wg.Go(
				func() {
					template := Movie2CountriesInsertable{}
					t, _ := template.IsInsertable()
					query := template.ConstructInsertQuery()
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NMovie2CountriesFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						for _, pc := range mi.ProductionCountries {
							queryFields = append(queryFields, t.QueryField)
							argFields = append(argFields, mi.MovieId)
							argFields = append(argFields, pc.Encoding)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v%v", query, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						DatabaseLogger.Println(err)
					}
					c <- true
				})
		}
		wg.Wait()
		return nil
	}
}

func InsertIntoMovie2CompaniesChunked(db *sql.DB, chunkSize *int) func(data *Insertable) error {
	return func(i *Insertable) error {
		ip, ok := (*i).(*InsertPipeline)
		if !ok {
			return fmt.Errorf("invalid interface (not a InsertPipeline)")
		}

		// prevent deadlocks
		c := make(chan bool, 1)
		defer close(c)
		c <- true
		var wg sync.WaitGroup
		for chunk := range slices.Chunk(*ip.Data, *chunkSize) {
			wg.Go(
				func() {
					template := Movie2CompaniesInsertable{}
					t, _ := template.IsInsertable()
					query := template.ConstructInsertQuery()
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NMovie2CompaniesFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						for _, pc := range mi.ProductionCompanies {
							queryFields = append(queryFields, t.QueryField)
							argFields = append(argFields, mi.MovieId)
							argFields = append(argFields, pc.Id)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v%v", query, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						DatabaseLogger.Println(err)
					}
					c <- true
				})
		}
		wg.Wait()
		return nil
	}
}

func UpdateLanguageForMovies(db *sql.DB, chunkSize *int) func(data *Insertable) error {
	return func(i *Insertable) error {
		ip, ok := (*i).(*InsertPipeline)
		if !ok {
			return fmt.Errorf("invalid interface (not a InsertPipeline)")
		}

		// prevent deadlocks
		c := make(chan bool, 1)
		defer close(c)
		c <- true
		var wg sync.WaitGroup
		for chunk := range slices.Chunk(*ip.Data, *chunkSize) {
			wg.Go(
				func() {
					template := Movie2CompaniesInsertable{}
					t, _ := template.IsInsertable()
					query := template.ConstructInsertQuery()
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NMovie2CompaniesFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						for _, pc := range mi.ProductionCompanies {
							queryFields = append(queryFields, t.QueryField)
							argFields = append(argFields, mi.MovieId)
							argFields = append(argFields, pc.Id)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v%v", query, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						DatabaseLogger.Println(err)
					}
					c <- true
				})
		}
		wg.Wait()
		return nil
	}
}

func TmdbMapFromStream(stream *[]string, data *Insertable) error {
	if len(*stream) != TmdbDataLength {
		return fmt.Errorf("invalid length (%v): expected (%v)\n",
			len(*stream), TmdbDataLength)
	}

	s := *stream
	target := &MovieInsertable{}

	if len(s[StreamTmdbIndices["Title"]]) <= 0 {
		return fmt.Errorf("%v is incorrect title", s[6])
	}

	if len(s[StreamTmdbIndices["OriginalLanguage"]]) != 2 {
		return fmt.Errorf("%v is incorrect language", s[5])
	}

	target.Budget, _ = strconv.ParseUint(s[StreamTmdbIndices["Budget"]], 10, 64)
	_ = json.Unmarshal([]byte(s[StreamTmdbIndices["Genre"]]), &target.Genres)
	target.MovieId, _ = strconv.ParseUint(s[StreamTmdbIndices["MovieId"]], 10, 64)
	_ = json.Unmarshal([]byte(s[StreamTmdbIndices["Keywords"]]), &target.Keywords)
	target.OriginalLanguage = s[StreamTmdbIndices["OriginalLanguage"]]
	target.Title = s[StreamTmdbIndices["Title"]]
	target.Overview = s[StreamTmdbIndices["Overview"]]
	target.Popularity, _ = strconv.ParseFloat(s[StreamTmdbIndices["Popularity"]], 64)
	_ = json.Unmarshal([]byte(s[StreamTmdbIndices["ProductionCompanies"]]), &target.ProductionCompanies)
	_ = json.Unmarshal([]byte(s[StreamTmdbIndices["ProductionCountries"]]), &target.ProductionCountries)
	target.ReleaseDate, _ = time.Parse("2006-01-02", s[StreamTmdbIndices["ReleaseDate"]])
	target.Revenue, _ = strconv.ParseInt(s[StreamTmdbIndices["Revenue"]], 10, 64)
	target.Runtime, _ = strconv.ParseInt(s[StreamTmdbIndices["Runtime"]], 10, 64)
	_ = json.Unmarshal([]byte(s[StreamTmdbIndices["SpokenLanguages"]]), &target.SpokenLanguages)
	target.Status = s[StreamTmdbIndices["Status"]]
	target.Tagline = s[StreamTmdbIndices["Tagline"]]
	target.AverageScore, _ = strconv.ParseFloat(s[StreamTmdbIndices["AverageScore"]], 64)
	target.TotalScore, _ = strconv.ParseUint(s[StreamTmdbIndices["TotalScore"]], 10, 64)
	*data = target
	return nil
}

func TmdbMapCreditsFromStream(stream *[]string, data *Insertable) error {
	if len(*stream) != TmdbCreditsDataLength {
		return fmt.Errorf("invalid length (%v): expected (%v)\n",
			len(*stream), TmdbCreditsDataLength)
	}

	tgt := &CastCrewMetadata{}
	s := *stream

	tgt.MovieId, _ = strconv.ParseUint(s[TmdbCreditsIndices["MovieId"]], 10, 64)
	json.Unmarshal([]byte(s[TmdbCreditsIndices["Cast"]]), &tgt.Cast)
	json.Unmarshal([]byte(s[TmdbCreditsIndices["Crew"]]), &tgt.Crew)

	*data = tgt
	return nil
}

func TmdbJoinBoth(tmdb, tmdbCredits, transformed *[]*Insertable) error {
	mappedIds := map[uint64]*CastCrewMetadata{}

	for _, ccm := range *tmdbCredits {
		target, ok := (*ccm).(*CastCrewMetadata)
		if !ok {
			continue
		}
		mappedIds[target.MovieId] = target
	}

	for _, mme := range *tmdb {
		source, ok := (*mme).(*MovieInsertable)
		if !ok {
			continue
		}
		source.Cast = mappedIds[source.MovieId].Cast
		source.Crew = mappedIds[source.MovieId].Crew
		*mme = source
		*transformed = append(*transformed, mme)
	}
	return nil
}

func (dfm *DatasetFileMetadata) String() string {
	return fmt.Sprintf(`
		DatasetFileMetadata { 
			Directory: %v
			Type: %v
			CreatedAt: %v
			ReadAt: %v
		}`,
		dfm.Directory, dfm.Type, *dfm.CreatedAt, dfm.ReadAt)
}
