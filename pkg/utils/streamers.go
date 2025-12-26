package utils

import (
	"encoding/csv"
	"io"
	"log"
	"os"
)

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

	// read until EOF
	for {
		record, err := csvReader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			l.Printf("Error while parsing: %v", err)
			continue
		}

		c <- record
	}
	return nil
}
