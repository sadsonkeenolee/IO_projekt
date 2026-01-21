package utils

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
)

// CsvStreamer streams the content of the CSV, it requires the channel that will
// be streamed into.
func CsvStreamer(path *string, l *log.Logger, c chan<- []string) error {
	if l == nil {
		l = log.New(os.Stderr, "", log.LstdFlags|log.Lmsgprefix)
		l.Println("No logger provided, using a default one.")
	}

	fd, err := os.Open(*path)
	if err != nil {
		return err
	}
	defer fd.Close()
	defer close(c)
	csvReader := csv.NewReader(fd)
	fmt.Printf("Reading %v\n", *path)

	// read until EOF
	for {
		record, err := csvReader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			continue
		}

		c <- record
	}
	return nil
}
