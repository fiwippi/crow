package api

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
)

const (
	// 4chan Domains
	APIDomain    = "a.4cdn.org" // This domain serves all 4chan API endpoints in the form of static json files.
	MediaDomain  = "i.4cdn.org" // This is the primary content domain used for serving user submitted media attached to posts.
	StaticDomain = "s.4cdn.org" // Serves all static site content including icons, banners, CSS and JavaScript files.
	BoardsDomain = "boards.4chan.org"

	// API Endpoints
	BoardsEndpoint     = "boards.json"  // A list of all boards and their attributes.
	CatalogEndpoint    = "catalog.json" // A JSON representation of a board catalog. Includes all OPs and their preview replies.
	ArchiveEndpoint    = "archive.json" // A list of all closed threads in a board archive. Archived threads no longer receive posts.
	ThreadListEndpoint = "threads.json" // A summarized list of all threads on a board including thread numbers, their modification time and reply count.
)

// Returns the Boards struct, defaults to HTTP if empty string for scheme is specified
func (c *Client) GetBoards(ifsm bool, scheme HTTPScheme) (*Boards, bool, error) {
	resp, err := c.createAndSendRequest("GET", APIDomain, "", BoardsEndpoint, scheme, ifsm)
	if err != nil {
		return nil, false, err
	} else if resp.StatusCode == 304 {
		return nil, false, nil
	}
	defer resp.Body.Close()

	var b Boards
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	err = json.Unmarshal(buf, &b)
	if err != nil {
		return nil, true, err
	}

	return &b, true, nil
}

// Returns the ThreadList struct for a specific board, defaults to HTTP if empty string for scheme is specified
func (c *Client) GetThreads(board string, ifsm bool, scheme HTTPScheme) (*ThreadList, bool, error) {
	resp, err := c.createAndSendRequest("GET", APIDomain, board, ThreadListEndpoint, scheme, ifsm)
	if err != nil {
		return nil, false, err
	} else if resp.StatusCode == 304 {
		return nil, false, nil
	}
	defer resp.Body.Close()

	var pages []ThreadListPage
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	err = json.Unmarshal(buf, &pages)
	if err != nil {
		return nil, true, err
	}

	tl := &ThreadList{
		Board: strings.Trim(board, "/"),
		Pages: pages,
	}

	return tl, true, nil
}

func (c *Client) RefreshThreads(tl *ThreadList, ifsm bool, scheme HTTPScheme) (*ThreadList, bool, error) {
	return c.GetThreads(tl.Board, ifsm, scheme)
}

// Returns the Catalog struct for a specific board, defaults to HTTP if empty string for scheme is specified
func (c *Client) GetCatalog(board string, ifsm bool, scheme HTTPScheme) (*Catalog, bool, error) {
	resp, err := c.createAndSendRequest("GET", APIDomain, board, CatalogEndpoint, scheme, ifsm)
	if err != nil {
		return nil, false, err
	} else if resp.StatusCode == 304 {
		return nil, false, nil
	}
	defer resp.Body.Close()

	var pages []CatalogPage
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	err = json.Unmarshal(buf, &pages)
	if err != nil {
		return nil, true, err
	}

	for i := range pages {
		for j := range pages[i].Threads {
			pages[i].Threads[j].Board = strings.Trim(board, "/")
			if pages[i].Threads[j].Filesize > 0 {
				pages[i].Threads[j].HasFile = true
			}
		}
	}

	ctl := &Catalog{
		Board: strings.Trim(board, "/"),
		Pages: pages,
	}

	return ctl, true, nil
}

func (c *Client) RefreshCatalog(ctl *Catalog, ifsm bool, scheme HTTPScheme) (*Catalog, bool, error) {
	return c.GetCatalog(ctl.Board, ifsm, scheme)
}

// Returns the Archive struct for a specific board, defaults to HTTP if empty string for scheme is specified
// Returns ErrNotFound on 404. If using If-Modified-Since headers and a 304 is received then false is returned
// for the bool and a nil pointer to the struct is returned.
func (c *Client) GetArchive(board string, ifsm bool, scheme HTTPScheme) (*Archive, bool, error) {
	resp, err := c.createAndSendRequest("GET", APIDomain, board, ArchiveEndpoint, scheme, ifsm)
	if err != nil {
		return nil, false, err
	} else if resp.StatusCode == 304 {
		return nil, false, nil
	}
	defer resp.Body.Close()

	var p []int
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	err = json.Unmarshal(buf, &p)
	if err != nil {
		return nil, true, err
	}

	return &Archive{strings.Trim(board, "/"), p}, true, nil
}

func (c *Client) RefreshArchive(a *Archive, ifsm bool, scheme HTTPScheme) (*Archive, bool, error) {
	return c.GetArchive(a.Board, ifsm, scheme)
}

// Returns the Page struct for a specific page on the board. Defaults to HTTP if empty string for scheme is specified
// Returns ErrNotFound on 404. If using If-Modified-Since headers and a 304 is received then false is returned
// for the bool and a nil pointer to the struct is returned.
func (c *Client) GetPage(page int, board string, ifsm bool, scheme HTTPScheme) (*Page, bool, error) {
	if page < 1 || page > 15 {
		return nil, false, ErrInvalidPageNum
	}

	resp, err := c.createAndSendRequest("GET", APIDomain, board, strconv.Itoa(page)+".json", scheme, ifsm)
	if err != nil {
		return nil, false, err
	} else if resp.StatusCode == 304 {
		return nil, false, nil
	}
	defer resp.Body.Close()

	var p Page
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	err = json.Unmarshal(buf, &p)
	if err != nil {
		return nil, true, err
	}

	for i := range p.Threads {
		for j := range p.Threads[i].Posts {
			p.Threads[i].Posts[j].Board = strings.Trim(board, "/")
			if p.Threads[i].Posts[j].Filesize > 0 {
				p.Threads[i].Posts[j].HasFile = true
				p.Threads[i].addThreadAttributes()
			}
		}
	}

	p.No = page
	p.Board = strings.Trim(board, "/")

	return &p, true, nil
}

func (c *Client) RefreshPage(p *Page, ifsm bool, scheme HTTPScheme) (*Page, bool, error) {
	return c.GetPage(p.No, p.Board, ifsm, scheme)
}

// Returns the Thread struct for a specific thread on the board. Defaults to HTTP if empty string for scheme is specified
// Returns ErrNotFound on 404. If using If-Modified-Since headers and a 304 is received then false is returned
// for the bool and a nil pointer to the struct is returned.
func (c *Client) GetThread(opID, board string, ifsm bool, scheme HTTPScheme) (*Thread, bool, error) {
	resp, err := c.createAndSendRequest("GET", APIDomain, board, "thread/"+opID+".json", scheme, ifsm)
	if err != nil {
		return nil, false, err
	} else if resp.StatusCode == 304 {
		return nil, false, nil
	}
	defer resp.Body.Close()

	var p Thread
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	err = json.Unmarshal(buf, &p)
	if err != nil {
		return nil, true, err
	}

	for i := range p.Posts {
		p.Posts[i].Board = strings.Trim(board, "/")
		if p.Posts[i].Filesize > 0 {
			p.Posts[i].HasFile = true
		}

	}

	p.addThreadAttributes()
	p.Board = strings.Trim(board, "/")

	return &p, true, nil
}

func (c *Client) RefreshThread(t *Thread, ifsm bool, scheme HTTPScheme) (*Thread, bool, error) {
	return c.GetThread(strconv.Itoa(t.No), t.Board, ifsm, scheme)
}
