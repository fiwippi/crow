package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

const (
	boardsEndpoint     = "boards.json"  // A list of all boards and their attributes.
	catalogEndpoint    = "catalog.json" // A JSON representation of a board catalog. Includes all OPs and their preview replies.
	archiveEndpoint    = "archive.json" // A list of all closed threads in a board archive. Archived threads no longer receive posts.
	threadListEndpoint = "threads.json" // A summarized list of all threads on a board including thread numbers, their modification time and reply count.

)

// boards.json

func (c *Client) GetBoards() (*Boards, bool, error) {
	return c.getBoards(time.Time{})
}

func (c *Client) getBoards(t time.Time) (*Boards, bool, error) {
	resp, mt, err := c.get(ApiDomain, "", boardsEndpoint, t)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 304 {
		return nil, false, nil
	}

	var b Boards
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}
	err = json.Unmarshal(buf, &b)
	if err != nil {
		return nil, false, err
	}

	b.modTime = modTime(mt)
	return &b, true, nil
}

func (c *Client) RefreshBoards(b *Boards) (*Boards, bool, error) {
	boards, mod, err := c.getBoards(b.modTime.time())
	if err == nil && mod {
		b.modTime = boards.modTime
	}
	return boards, mod, err
}

// threads.json

func (c *Client) GetThreads(board string) (*ThreadList, bool, error) {
	return c.getThreads(board, time.Time{})
}

func (c *Client) getThreads(board string, t time.Time) (*ThreadList, bool, error) {
	resp, mt, err := c.get(ApiDomain, board, threadListEndpoint, t)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 304 {
		return nil, false, nil
	}

	tl := &ThreadList{
		modTime: modTime(mt),
		Board:   strings.Trim(board, "/"),
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	err = json.Unmarshal(buf, &tl.Pages)
	if err != nil {
		return nil, true, err
	}

	return tl, true, nil
}

func (c *Client) RefreshThreads(tl *ThreadList) (*ThreadList, bool, error) {
	threadlist, mod, err := c.getThreads(tl.Board, tl.modTime.time())
	if err == nil && mod {
		tl.modTime = threadlist.modTime
	}
	return threadlist, mod, err
}

// catalog.json

func (c *Client) GetCatalog(board string) (*Catalog, bool, error) {
	return c.getCatalog(board, time.Time{})
}

func (c *Client) getCatalog(board string, t time.Time) (*Catalog, bool, error) {
	resp, mt, err := c.get(ApiDomain, board, catalogEndpoint, t)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 304 {
		return nil, false, nil
	}

	ctl := &Catalog{
		modTime: modTime(mt),
		Board:   strings.Trim(board, "/"),
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	err = json.Unmarshal(buf, &ctl.Pages)
	if err != nil {
		return nil, true, err
	}

	for i := range ctl.Pages {
		for j := range ctl.Pages[i].Threads {
			ctl.Pages[i].Threads[j].Board = ctl.Board
			ctl.Pages[i].Threads[j].HasFile = ctl.Pages[i].Threads[j].Filesize > 0
		}
	}

	return ctl, true, nil
}

func (c *Client) RefreshCatalog(ctl *Catalog) (*Catalog, bool, error) {
	catalog, mod, err := c.getCatalog(ctl.Board, ctl.modTime.time())
	if err == nil && mod {
		ctl.modTime = catalog.modTime
	}
	return catalog, mod, err
}

// archive.json

func (c *Client) GetArchive(board string) (*Archive, bool, error) {
	return c.getArchive(board, time.Time{})
}

func (c *Client) getArchive(board string, t time.Time) (*Archive, bool, error) {
	resp, mt, err := c.get(ApiDomain, board, archiveEndpoint, t)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 304 {
		return nil, false, nil
	}

	var p []int
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	err = json.Unmarshal(buf, &p)
	if err != nil {
		return nil, true, err
	}

	a := &Archive{
		modTime: modTime(mt),
		Board:   strings.Trim(board, "/"),
		PostIDs: p,
	}

	return a, true, nil
}

func (c *Client) RefreshArchive(a *Archive) (*Archive, bool, error) {
	archive, mod, err := c.getArchive(a.Board, a.modTime.time())
	if err == nil && mod {
		a.modTime = archive.modTime
	}
	return archive, mod, err
}

// [board]/[1-15].json

func (c *Client) GetPage(board string, page int) (*Page, bool, error) {
	return c.getPage(board, page, time.Time{})
}

func (c *Client) getPage(board string, page int, t time.Time) (*Page, bool, error) {
	if page < 1 || page > 15 {
		return nil, false, fmt.Errorf("invalid page num, should be in the range 1-15 inclusive")
	}

	resp, mt, err := c.get(ApiDomain, board, strconv.Itoa(page)+".json", t)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 304 {
		return nil, false, nil
	}

	var p Page
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	err = json.Unmarshal(buf, &p)
	if err != nil {
		return nil, true, err
	}

	p.No = page
	p.Board = strings.Trim(board, "/")
	p.modTime = modTime(mt)

	for i := range p.Threads {
		for j := range p.Threads[i].Posts {
			p.Threads[i].Posts[j].Board = p.Board
			if p.Threads[i].Posts[j].Filesize > 0 {
				p.Threads[i].Posts[j].HasFile = true
				p.Threads[i].addThreadAttributes(modTime(mt))
			}
		}
	}

	return &p, true, nil
}

func (c *Client) RefreshPage(p *Page) (*Page, bool, error) {
	page, mod, err := c.getPage(p.Board, p.No, p.modTime.time())
	if err == nil && mod {
		p.modTime = page.modTime
	}
	return page, mod, err
}

// [board]/thread/[op ID].json

func (c *Client) GetThread(board string, opID int) (*Thread, bool, error) {
	return c.getThread(board, opID, time.Time{})
}

func (c *Client) getThread(board string, opID int, t time.Time) (*Thread, bool, error) {
	resp, mt, err := c.get(ApiDomain, board, fmt.Sprintf("thread/%d.json", opID), t)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 304 {
		return nil, false, nil
	}

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

	p.addThreadAttributes(modTime(mt))
	p.Board = strings.Trim(board, "/")

	return &p, true, nil
}

func (c *Client) RefreshThread(th *Thread) (*Thread, bool, error) {
	thread, mod, err := c.getThread(th.Board, th.No, th.modTime.time())
	if err == nil && mod {
		th.modTime = thread.modTime
	}
	return thread, mod, err
}
