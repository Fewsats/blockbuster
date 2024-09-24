package lightning

import (
	"net/http"
)

// HTTPClient is the interface for making HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
