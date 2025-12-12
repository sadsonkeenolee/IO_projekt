package database

import (
	"database/sql"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BookStreamLength int = 12
	NBookFields      int = 10
	NAuthorsFields   int = 2
)

var (
	BookInsertQuery Query = Query{
		Content: `INSERT IGNORE INTO books(ID, title, rating, isbn, isbn13, language_code, pages, total_ratings, release_date, publisher) VALUES`,
		Fields:  `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
	}
	BookAuthorsQuery Query = Query{
		Content: `INSERT IGNORE INTO authors(book_id, author) VALUES`,
		Fields:  `(?, ?)`,
	}
)

var BookIndices map[string]int = map[string]int{
	"BookId":      0,
	"Title":       1,
	"Authors":     2,
	"Rating":      3,
	"Isbn":        4,
	"Isbn13":      5,
	"Language":    6,
	"Pages":       7,
	"TotalRating": 8,
	"ReleaseDate": 10,
	"Publisher":   11,
}

type BookInsertable struct {
	BookId      uint64    `json:"book_id"`
	Title       string    `json:"title"`
	Authors     []string  `json:"authors"`
	Rating      float64   `json:"score"`
	Isbn        string    `json:"isbn"`
	Isbn13      string    `json:"isbn13"`
	Language    string    `json:"language"`
	Pages       int64     `json:"pages"`
	TotalRating int64     `json:"total_rating"`
	ReleaseDate time.Time `json:"release_date"`
	Publisher   string    `json:"publisher"`
}

type BookQueryable struct {
	Id uint64 `json:"id"`
	BookInsertable
}

func (bq *BookQueryable) IsQueryable() bool {
	return true
}

func (bi *BookInsertable) IsInsertable() bool {
	return true
}

func CastFromBookInsertableToInsertable(item *BookInsertable) (Insertable, error) {
	return item, nil
}

func CastFromInsertableToBookInsertable(item *Insertable) (*BookInsertable, error) {
	bm, ok := (*item).(*BookInsertable)
	if !ok {
		return nil, fmt.Errorf("cast from insertable to movie insertable failed")
	}
	return bm, nil
}

func BookFromStream(stream *[]string, data *Insertable) error {
	if len(*stream) != BookStreamLength {
		return fmt.Errorf("invalid length (%v): expected (%v)\n",
			len(*stream), BookStreamLength)
	}

	s := *stream
	target := &BookInsertable{}
	target.BookId, _ = strconv.ParseUint(s[BookIndices["BookId"]], 10, 64)
	target.Title = s[BookIndices["Title"]]
	target.Authors = strings.Split(s[BookIndices["Authors"]], "/")
	target.Rating, _ = strconv.ParseFloat(s[BookIndices["Rating"]], 64)
	target.Isbn = s[BookIndices["Isbn"]]
	target.Isbn13 = s[BookIndices["Isbn13"]]
	target.Language = s[BookIndices["Language"]]
	target.Pages, _ = strconv.ParseInt(s[BookIndices["Pages"]], 10, 64)
	target.TotalRating, _ = strconv.ParseInt(s[BookIndices["TotalRating"]], 10, 64)
	target.ReleaseDate, _ = time.Parse("2006-01-02", s[TmdbIndices["ReleaseDate"]])
	target.Publisher = s[BookIndices["Publisher"]]
	*data = target
	return nil

}

func InsertIntoAuthorsChunked(db *sql.DB, chunkSize *int) func(data *Insertable) error {
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
					// TODO: Znalezc sposob na znalezienie dlugosci tego
					var queryFields []string = []string{}
					var argFields []any = []any{}
					for _, item := range chunk {
						bi, ok := (*item).(*BookInsertable)
						if !ok {
							continue
						}

						idTracker.mu.Lock()
						if _, ok := idTracker.idsSeen[bi.BookId]; ok {
							idTracker.mu.Unlock()
							continue
						}

						idTracker.idsSeen[bi.BookId] = true
						idTracker.mu.Unlock()
						for _, author := range bi.Authors {
							queryFields = append(queryFields, ip.Fields)
							argFields = append(argFields, bi.BookId)
							argFields = append(argFields, author)
						}
					}

					// here put insert statements
					<-c
					stmt := fmt.Sprintf("%v %v", ip.Content, strings.Join(queryFields, ","))
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

func InsertIntoBooksChunked(db *sql.DB, chunkSize *int) func(data *Insertable) error {
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
					var queryFields []string = make([]string, 0, NBookFields)
					var argFields []any = make([]any, 0, NBookFields*(*chunkSize))
					for _, item := range chunk {
						bi, ok := (*item).(*BookInsertable)
						if !ok {
							continue
						}

						idTracker.mu.Lock()
						if _, ok := idTracker.idsSeen[bi.BookId]; ok {
							idTracker.mu.Unlock()
							continue
						}
						idTracker.idsSeen[bi.BookId] = true
						idTracker.mu.Unlock()
						queryFields = append(queryFields, ip.Fields)
						argFields = append(argFields, bi.BookId)
						argFields = append(argFields, bi.Title)
						argFields = append(argFields, bi.Rating)
						argFields = append(argFields, bi.Isbn)
						argFields = append(argFields, bi.Isbn13)
						argFields = append(argFields, bi.Language)
						argFields = append(argFields, bi.Pages)
						argFields = append(argFields, bi.TotalRating)
						argFields = append(argFields, bi.ReleaseDate)
						argFields = append(argFields, bi.Publisher)
					}

					// here put insert statements
					<-c
					stmt := fmt.Sprintf("%v %v", ip.Content, strings.Join(queryFields, ","))
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
