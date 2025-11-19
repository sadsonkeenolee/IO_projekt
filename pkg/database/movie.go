package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

const (
	TmdbDataLength        int = 20
	TmdbCreditsDataLength int = 4
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

func (ccm *CastCrewMetadata) Insertable() {}

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

func (mi *MovieInsertable) Insertable() {}

type MovieQueryable struct {
	Id uint64
	MovieInsertable
}

func (mq *MovieQueryable) Queryable() {}

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

func InsertIntoMovies(db *sql.DB) func(data *Insertable) error {
	movieIdsSeen := map[uint64]bool{}
	return func(data *Insertable) error {
		mme, err := CastFromInsertableToMovieInsertable(data)
		if err != nil {
			return err
		}

		if _, ok := movieIdsSeen[mme.MovieId]; ok {
			return fmt.Errorf("element already added")
		}

		_, err = db.Exec("INSERT INTO movies(budget,tmdb_id,title,overview,popularity,release_date,revenue,runtime,status,tagline,vote_average,vote_total) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)",
			mme.Budget, mme.MovieId, mme.Title, mme.Overview, mme.Popularity,
			mme.ReleaseDate, mme.Revenue, mme.Runtime, mme.Status, mme.Tagline,
			mme.AverageScore, mme.TotalScore)

		if err != nil {
			return err
		}

		movieIdsSeen[mme.MovieId] = true
		return nil
	}
}

func InsertIntoLanguages(db *sql.DB) func(data *Insertable) error {
	encodingsSeen := map[string]bool{}
	return func(data *Insertable) error {
		mme, err := CastFromInsertableToMovieInsertable(data)

		if err != nil {
			return fmt.Errorf("invalid interface")
		}

		for _, sl := range mme.SpokenLanguages {
			if _, ok := encodingsSeen[sl.Name]; ok {
				continue
			}

			_, err := db.Exec("INSERT INTO languages(encoding, name) VALUES (?, ?)", sl.Encoding, sl.Name)
			if err != nil {
				return err
			}
			encodingsSeen[sl.Name] = true
		}
		return nil
	}
}

func InsertIntoKeywords(db *sql.DB) func(data *Insertable) error {
	keywordsSeen := map[uint64]bool{}
	return func(data *Insertable) error {
		mme, err := CastFromInsertableToMovieInsertable(data)
		if err != nil {
			return fmt.Errorf("invalid interface")
		}

		for _, k := range mme.Keywords {
			if _, ok := keywordsSeen[k.Id]; ok {
				continue
			}
			_, err := db.Exec("INSERT INTO keywords(ID, name) VALUES (?, ?)", k.Id, k.Name)
			if err != nil {
				return err
			}
			keywordsSeen[k.Id] = true
		}
		return nil
	}
}

func InsertIntoGenres(db *sql.DB) func(data *Insertable) error {
	genresSeen := map[uint64]bool{}
	return func(data *Insertable) error {
		mme, err := CastFromInsertableToMovieInsertable(data)
		if err != nil {
			return fmt.Errorf("invalid interface")
		}

		for _, g := range mme.Genres {
			if _, ok := genresSeen[g.Id]; ok {
				continue
			}
			_, err := db.Exec("INSERT INTO genres(ID, name) VALUES (?, ?)", g.Id, g.Name)
			if err != nil {
				return err
			}
			genresSeen[g.Id] = true
		}
		return nil
	}
}

func InsertIntoCountries(db *sql.DB) func(data *Insertable) error {
	countrySeen := map[string]bool{}
	return func(data *Insertable) error {
		mme, err := CastFromInsertableToMovieInsertable(data)
		if err != nil {
			return fmt.Errorf("invalid interface")
		}
		for _, c := range mme.ProductionCountries {
			if _, ok := countrySeen[c.Name]; ok {
				continue
			}
			_, err := db.Exec("INSERT INTO countries(encoding, name) VALUES (?, ?)", c.Encoding, c.Name)
			if err != nil {
				return err
			}
			countrySeen[c.Name] = true
		}
		return nil
	}
}

func InsertIntoCompanies(db *sql.DB) func(data *Insertable) error {
	companiesSeen := map[uint64]bool{}
	return func(data *Insertable) error {
		mme, err := CastFromInsertableToMovieInsertable(data)
		if err != nil {
			return fmt.Errorf("invalid interface")
		}

		for _, c := range mme.ProductionCompanies {
			if _, ok := companiesSeen[c.Id]; ok {
				continue
			}
			_, err := db.Exec("INSERT INTO companies(ID, name) VALUES (?, ?)", c.Id, c.Name)
			if err != nil {
				return err
			}
			companiesSeen[c.Id] = true
		}
		return nil
	}
}

func InsertIntoMovie2Languages(db *sql.DB) func(data *Insertable) error {
	return func(data *Insertable) error {
		mme, err := CastFromInsertableToMovieInsertable(data)
		if err != nil {
			return fmt.Errorf("invalid interface")
		}
		for _, sl := range mme.SpokenLanguages {
			_, err := db.Exec("INSERT INTO movie2languages(movie_id, language_id) VALUES (?, ?)", mme.MovieId, sl.Encoding)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func InsertIntoMovie2Keywords(db *sql.DB) func(data *Insertable) error {
	return func(data *Insertable) error {
		mme, err := CastFromInsertableToMovieInsertable(data)
		if err != nil {
			return fmt.Errorf("invalid interface")
		}

		for _, k := range mme.Keywords {
			_, err := db.Exec("INSERT INTO movie2keywords(movie_id, keyword_id) VALUES (?, ?)", mme.MovieId, k.Id)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func InsertIntoMovie2Genres(db *sql.DB) func(data *Insertable) error {
	return func(data *Insertable) error {
		mme, err := CastFromInsertableToMovieInsertable(data)
		if err != nil {
			return fmt.Errorf("invalid interface")
		}

		for _, g := range mme.Genres {
			_, err := db.Exec("INSERT INTO movie2genres(movie_id, genre_id) VALUES (?, ?)", mme.MovieId, g.Id)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func InsertIntoMovie2Countries(db *sql.DB) func(data *Insertable) error {
	return func(data *Insertable) error {
		mme, err := CastFromInsertableToMovieInsertable(data)
		if err != nil {
			return fmt.Errorf("invalid interface")
		}

		for _, pc := range mme.ProductionCountries {
			_, err := db.Exec("INSERT INTO movie2countries(movie_id, country_en) VALUES (?, ?)", mme.MovieId, pc.Encoding)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func InsertIntoMovie2Companies(db *sql.DB) func(data *Insertable) error {
	return func(data *Insertable) error {
		mme, err := CastFromInsertableToMovieInsertable(data)
		if err != nil {
			return fmt.Errorf("invalid interface")
		}
		for _, pc := range mme.ProductionCompanies {
			_, err := db.Exec("INSERT INTO movie2companies(movie_id, company_id) VALUES (?, ?)", mme.MovieId, pc.Id)
			if err != nil {
				return err
			}
		}
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
