package api

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

var (
	ErrInvalidSpoilerNum  = fmt.Errorf("invalid spoiler num, should be in the range 1-5 inclusive")
	ErrInvalidAssetFormat = fmt.Errorf("assets should be formatted as 'folder/name.ext' or 'folder/name.num.ext'")
)

type Media struct {
	Body     io.ReadCloser // The response body
	Board    string        // Board the media is from
	ID, Ext  string        // The image ID and extension. The extension has the dot at the beginning.
	Filename string        // Filename of the image if applicable, only works ig FromPost() method is used
	URL      string        // URL to the image resource
	MD5      string        // Base64 encoded MD5 hash of the response content
}

func (c *Client) getFileFromID(id, ext, filename, domain, board, endpoint string) (*Media, error) {
	id = strings.ToLower(id)
	endpoint = strings.ToLower(endpoint)

	resp, _, err := c.get(domain, board, endpoint, time.Time{})
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	hash := md5.Sum(data)
	media := Media{
		Body:     ioutil.NopCloser(bytes.NewReader(data)),
		Filename: filename,
		ID:       id,
		Ext:      ext,
		URL:      c.url(domain, board, endpoint),
		MD5:      base64.StdEncoding.EncodeToString(hash[:]),
	}

	return &media, nil
}

func (c *Client) GetFile(p *Post) (*Media, error) {
	return c.getFileFromID(p.ImageID.String(), p.Ext, p.Filename, MediaDomainA, p.Board, p.ImageID.String()+p.Ext)
}

func (c *Client) GetThumbnail(p *Post) (*Media, error) {
	return c.getFileFromID(p.ImageID.String()+"s", ".jpg", p.Filename+"-s", MediaDomainA, p.Board, p.ImageID.String()+"s.jpg")
}

func (c *Client) GetFlag(flagCode string) (*Media, error) {
	return c.getFileFromID(flagCode, ".gif", flagCode, StaticDomain, "image/country", flagCode+".gif")
}

func (c *Client) GetTrollFlag(flagCode string) (*Media, error) {
	return c.getFileFromID(flagCode, ".gif", flagCode, StaticDomain, "image/country/troll", flagCode+".gif")
}

func (c *Client) GetCustomSpoiler(board string, num int) (*Media, error) {
	if num < 1 || num > 5 {
		return nil, ErrInvalidSpoilerNum
	}

	spoiler := fmt.Sprintf("spoiler-%s%d", board, num)
	return c.getFileFromID(spoiler, ".png", spoiler, StaticDomain, "image/", spoiler+".png")

}

func (c *Client) GetStaticAsset(endpoint string) (*Media, error) {
	elem := strings.Split(endpoint, "/")
	if len(elem) < 2 {
		return nil, ErrInvalidAssetFormat
	}
	board, route := elem[0]+"/", strings.Join(elem[1:], "/")

	elem = strings.Split(route, ".")
	if len(elem) < 2 {
		return nil, ErrInvalidAssetFormat
	}
	name := strings.Join(elem[:len(elem)-1], ".")
	ext := "." + elem[len(elem)-1]

	return c.getFileFromID(name, ext, name, StaticDomain, board, name+ext)
}

func VerifyMD5(p *Post, m *Media) bool {
	return p.MD5 == m.MD5
}
