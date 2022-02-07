package api

import (
	"os"
	"strings"
	"testing"
)

// Main client used to avoid rate limiting
var mc *Client

func TestMain(m *testing.M) {
	mc = DefaultClient()
	os.Exit(m.Run())
}

func TestPostAndHasFileAdded(t *testing.T) {
	// Test in GetCatalog()
	catalog, modified, err := mc.GetCatalog("/g/")
	if catalog == nil {
		t.Errorf("failure to get catalog, modified: %v err: %s\n", modified, err)
		return
	}
	for i := range catalog.Pages {
		for j := range catalog.Pages[i].Threads {
			if catalog.Pages[i].Threads[j].Board == "" {
				t.Errorf("board attribute not added to post in Catalog\n")
			}
			if catalog.Pages[i].Threads[j].Filesize > 0 && !catalog.Pages[i].Threads[j].HasFile {
				t.Errorf("hasFile attribute not added to post in Catalog even though filesize > 0\n")
			}
		}
	}

	// Test in GetPage()
	page, modified, err := mc.GetPage("/g/", 1)
	if page == nil {
		t.Errorf("failure to get page, modified: %v err: %s\n", modified, err)
		return
	}
	for i := range page.Threads {
		for j := range page.Threads[i].Posts {
			if page.Threads[i].Posts[j].Board == "" {
				t.Errorf("board attribute not added to post in Page\n")
			}
			if page.Threads[i].Posts[j].Filesize > 0 && !page.Threads[i].Posts[j].HasFile {
				t.Errorf("hasFile attribute not added to post in Page even though filesize > 0\n")
			}
		}
	}

	// Test added in GetThread()
	thread, modified, err := mc.GetThread("/po/", 570368)
	if thread == nil {
		t.Errorf("failure to get thread, modified: %v err: %s\n", modified, err)
		return
	}
	for i := range thread.Posts {
		if thread.Posts[i].Board == "" {
			t.Errorf("board attribute not added to post in Thread\n")
		}
		if thread.Posts[i].Filesize > 0 && !thread.Posts[i].HasFile {
			t.Errorf("hasFile attribute not added to post in Thread even though filesize > 0\n")
		}
	}
}

func TestPageNumValidation(t *testing.T) {
	test := func(err error) bool {
		if err == nil {
			return false
		}
		return strings.Contains(err.Error(), "invalid page num")
	}

	_, _, err := mc.GetPage("/g/", 0)
	if !test(err) {
		t.Error("invalid page num '0' but interpreted as valid:", err)
	}

	_, _, err = mc.GetPage("/g/", 16)
	if !test(err) {
		t.Error("invalid page num '16' but interpreted as valid:", err)
	}

	_, _, err = mc.GetPage("/g/", 1)
	if test(err) {
		t.Error("page num '1' valid but interpreted as invalid:", err)
	}

	_, _, err = mc.GetPage("/g/", 15)
	if test(err) {
		t.Error("page num '15' valid but interpreted as invalid:", err)
	}
}

func TestIfModifiedSinceHeader(t *testing.T) {
	mc.IFMS = true

	// Checks the first non-fresh request will return a 200
	b, modified, err := mc.GetBoards()
	if err != nil {
		t.Errorf("failure to get boards, modified: %v err: %s\n", modified, err)
		return
	}
	if !modified {
		t.Error("resource should have been modified since last request but returns as unchanged")
		return
	}

	// Checks subsequent requests return a 304
	_, modified, err = mc.RefreshBoards(b)
	if err != nil {
		t.Errorf("failure to get boards, modified: %v err: %s\n", modified, err)
		return
	}
	if modified {
		t.Error("resource should be unmodified but is modified")
		return
	}

	// Check that once turned off then no it can't be modified
	mc.IFMS = false
	_, modified, err = mc.RefreshBoards(b)
	if err != nil {
		t.Errorf("failure to get boards, modified: %v err: %s\n", modified, err)
		return
	}
	if !modified {
		t.Error("resource should be modified but is unmodified")
		return
	}
}

func Test404(t *testing.T) {
	// Checks Archive which doesn't exist returns 404
	_, _, err := mc.GetArchive("/bant/")
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("archive which doesn't exist did not return 404, error: %s\n", err)
	}

	// Checks page which doesn't exist returns 404
	_, _, err = mc.GetPage("/g/", 15)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("page which doesn't exist did not return 404, error: %s\n", err)
	}

	// Checks thread which doesn't exist
	_, _, err = mc.GetThread("/g/", 999999999999)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("thread which doesn't exist did not return 404, error: %s\n", err)
	}
}

func TestUnmarshalAndBoardAdded(t *testing.T) {
	// Checks Boards can be unmarshalled
	board, modified, err := mc.GetBoards()
	if board == nil {
		t.Errorf("failure to get boards, modified: %v err: %s\n", modified, err)
	}

	// Checks ThreadList can be unmarshalled
	threadList, modified, err := mc.GetThreads("/g/")
	if threadList == nil {
		t.Errorf("failure to get thread list, modified: %v err: %s\n", modified, err)
		return
	}
	if threadList.Board == "" {
		t.Errorf("board not present for threadlist, modified: %v err: %s\n", modified, err)
	}

	// Checks Catalog can be unmarshalled
	catalog, modified, err := mc.GetCatalog("/g/")
	if catalog == nil {
		t.Errorf("failure to get catalog, modified: %v err: %s\n", modified, err)
		return
	}
	if catalog.Board == "" {
		t.Errorf("board not present for catalog, modified: %v err: %s\n", modified, err)
	}

	// Checks Archive can be unmarshalled
	archive, modified, err := mc.GetArchive("/g/")
	if archive == nil {
		t.Errorf("failure to get archive, modified: %v err: %s\n", modified, err)
		return
	}
	if archive.Board == "" {
		t.Errorf("board not present for archive, modified: %v err: %s\n", modified, err)
	}

	// Checks a page from a board can be unmarshalled
	page, modified, err := mc.GetPage("/g/", 1)
	if page == nil {
		t.Errorf("failure to get page, modified: %v err: %s\n", modified, err)
		return
	}
	if page.Board == "" {
		t.Errorf("board not present for page, modified: %v err: %s\n", modified, err)
	}

	// Checks a thread from a board can be unmarshalled
	thread, modified, err := mc.GetThread("/po/", 570368)
	if thread == nil {
		t.Errorf("failure to get thread, modified: %v err: %s\n", modified, err)
		return
	}
	if thread.Board == "" {
		t.Errorf("board not present for thread, modified: %v err: %s\n", modified, err)
	}
}
