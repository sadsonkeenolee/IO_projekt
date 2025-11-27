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
	NMovieFields           int = 12
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

var (
	MovieInsertQuery Query = Query{
		Content: "INSERT IGNORE INTO movies(budget,tmdb_id,title,overview,popularity,release_date,revenue,runtime,status,tagline,vote_average,vote_total) VALUES",
		Fields:  "(?,?,?,?,?,?,?,?,?,?,?,?)",
	}
	LanguagesInsertQuery Query = Query{
		Content: "INSERT IGNORE INTO languages(encoding, name) VALUES",
		Fields:  "(?, ?)",
	}
	KeywordsInsertQuery Query = Query{
		Content: "INSERT IGNORE INTO keywords(ID, name) VALUES",
		Fields:  "(?, ?)",
	}
	GenreInsertQuery Query = Query{
		Content: "INSERT IGNORE INTO genres(ID, name) VALUES",
		Fields:  "(?, ?)",
	}
	CountryInsertQuery Query = Query{
		Content: "INSERT IGNORE INTO countries(encoding, name) VALUES",
		Fields:  "(?, ?)",
	}
	CompananyInsertQuery Query = Query{
		Content: "INSERT IGNORE INTO companies(ID, name) VALUES",
		Fields:  "(?, ?)",
	}
	Movie2LanguagesInsertQuery = Query{
		Content: "INSERT IGNORE INTO movie2languages(movie_id, language_id) VALUES",
		Fields:  "(?, ?)",
	}
	Movie2KeywordsInsertQuery = Query{
		Content: "INSERT IGNORE INTO movie2keywords(movie_id, keyword_id) VALUES",
		Fields:  "(?, ?)",
	}
	Movie2GenresInsertQuery = Query{
		Content: "INSERT IGNORE INTO movie2genres(movie_id, genre_id) VALUES",
		Fields:  "(?, ?)",
	}
	Movie2CountriesInsertQuery = Query{
		Content: "INSERT IGNORE INTO movie2countries(movie_id, country_en) VALUES",
		Fields:  "(?, ?)",
	}

	Movie2CompaniesInsertQuery = Query{
		Content: "INSERT IGNORE INTO movie2companies(movie_id, company_id) VALUES",
		Fields:  "(?, ?)",
	}
)

type IdName struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

type CodeCountry struct {
	Encoding string `json:"iso_3166_1"`
	Name     string `json:"name"`
}

type CodeLanguage struct {
	Encoding string `json:"iso_639_1"`
	Name     string `json:"name"`
}

type Genre = IdName
type Keywords = IdName
type ProductionCompanies = IdName

type CastCrewMetadata struct {
	MovieId uint64
	Cast    []CastMember
	Crew    []CrewMember
}

func (ccm *CastCrewMetadata) IsInsertable() bool {
	// assume the struct is insertable, while batching inserts, let it fail on the
	// inserting level
	return true
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
	Budget              uint64
	Genres              []Genre
	MovieId             uint64
	Keywords            []Keywords
	OriginalLanguage    string
	Title               string
	Overview            string
	Popularity          float64
	ProductionCompanies []ProductionCompanies
	ProductionCountries []CodeCountry
	ReleaseDate         time.Time
	Revenue             int64
	Runtime             int64
	SpokenLanguages     []CodeLanguage
	Status              string
	Tagline             string
	AverageScore        float64
	TotalScore          uint64
	Cast                []CastMember
	Crew                []CrewMember
}

var DeadlockError = fmt.Errorf("Error 1213 (40001): Deadlock found when trying to get lock; try restarting transaction")

// placeholder
func SqlErrorCode(err error) uint {
	if err == DeadlockError {
		return 1213
	}
	return 0
}

func (mi *MovieInsertable) IsInsertable() bool {
	return true
}

type MovieQueryable struct {
	Id uint64
	MovieInsertable
}

