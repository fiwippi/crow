package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

var gmt = func() *time.Location {
	t, err := time.LoadLocation("GMT")
	if err != nil {
		log.Fatalf("failed to load GMT time: %s\n", err)
	}
	return t
}()

var ErrNotFound = fmt.Errorf("404 not found")

const (
	ApiDomain    = "a.4cdn.org"       // This domain serves all 4chan API endpoints in the form of static json files.
	MediaDomainA = "i.4cdn.org"       // This is the primary content domain used for serving user submitted media attached to posts.
	MediaDomainB = "is2.4chan.org"    // Some media files also served here
	StaticDomain = "s.4cdn.org"       // Serves all static site content including icons, banners, CSS and JavaScript files.
	BoardsDomain = "boards.4chan.org" // Serves the front-end html data
)

// Client makes requests to the 4chan api and media/static endpoints
type Client struct {
	api          *http.Client
	media        *http.Client
	apiLimiter   *rate.Limiter // Should be no more than 1 request per second
	mediaLimiter *rate.Limiter

	// SSL decides whether the API should make requests using HTTPS
	// Note:
	//   - Make API requests using the same protocol as the app.
	//   - Only use SSL when a user is accessing your app over HTTPS.
	SSL bool
	// Supplies the If-Modified-Since header, i.e. the time that the URL
	// was last accessed is sent to the API. A response with a 304 status
	// code is returned if no changes have occurred since then (meaning
	// no new content exists)
	IFMS bool
}

// DefaultClient returns client with at most 1 request to the
// api per second and 8 requests per sec to media endpoints.
// It used SSL and the If-Modified-Since header by default.
func DefaultClient() *Client {
	return NewClient(1, 8, true, true)
}

func NewClient(apiPerSec, mediaPerSec int, ssl, ifms bool) *Client {
	return &Client{
		api:          &http.Client{Timeout: 30 * time.Second},
		media:        &http.Client{Timeout: 30 * time.Second},
		apiLimiter:   rate.NewLimiter(rate.Every(time.Second/time.Duration(apiPerSec)), 1),
		mediaLimiter: rate.NewLimiter(rate.Every(time.Second/time.Duration(mediaPerSec)), 1),
		SSL:          ssl,
		IFMS:         ifms,
	}
}

// do sends a request with the appropriate rate limiting for the
// specific subdomain, errors are returned on status codes 400-500
func (c *Client) do(method, domain, board, endpoint string, lastAccessed time.Time) (*http.Response, time.Time, error) {
	// Create the *http.Request
	req, err := http.NewRequest(method, c.url(domain, board, endpoint), nil)
	if err != nil {
		return nil, time.Time{}, err
	}

	// Setting the If-Modified-Since Header
	if c.IFMS && lastAccessed != (time.Time{}) {
		req.Header.Set("If-Modified-Since", lastAccessed.Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	}

	// Choose the correct rate limiter and http client
	var rl *rate.Limiter
	var client *http.Client
	switch domain {
	case ApiDomain:
		client = c.api
		rl = c.apiLimiter
	case MediaDomainA, MediaDomainB, StaticDomain, BoardsDomain:
		client = c.media
		rl = c.mediaLimiter
	default:
		return nil, time.Time{}, fmt.Errorf("request subdomain is invalid: %s", domain)
	}

	// Rate limit if needed and send the request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = rl.Wait(ctx)
	if err != nil {
		return nil, time.Time{}, err
	}
	t := time.Now().In(gmt)
	resp, err := client.Do(req)
	if err != nil {
		return nil, time.Time{}, err
	}

	// Returns an error on status codes 400-500
	if resp.StatusCode == 404 {
		return nil, t, ErrNotFound
	} else if resp.StatusCode >= 400 && resp.StatusCode <= 500 {
		return nil, t, fmt.Errorf("response from api has invalid status: %d", resp.StatusCode)
	}

	return resp, t, nil
}

// get sends a GET request as specified in the do method
func (c *Client) get(domain, board, endpoint string, lastAccessed time.Time) (*http.Response, time.Time, error) {
	return c.do("GET", domain, board, endpoint, lastAccessed)
}

// url formats the request url
func (c *Client) url(domain, board, endpoint string) string {
	scheme := "http"
	if c.SSL {
		scheme = "https"
	}

	if board != "" {
		board := strings.Trim(board, "/")
		return fmt.Sprintf("%s://%s/%s/%s", scheme, domain, board, endpoint)
	} else {
		return fmt.Sprintf("%s://%s/%s", scheme, domain, endpoint)
	}
}
