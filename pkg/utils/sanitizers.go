package utils

import (
	"os"
	"strings"
)

func PrepareFetchUrl(item string) string {
	url := os.Getenv("TMDB_FETCH_URL")
	url = strings.ReplaceAll(url, " ", "%20")
	return url
}
