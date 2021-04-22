package archiver

import (
	"fmt"
	"github.com/fiwippi/crow/pkg/api"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
	"os"
	"strings"
	"time"
)

// Key for maps to take less space than bool
type void struct {}
var exists void

// This should be the entrypoint to the package since NewArchiver()
// should be the first function someone calls so is fine to instantiate
// the log here
func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"})
}

// Archives a thread, supports archiving multiple threads at once
type Archiver struct {
	c         *api.Client    // Client used to make requests
	d         map[int]*state // Holds download state for a specific thread
	outputDir string         // Output directory to save all threads to
}

// Returns a new archiver using an api client
func NewArchiver(c *api.Client, outputDir string) *Archiver {
	if outputDir == "" {
		outputDir = "./"
	}
	if !strings.HasSuffix(outputDir, "/") {
		outputDir += "/"
	}
	outputDir += "4chan/"

	return &Archiver{
		c: c,
		d: make(map[int]*state),
		outputDir: outputDir,
	}
}

// Archives a thread object
func (a *Archiver) ArchiveThread(t *api.Thread, overwrite, validateMD5 bool) error {
	start := time.Now()
	log.Info().Int("no", t.No).Str("board", t.Board).Msg("Archiving thread...")

	// Create the thread state
	a.d[t.No] = createState(a.outputDir, t, overwrite, validateMD5)

	// Get the HTML page of the thread and write it to a file
	data, err := a.c.GetThreadHTML(t, "http")
	if err != nil {
		return err
	}
	defer data.Close()

	// Start downloading all images in the thread
	log.Info().Int("no", t.No).Str("board", t.Board).Msg("Downloading files")
	a.d[t.No].wg.Add(1)
	go a.dlThreadFiles(t)

	// Save the icons to the asset dir
	log.Info().Int("no", t.No).Str("board", t.Board).Msg("Saving icons")
	a.d[t.No].wg.Add(1)
	go a.saveIcons(t)

	// Process the html by linking to local assets
	nodes := make(chan *html.Node)
	errs := make(chan error)
	a.d[t.No].wg.Add(1)
	go a.formatHTML(data, errs, nodes, t)
	err = <- errs
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse html doc")
		return err
	}
	root := <-nodes

	// Write the html to a file
	f, err := os.Create(fmt.Sprintf("%s/thread.html", a.d[t.No].outputDir))
	if err != nil {
		return err
	}
	defer f.Close()

	log.Info().Int("no", t.No).Str("board", t.Board).Msg("Rendering HTML...")
	err = html.Render(f, root)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render html to file")
		return err
	}
	log.Info().Int("no", t.No).Str("board", t.Board).Msg("Done rendering...")

	// Wait until everything is downloaded
	log.Info().Int("no", t.No).Str("board", t.Board).Msg("Ensuring downloading completed...")
	a.d[t.No].wg.Wait()
	end := time.Since(start).Round(time.Second)

	log.Info().Int("no", t.No).Str("time taken", end.String()).Str("board", t.Board).Msg("Archiving Done!")
	return nil
}