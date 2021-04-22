package api

import (
	"io"
	"strconv"
)

func (c *Client) GetThreadHTML(t *Thread, scheme HTTPScheme) (io.ReadCloser, error) {
	resp, err := c.createAndSendRequest("GET", BoardsDomain, t.Board, "thread/" + strconv.Itoa(t.No), scheme, false)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
