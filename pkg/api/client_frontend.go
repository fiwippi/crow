package api

import (
	"fmt"
	"io"
	"time"
)

func (c *Client) GetThreadHTML(t *Thread) (io.ReadCloser, error) {
	// The thread HTML page is a different object to an actual thread so
	// we treat it as if it's always fresh so we supply a time.Time{} so
	// we don't send an If-Modified-Since header. To only fetch the HTML
	// page if there's new content then first use the *Thread object to
	// check if it's updated and then if needed call this function
	resp, _, err := c.get(BoardsDomain, t.Board, fmt.Sprintf("thread/%d", t.No), time.Time{})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
