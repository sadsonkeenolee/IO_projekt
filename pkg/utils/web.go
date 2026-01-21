package utils

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SanitizeUrl constructs a link from url and item and replaces all the spaces
// to normalize the url.
func SanitizeUrl(url, item string) string {
	return strings.ReplaceAll(fmt.Sprintf(url, item), " ", "%20")
}

// DoRequest attempts to make a given request, on success returns the response
// in bytes
func DoRequest(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
