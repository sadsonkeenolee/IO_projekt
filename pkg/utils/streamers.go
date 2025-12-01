package utils

import (
	"encoding/csv"
	"io"
	"log"
	"os"
)

func CsvStreamer(path *string, l *log.Logger, c chan<- []string) error {
	if l == nil {
		panic("CsvStreamer: No logging instance provided.")
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
