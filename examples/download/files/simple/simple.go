package main

import (
	"flag"
	"fmt"
	"github.com/fiwippi/crow/pkg/api"
	"io"
	"log"
	"os"
	"time"
)

var id        = flag.String("thread", "", "The id of the thread, e.g. \"570368\"")
var board     = flag.String("board", "", "The board directory, e.g. \"po\"")
var outputDir = flag.String("output-dir", "./", "Directory to output the downloaded images")
var dryRun    = flag.Bool("dry", false, "Show output of program without saving files to disk")

func save(m *api.Media) {
	defer m.Body.Close()

	if !*dryRun {
		// Create the file on the host
		out, err := os.Create(*outputDir + m.FilenameExt)
		if err != nil {
			log.Printf("Failed to create file: %s, err: %s\n", m.FilenameExt, err)
			return
		}
		defer out.Close()

		// Copy the contents to the file
		_, err = io.Copy(out, m.Body)
		if err != nil {
			log.Printf("Failed to write to file: %s, err: %s\n", m.FilenameExt, err)
			return
		}
	}
}

func main() {
	// Parse flags
	flag.Parse()
	if *id == "" || *board == "" {
		log.Fatalf("Must specify a thread ID (e.g. \"570368\") and a board directory (e.g. \"po\")\n")
	}

	// Create the client and retrieve the thread
	c := api.DefaultClient()
	t, _, err := c.GetThread(*id, *board, false, "http")
	if err != nil {
		log.Fatalf("Failed to get the thread: %s\n", err)
	}

	// Start downloading files from the thread
	fmt.Println("Downloading...")
	for _, p := range t.Posts {
		if p.HasFile {
			m, err := c.GetFileFromPost(p, "http")
			if err != nil {
				log.Printf("Failed to download file: %s\n", p.Filename+p.Ext)
				continue
			}

			fmt.Printf("%s Saving: '%s'\n", time.Now().Format("15:04:05"), *outputDir + m.FilenameExt)
			go save(m)
		}
	}
	fmt.Println("Done!")
}