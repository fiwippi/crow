package api

import (
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var gmt *time.Location

func init() {
	var err error
	gmt, err = time.LoadLocation("GMT")
	if err != nil {
		log.Fatalf("Failed to load GMT time: %s\n", err)
	}
}

type Client struct {
	apiHTTPClient    *http.Client         // Client used to make requests to the api endpoint
	mediaHTTPClient  *http.Client         // Client used to make requests to the media and static endpoints
	apiRateLimiter   *rate.Limiter        // Should not be more than 1 request second
	mediaRateLimiter *rate.Limiter        // Find limit for this
	lastAccessed     map[string]time.Time // Keeps track of the last time an endpoint was accessed to see if an If-Modified-Since header should be set
	mu               sync.Mutex           // Synchronises access when editing the lastAccessed map
}

//
func formatURL(subdomain, board, endpoint string, scheme HTTPScheme) string {
	if board != "" {
		board = strings.Trim(board, "/")
		return fmt.Sprintf("%s://%s/%s/%s", scheme, subdomain, board, endpoint)
	}
	return fmt.Sprintf("%s://%s/%s", scheme, subdomain, endpoint)
}

// Returns client with at most 1 request to the api per second
// and 4 requests per sec to media endpoints
func DefaultClient() *Client {
	return NewClient(1, 8)
}

// How many requests to make to the api and media endpoints
// per second
func NewClient(apiPerSec, mediaPerSec int) *Client {
	return &Client{
		apiHTTPClient:    &http.Client{Timeout: 30 * time.Second},
		mediaHTTPClient:  &http.Client{Timeout: 30 * time.Second},
		apiRateLimiter:   rate.NewLimiter(rate.Every(time.Second / time.Duration(apiPerSec)), 1),
		mediaRateLimiter: rate.NewLimiter(rate.Every(time.Second / time.Duration(mediaPerSec)), 1),
		mu:               sync.Mutex{},
		lastAccessed:     make(map[string]time.Time),
	}
}

// Clears the lastAccessed map
func (c *Client) ResetLastAccessed() {
	c.mu.Lock()
	c.lastAccessed = make(map[string]time.Time)
	c.mu.Unlock()
}

// Error returned as long as the status code is not 200 or 304
// If ifsm then does not set an If-Modified-Since Header but always keeps track of lastAccessed
func (c *Client) sendRequest(req *http.Request, ifsm bool, subdomain string) (*http.Response, error) {
	// Setting the If-Modified-Since Header
	if ifsm {
		t, found := c.lastAccessed[req.URL.String()]
		if found {
			req.Header.Set("If-Modified-Since", t.Format("Mon, 02 Jan 2006 15:04:05 GMT"))
		}
	}

	// Rate limiting and send the request
	var rl *rate.Limiter
	var client *http.Client
	if subdomain == APIDomain {
		client = c.apiHTTPClient
		rl = c.apiRateLimiter
	} else if subdomain == MediaDomain || subdomain == StaticDomain || subdomain == BoardsDomain {
		client = c.mediaHTTPClient
		rl = c.mediaRateLimiter
	} else {
		return nil, ErrEndpointType
	}

	ctx := context.Background()
	err := rl.Wait(ctx) // This is a blocking call. Honors the rate limit
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Sets the If-Set-Modified Headers
	if ifsm {
		c.mu.Lock()
		c.lastAccessed[req.URL.String()] = time.Now().In(gmt)
		c.mu.Unlock()
	}

	// If the status code is not 200 or 304 then return an error
	if resp.StatusCode != 200 && resp.StatusCode != 304 {
		return nil, ErrNotFound
	}

	return resp, nil
}

// Wraps sendRequest to be able to specify the method and url already
func (c *Client) createAndSendRequest(method, subdomain, board, endpoint string, scheme HTTPScheme, ifsm bool) (*http.Response, error){
	if scheme == "" {
		scheme = HTTP
	} else if !validScheme(scheme) {
		return nil, ErrInvalidScheme
	}

	url := formatURL(subdomain, board, endpoint, scheme)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendRequest(req, ifsm, subdomain)
	if err != nil {
		return nil, err
	}
	return resp, nil
}