func (mq *MovieQueryable) Queryable() bool {
	return true
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
	_, _ = tx.Exec("set autocommit=0")
	_, _ = tx.Exec("set unique_checks=0")
	_, _ = tx.Exec("set foreign_key_checks=0")
	_, _ = tx.Exec(*stmt, *argFields...)
	_, _ = tx.Exec("set foreign_key_checks=1")
	_, _ = tx.Exec("set unique_checks=1")
	_, _ = tx.Exec("commit")
	err = tx.Commit()
	return err
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
						queryFields = append(queryFields, ip.Fields)
						argFields = append(argFields, mi.Budget)
						argFields = append(argFields, mi.MovieId)
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
					stmt := fmt.Sprintf("%v %v", ip.Content, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						fmt.Println(err)
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
							queryFields = append(queryFields, ip.Fields)
							argFields = append(argFields, sl.Encoding)
							argFields = append(argFields, sl.Name)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v %v", ip.Content, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						fmt.Println(err)
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
							queryFields = append(queryFields, ip.Fields)
							argFields = append(argFields, k.Id)
							argFields = append(argFields, k.Name)
						}
					}

					<-c
					stmt := fmt.Sprintf("%v %v", ip.Content, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						fmt.Println(err)
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
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NGenreFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}

						for _, g := range mi.Genres {
							encodingTracker.mu.Lock()
							if _, ok := encodingTracker.genresSeen[g.Id]; ok {
								encodingTracker.mu.Unlock()
								continue
							}
							encodingTracker.genresSeen[g.Id] = true
							encodingTracker.mu.Unlock()
							queryFields = append(queryFields, ip.Fields)
							argFields = append(argFields, g.Id)
							argFields = append(argFields, g.Name)
						}
					}

					<-c
					stmt := fmt.Sprintf("%v %v", ip.Content, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						fmt.Println(err)
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
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NCountryFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}

						for _, pc := range mi.ProductionCountries {
							encodingTracker.mu.Lock()
							if _, ok := encodingTracker.countriesSeen[pc.Name]; ok {
								encodingTracker.mu.Unlock()
								continue
							}
							encodingTracker.countriesSeen[pc.Encoding] = true
							encodingTracker.mu.Unlock()
							queryFields = append(queryFields, ip.Fields)
							argFields = append(argFields, pc.Encoding)
							argFields = append(argFields, pc.Name)
						}
					}

					<-c
					stmt := fmt.Sprintf("%v %v", ip.Content, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						fmt.Println(err)
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
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NCompanyFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}

						for _, pc := range mi.ProductionCompanies {
							encodingTracker.mu.Lock()
							if _, ok := encodingTracker.companiesSeen[pc.Id]; ok {
								encodingTracker.mu.Unlock()
								continue
							}
							encodingTracker.companiesSeen[pc.Id] = true
							encodingTracker.mu.Unlock()
							queryFields = append(queryFields, ip.Fields)
							argFields = append(argFields, pc.Id)
							argFields = append(argFields, pc.Name)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v %v", ip.Content, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						fmt.Println(err)
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
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NMovie2LanguagesFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						for _, sl := range mi.SpokenLanguages {
							queryFields = append(queryFields, ip.Fields)
							argFields = append(argFields, mi.MovieId)
							argFields = append(argFields, sl.Encoding)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v %v;", ip.Content, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						fmt.Println(err)
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
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NMovie2KeywordsFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						for _, k := range mi.Keywords {
							queryFields = append(queryFields, ip.Fields)
							argFields = append(argFields, mi.MovieId)
							argFields = append(argFields, k.Id)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v %v;", ip.Content, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						fmt.Println(err)
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
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NMovie2GenresFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						for _, g := range mi.Genres {
							queryFields = append(queryFields, ip.Fields)
							argFields = append(argFields, mi.MovieId)
							argFields = append(argFields, g.Id)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v %v;", ip.Content, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						fmt.Println(err)
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
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NMovie2CountriesFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						for _, pc := range mi.ProductionCountries {
							queryFields = append(queryFields, ip.Fields)
							argFields = append(argFields, mi.MovieId)
							argFields = append(argFields, pc.Encoding)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v %v;", ip.Content, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						fmt.Println(err)
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
					var queryFields []string = make([]string, 0, *chunkSize)
					var argFields []any = make([]any, 0, NMovie2CompaniesFields*(*chunkSize))
					for _, item := range chunk {
						mi, ok := (*item).(*MovieInsertable)
						if !ok {
							continue
						}
						for _, pc := range mi.ProductionCompanies {
							queryFields = append(queryFields, ip.Fields)
							argFields = append(argFields, mi.MovieId)
							argFields = append(argFields, pc.Id)
						}
					}
					<-c
					stmt := fmt.Sprintf("%v %v;", ip.Content, strings.Join(queryFields, ","))
					err := InsertStmt(db, &stmt, &argFields)
					if err != nil {
						fmt.Println(err)
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

	target.Budget, _ = strconv.ParseUint(s[0], 10, 64)
	_ = json.Unmarshal([]byte(s[1]), &target.Genres)
	target.MovieId, _ = strconv.ParseUint(s[3], 10, 64)
	_ = json.Unmarshal([]byte(s[4]), &target.Keywords)
	target.OriginalLanguage = s[5]
	target.Title = s[6]
	target.Overview = s[7]
	target.Popularity, _ = strconv.ParseFloat(s[8], 64)
	_ = json.Unmarshal([]byte(s[9]), &target.ProductionCompanies)
	_ = json.Unmarshal([]byte(s[10]), &target.ProductionCountries)
	target.ReleaseDate, _ = time.Parse("2006-01-02", s[11])
	target.Revenue, _ = strconv.ParseInt(s[12], 10, 64)
	target.Runtime, _ = strconv.ParseInt(s[13], 10, 64)
	_ = json.Unmarshal([]byte(s[14]), &target.SpokenLanguages)
	target.Status = s[15]
	target.Tagline = s[17]
	target.AverageScore, _ = strconv.ParseFloat(s[18], 64)
	target.TotalScore, _ = strconv.ParseUint(s[19], 10, 64)
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

	tgt.MovieId, _ = strconv.ParseUint(s[0], 10, 64)
	json.Unmarshal([]byte(s[2]), &tgt.Cast)
	json.Unmarshal([]byte(s[3]), &tgt.Crew)

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
	if dfm.ReadAt == nil {
		ni := ItemNotFound
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
