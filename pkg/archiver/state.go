package archiver

import (
	"fmt"
	"sync"

	"github.com/fiwippi/crow/pkg/api"
)

// Holds data for download state of each thread
type state struct {
	downloaded                                           map[string]void // Keeps track of files which have already been downloaded
	wg                                                   *sync.WaitGroup // Ensures all processed done before exiting function
	outputDir, thumbDir, imgDir, cssDir, jsDir, assetDir string          // Dirs where to save files
	overwrite                                            bool            // Whether to overwrite files which already exist
	md5                                                  bool            // Whether to validate MD5 of downloaded images
}

func createState(outputDir string, t *api.Thread, overwrite, validateMD5 bool) *state {
	return &state{
		md5:        validateMD5,
		overwrite:  overwrite,
		downloaded: make(map[string]void),
		wg:         &sync.WaitGroup{},
		outputDir:  fmt.Sprintf("%s%s/%d/", outputDir, t.Board, t.No),
		thumbDir:   fmt.Sprintf("%s%s/%d/%s/", outputDir, t.Board, t.No, "thumbs"),
		imgDir:     fmt.Sprintf("%s%s/%d/%s/", outputDir, t.Board, t.No, "images"),
		cssDir:     fmt.Sprintf("%s%s/%d/%s/", outputDir, t.Board, t.No, "css"),
		jsDir:      fmt.Sprintf("%s%s/%d/%s/", outputDir, t.Board, t.No, "js"),
		assetDir:   fmt.Sprintf("%s%s/%d/%s/", outputDir, t.Board, t.No, "assets"),
	}
}
