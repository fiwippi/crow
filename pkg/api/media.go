package api

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

type Media struct {
	Body        io.ReadCloser // The response body
	Board       string        // Board the media is from
	ID, Ext     string        // The image ID and extension
	Filename    string        // Filename of the image if applicable, only works ig FromPost() method is used
	FilenameExt string        // The Filename and Ext added together
	URL         string        // URL to the image resource
	MD5         string        // Base64 encoded MD5 hash of the response content
}

func (c *Client) getFileFromID(id, ext, filename, subdomain, board, endpoint string, scheme HTTPScheme) (*Media, error) {
	id = strings.ToLower(id)
	endpoint = strings.ToLower(endpoint)

	resp, err := c.createAndSendRequest("GET", subdomain, board, endpoint, scheme, false)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	hash := md5.Sum(data)
	media := Media{
		Body:        ioutil.NopCloser(bytes.NewReader(data)),
		Filename:    filename,
		ID:          id,
		Ext:         ext,
		URL:         formatURL(subdomain, board, endpoint, scheme),
		FilenameExt: filename + ext,
		MD5:         base64.StdEncoding.EncodeToString(hash[:]),
	}

	return &media, nil
}

func (c *Client) GetFileFromID(id, ext, board string, scheme HTTPScheme) (*Media, error) {
	return c.getFileFromID(id, ext, "", MediaDomain, board, id+ext, scheme)
}

func (c *Client) GetFileFromPost(p *Post, scheme HTTPScheme) (*Media, error) {
	return c.getFileFromID(p.ImageID.String(), p.Ext, p.Filename, MediaDomain, p.Board, p.ImageID.String()+p.Ext, scheme)
}

func (c *Client) GetThumbnailFromID(id, board string, scheme HTTPScheme) (*Media, error) {
	return c.GetFileFromID(id+"s", ".jpg", board, scheme)
}

func (c *Client) GetThumbnailFromPost(p *Post, scheme HTTPScheme) (*Media, error) {
	return c.getFileFromID(p.ImageID.String()+"s", ".jpg", p.Filename+"-s", MediaDomain, p.Board, p.ImageID.String()+"s.jpg", scheme)
}

func (c *Client) GetFlagFromCode(flagCode string, scheme HTTPScheme) (*Media, error) {
	return c.getFileFromID(flagCode, ".gif", flagCode, StaticDomain, "image/country", flagCode+".gif", scheme)
}

func (c *Client) GetTrollFlagFromCode(flagCode string, scheme HTTPScheme) (*Media, error) {
	return c.getFileFromID(flagCode, ".gif", flagCode, StaticDomain, "image/country/troll", flagCode+".gif", scheme)
}

func (c *Client) GetCustomSpoiler(num int, board string, scheme HTTPScheme) (*Media, error) {
	if num < 1 || num > 5 {
		return nil, ErrInvalidSpoilerNum
	}

	spoiler := fmt.Sprintf("spoiler-%s%d", board, num)
	return c.getFileFromID(spoiler, ".png", spoiler, StaticDomain, "image/", spoiler+".png", scheme)

}

func (c *Client) GetStaticAsset(endpoint string, scheme HTTPScheme) (*Media, error) {
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

	return c.getFileFromID(name, ext, name, StaticDomain, board, name+ext, scheme)
}

func VerifyMD5(p *Post, m *Media) bool {
	return p.MD5 == m.MD5
}
