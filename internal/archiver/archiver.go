package archiver

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"

	"github.com/fiwippi/crow/internal/log"
	"github.com/fiwippi/crow/pkg/api"
)

// Archiver archives a thread
type archiver struct {
	c          *api.Client
	wg         *sync.WaitGroup
	downloaded map[string]struct{} // Keeps track of files which have already been downloaded

	// Settings for downloading files
	overwrite bool // Whether to overwrite files which already exist
	md5       bool // Whether to validate MD5 of downloaded images

	// Output directories
	outputDir string // Dirs where to save files
	thumbDir  string // Sub-dir to save thumbnails
	imgDir    string // Sub-dir to save images
	cssDir    string // Sub-dir to save css
	jsDir     string // Sub-dir to save javascript
	assetDir  string // Sub-dir to save static assets
}

func Archive(c *api.Client, t *api.Thread, dst string, overwrite, md5 bool) error {
	// Ensure valid thread
	if t == nil {
		return fmt.Errorf("thread is invalid since it's nil")
	}

	// Format the destination directory
	if dst == "" {
		dst = "./"
	}
	if strings.HasSuffix(dst, "/") {
		strings.TrimSuffix(dst, "/")
	}
	dst = fmt.Sprintf("%s/4chan/%s/%d/", dst, t.Board, t.No)

	// Create the archiver
	a := &archiver{
		c:          c,
		wg:         &sync.WaitGroup{},
		downloaded: make(map[string]struct{}),
		overwrite:  overwrite,
		md5:        md5,
		outputDir:  dst,
		thumbDir:   fmt.Sprintf("%s%s/", dst, "thumbs"),
		imgDir:     fmt.Sprintf("%s%s/", dst, "images"),
		cssDir:     fmt.Sprintf("%s%s/", dst, "css"),
		jsDir:      fmt.Sprintf("%s%s/", dst, "js"),
		assetDir:   fmt.Sprintf("%s%s/", dst, "assets"),
	}
	start := time.Now()

	// Download the thread's HTML page
	data, err := a.c.GetThreadHTML(t)
	if err != nil {
		return err
	}
	defer data.Close()

	// Begin downloading the thread images
	log.Info().Int("no", t.No).Str("board", t.Board).Msg("downloading files")
	a.wg.Add(1)
	go a.dlThreadFiles(t)

	// Save the icons to the asset dir
	log.Info().Int("no", t.No).Str("board", t.Board).Msg("saving icons")
	a.wg.Add(1)
	go a.saveIcons(t)

	// Process the html by linking to local assets
	nodes := make(chan *html.Node)
	errs := make(chan error)
	a.wg.Add(1)
	go a.formatHTML(data, errs, nodes, t)
	err = <-errs
	if err != nil {
		log.Error().Err(err).Msg("failed to parse html doc")
		close(nodes)
		return err
	}
	root := <-nodes

	// Write the html to a file
	f, err := os.Create(fmt.Sprintf("%s/thread.html", a.outputDir))
	if err != nil {
		return err
	}
	defer f.Close()

	log.Info().Int("no", t.No).Str("board", t.Board).Msg("rendering HTML...")
	err = html.Render(f, root)
	if err != nil {
		log.Error().Err(err).Msg("failed to render html to file")
		return err
	}
	log.Info().Int("no", t.No).Str("board", t.Board).Msg("done rendering...")

	// Wait until everything is downloaded
	log.Info().Int("no", t.No).Str("board", t.Board).Msg("ensuring downloading completed...")
	a.wg.Wait()
	end := time.Since(start).Round(time.Second)

	log.Info().Int("no", t.No).Str("time_taken", end.String()).Str("board", t.Board).Msg("archiving done!")
	return nil
}
