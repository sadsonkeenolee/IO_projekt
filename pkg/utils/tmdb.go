package utils

import (
	"time"

	"github.com/sadsonkeenolee/IO_projekt/pkg/database"
)

type TmdbTvSchema struct {
	Adult               bool                          `json:"adult"`
	BackdropPath        string                        `json:"backdrop_path"`
	Budget              uint64                        `json:"budget"`
	Genres              []TmdbGenreSchema             `json:"genres"`
	ImdbId              uint64                        `json:"id"`
	OriginalLanguage    string                        `json:"original_language"`
	OriginalTitle       string                        `json:"name"`
	Overview            string                        `json:"overview"`
	Popularity          float64                       `json:"popularity"`
	PosterPath          string                        `json:"poster_path"`
	ProductionCompanies []TmdbProductionCompanySchema `json:"production_companies"`
	ProductionCountries []TmdbProductionCountrySchema `json:"production_countries"`
	ReleaseDate         string                        `json:"first_air_date"`
	Revenue             int64                         `json:"revenue"`
	Runtime             int64                         `json:"runtime"`
	SpokenLanguages     []TmdbSpokenLanguageSchema    `json:"spoken_languages"`
	Status              string                        `json:"status"`
	Tagline             string                        `json:"tagline"`
	Video               bool                          `json:"video"`
	VoteAverage         float64                       `json:"vote_average"`
	VoteCount           uint64                        `json:"vote_count"`
}

func (m *TmdbTvSchema) IntoMovieInsertable() *database.MovieInsertable {
	parsedDate, _ := time.Parse("2006-01-02", m.ReleaseDate)
	return &database.MovieInsertable{
		Budget:              m.Budget,
		Genres:              []database.Genre{},
		MovieId:             m.ImdbId,
		OriginalLanguage:    m.OriginalLanguage,
		Title:               m.OriginalTitle,
		Overview:            m.Overview,
		Popularity:          m.Popularity,
		ReleaseDate:         parsedDate,
		Revenue:             m.Revenue,
		Runtime:             m.Runtime,
		Status:              m.Status,
		Tagline:             m.Tagline,
		AverageScore:        m.VoteAverage,
		TotalScore:          m.VoteCount,
		ProductionCompanies: mapProductionCompanies(m.ProductionCompanies),
		ProductionCountries: mapProductionCountries(m.ProductionCountries),
		SpokenLanguages:     mapLanguages(m.SpokenLanguages),
		Keywords:            []database.Keywords{},
		Cast:                []database.CastMember{},
		Crew:                []database.CrewMember{},
	}
}

func mapProductionCompanies(pcs []TmdbProductionCompanySchema) []database.ProductionCompaniesInsertable {
	out := make([]database.ProductionCompaniesInsertable, len(pcs))
	for i, v := range pcs {
		var pci database.ProductionCompaniesInsertable
		pci.Id = uint64(v.ID)
		pci.Name = v.Name
		out[i] = pci
	}
	return out
}

func mapProductionCountries(pcs []TmdbProductionCountrySchema) []database.CodeCountry {
	out := make([]database.CodeCountry, len(pcs))
	for i, v := range pcs {
		out[i] = database.CodeCountry{
			Encoding: v.Iso31661,
			Name:     v.Name,
		}
	}
	return out
}

func mapLanguages(sls []TmdbSpokenLanguageSchema) []database.LanguageEncoding {
	out := make([]database.LanguageEncoding, len(sls))
	for i, v := range sls {
		out[i] = database.LanguageEncoding{
			Encoding: v.Iso6391,
			Name:     v.Name,
		}
	}
	return out
}

type TmdbCollectionSchema struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
}

type TmdbGenreSchema struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TmdbProductionCompanySchema struct {
	ID            int     `json:"id"`
	LogoPath      *string `json:"logo_path"`
	Name          string  `json:"name"`
	OriginCountry string  `json:"origin_country"`
}

type TmdbProductionCountrySchema struct {
	Iso31661 string `json:"iso_3166_1"`
	Name     string `json:"name"`
}

type TmdbSpokenLanguageSchema struct {
	EnglishName string `json:"english_name"`
	Iso6391     string `json:"iso_639_1"`
	Name        string `json:"name"`
}

type TmdbSearchResponseSchema struct {
	Timestamp int64                     `json:"timestamp"`
	Content   TmdbResponseContentSchema `json:"content"`
}

type TmdbResponseContentSchema struct {
	Page         int                `json:"page"`
	Results      []TmdbResultSchema `json:"results"`
	TotalPages   int                `json:"total_pages"`
	TotalResults int                `json:"total_results"`
}

type TmdbResultSchema struct {
	Adult            bool     `json:"adult"`
	BackdropPath     *string  `json:"backdrop_path"`
	FirstAirDate     string   `json:"first_air_date"`
	GenreIds         []int    `json:"genre_ids"`
	Id               uint64   `json:"id"`
	Name             string   `json:"name"`
	OriginCountry    []string `json:"origin_country"`
	OriginalLanguage string   `json:"original_language"`
	OriginalName     string   `json:"original_name"`
	Overview         string   `json:"overview"`
	Popularity       float64  `json:"popularity"`
	PosterPath       *string  `json:"poster_path"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
}
