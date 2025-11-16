package database

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type DatasetFileMetadata struct {
	Directory string
	Type      string
	IsRead    bool
	CreatedAt *string
	ReadAt    *string
}

const (
	NoInformation         string = "No information"
	TmdbDataLength        int    = 20
	TmdbCreditsDataLength int    = 4
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

type MovieMetadataExtracted struct {
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

type CastCrewMetadata struct {
	MovieId uint64
	Cast    []CastMember
	Crew    []CrewMember
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

func TmdbMapFromStream(stream *[]string, data *any) {
	if len(*stream) != TmdbDataLength {
		fmt.Printf("invalid length: %v, expected: %v\n", len(*stream), TmdbDataLength)
		return
	}

	s := *stream
	*data = MovieMetadataExtracted{}
	target, ok := (*data).(MovieMetadataExtracted)
	if !ok {
		fmt.Println("invalid interface")
		return
	}

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
}

func TmdbMapCreditsFromStream(stream *[]string, data *any) {
	if len(*stream) != TmdbCreditsDataLength {
		fmt.Printf("invalid length: %v, expected: %v\n", len(*stream), TmdbCreditsDataLength)
		return
	}

	s := *stream
	*data = CastCrewMetadata{}
	target, ok := (*data).(CastCrewMetadata)
	if !ok {
		fmt.Println("castcrew invalid interface")
		return
	}

	target.MovieId, _ = strconv.ParseUint(s[0], 10, 64)
	json.Unmarshal([]byte(s[2]), &target.Cast)
	json.Unmarshal([]byte(s[3]), &target.Crew)

	*data = target
}

func TmdbJoinBoth(tmdb, tmdbCredits, transformed *[]*any) {
	mappedIds := map[uint64]*CastCrewMetadata{}

	for _, ccm := range *tmdbCredits {
		target, ok := (*ccm).(CastCrewMetadata)
		if !ok {
			fmt.Println("invalid interface")
			continue
		}
		mappedIds[target.MovieId] = &target
	}

	for _, mme := range *tmdb {
		source, ok := (*mme).(MovieMetadataExtracted)
		if !ok {
			fmt.Println("invalid interface")
			continue
		}
		source.Cast = mappedIds[source.MovieId].Cast
		source.Crew = mappedIds[source.MovieId].Crew
		*mme = source
		*transformed = append(*transformed, mme)
	}
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
