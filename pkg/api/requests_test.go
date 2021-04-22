package api

import (
	"os"
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
	catalog, modified, err := mc.GetCatalog("/g/", false, HTTP)
	if catalog == nil {
		t.Errorf("Failure to get catalog, modified: %v err: %s\n", modified, err)
	}
	for i := range catalog.Pages {
		for j := range catalog.Pages[i].Threads {
			if catalog.Pages[i].Threads[j].Board == "" {
				t.Errorf("Board attribute not added to post in Catalog\n")
			}
			if catalog.Pages[i].Threads[j].Filesize > 0 && !catalog.Pages[i].Threads[j].HasFile {
				t.Errorf("HasFile attribute not added to post in Catalog even though filesize > 0\n")
			}
		}
	}

	// Test in GetPage()
	page, modified, err := mc.GetPage(1, "/g/", false, HTTP)
	if page == nil {
		t.Errorf("Failure to get page, modified: %v err: %s\n", modified, err)
	}
	for i := range page.Threads {
		for j := range page.Threads[i].Posts {
			if page.Threads[i].Posts[j].Board == "" {
				t.Errorf("Board attribute not added to post in Page\n")
			}
			if page.Threads[i].Posts[j].Filesize > 0 && !page.Threads[i].Posts[j].HasFile {
				t.Errorf("HasFile attribute not added to post in Page even though filesize > 0\n")
			}
		}
	}

	// Test added in GetThread()
	thread, modified, err := mc.GetThread("570368", "/po/", false, HTTP)
	if thread == nil {
		t.Errorf("Failure to get thread, modified: %v err: %s\n", modified, err)
	}
	for i := range thread.Posts {
		if thread.Posts[i].Board == "" {
			t.Errorf("Board attribute not added to post in Thread\n")
		}
		if thread.Posts[i].Filesize > 0 && !thread.Posts[i].HasFile {
			t.Errorf("HasFile attribute not added to post in Thread even though filesize > 0\n")
		}
	}
}

func TestPageNumValidation(t *testing.T) {
	_, _, err := mc.GetPage(0, "/g/", false, HTTP)
	if err != ErrInvalidPageNum {
		t.Error("Invalid page num '0' but interpreted as valid")
	}

	_, _, err = mc.GetPage(16, "/g/", false, HTTP)
	if err != ErrInvalidPageNum {
		t.Error("Invalid page num '16' but interpreted as valid")
	}

	_, _, err = mc.GetPage(1, "/g/", false, HTTP)
	if err == ErrInvalidPageNum {
		t.Error("Page num '1' valid but interpreted as invalid")
	}

	_, _, err = mc.GetPage(15, "/g/", false, HTTP)
	if err == ErrInvalidPageNum {
		t.Error("Page num '15' valid but interpreted as invalid")
	}
}

func TestIfModifiedSinceHeader(t *testing.T) {
	// Resets the client
	mc.ResetLastAccessed()

	// Checks the first non-fresh request will return a 200
	_, modified, err := mc.GetBoards(true, HTTP)
	if err != nil {
		t.Errorf("Failure to get boards, modified: %v err: %s\n", modified, err)
	}
	if !modified {
		t.Error("Resource should have been modified since last request but returns as unchanged")
	}

	// Checks subsequent requests return a 304
	_, modified, err = mc.GetBoards(true, HTTP)
	if err != nil {
		t.Errorf("Failure to get boards, modified: %v err: %s\n", modified, err)
	}
	if modified {
		t.Error("Resource should have remained unmodified since last request, but changed")
	}
}

func TestRequestWithEmptyScheme(t *testing.T) {
	// Expecting a 200 request with an empty scheme
	_, modified, err := mc.GetBoards(false, "")
	if err != nil || !modified {
		t.Errorf("Failure to get boards, modified: %v err: %s\n", modified, err)
	}
}

func Test404(t *testing.T) {
	// Resets the client
	mc.ResetLastAccessed()

	// Checks Archive which doesn't exist returns 404
	_, _, err := mc.GetArchive("/bant/", false, HTTP)
	if err != ErrNotFound {
		t.Errorf("Archive which doesn't exist did not return 404, error: %d\n", err)
	}

	// Checks page which doesn't exist returns 404
	_, _, err= mc.GetPage(15, "/g/", false, HTTP)
	if err != ErrNotFound {
		t.Errorf("Page which doesn't exist did not return 404, error: %d\n", err)
	}

	// Checks thread which doesn't exist
	_, _, err = mc.GetThread("abcdefg", "/g/", false, HTTP)
	if err != ErrNotFound {
		t.Errorf("Thread which doesn't exist did not return 404, error: %d\n", err)
	}
}

func TestUnmarshalAndBoardAdded(t *testing.T) {
	// Resets the client
	mc.ResetLastAccessed()

	// Checks Boards can be unmarshaled
	board, modified, err := mc.GetBoards(false, HTTP)
	if board == nil {
		t.Errorf("Failure to get boards, modified: %v err: %s\n", modified, err)
	}

	// Checks ThreadList can be unmarshaled
	threadList, modified, err := mc.GetThreads("/g/", false, HTTP)
	if threadList == nil {
		t.Errorf("Failure to get thread list, modified: %v err: %s\n", modified, err)
	}
	if threadList.Board == "" {
		t.Errorf("Board not present for threadlist, modified: %v err: %s\n", modified, err)
	}

	// Checks Catalog can be unmarshaled
	catalog, modified, err := mc.GetCatalog("/g/", false, HTTP)
	if catalog == nil {
		t.Errorf("Failure to get catalog, modified: %v err: %s\n", modified, err)
	}
	if catalog.Board == "" {
		t.Errorf("Board not present for catalog, modified: %v err: %s\n", modified, err)
	}

	// Checks Archive can be unmarshaled
	archive, modified, err := mc.GetArchive("/g/", false, HTTP)
	if archive == nil {
		t.Errorf("Failure to get archive, modified: %v err: %s\n", modified, err)
	}
	if archive.Board == "" {
		t.Errorf("Board not present for archive, modified: %v err: %s\n", modified, err)
	}

	// Checks a page from a board can be unmarshaled
	page, modified, err := mc.GetPage(1, "/g/", false, HTTP)
	if page == nil {
		t.Errorf("Failure to get page, modified: %v err: %s\n", modified, err)
	}
	if page.Board == "" {
		t.Errorf("Board not present for page, modified: %v err: %s\n", modified, err)
	}

	// Checks a thread from a board can be unmarshaled
	thread, modified, err := mc.GetThread("570368", "/po/", false, HTTP)
	if thread == nil {
		t.Errorf("Failure to get thread, modified: %v err: %s\n", modified, err)
	}
	if thread.Board == "" {
		t.Errorf("Board not present for thread, modified: %v err: %s\n", modified, err)
	}
}