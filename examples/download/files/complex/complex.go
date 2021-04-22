package main

import (
	"flag"
	"fmt"
	"github.com/fiwippi/crow/pkg/api"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type void struct {}
var exists void

var id          = flag.String("thread", "", "The id of the thread, e.g. \"570368\"")
var board       = flag.String("board", "", "The board directory, e.g. \"po\"")
var outputDir   = flag.String("output-dir", "./", "Directory to output the downloaded images")
var dryRun      = flag.Bool("dry", false, "Show output of program without saving files to disk")
var runOnce     = flag.Bool("run-once", false, "Download the thread once and exit without checking for updates")
var interval    = flag.Duration("interval", 5 * time.Minute, "How often to check if the thread updated")
var validateMD5 = flag.Bool("valid-md5", true, "Only download files with valid MD5 hashes")

func save(m *api.Media, wg *sync.WaitGroup) {
	defer m.Body.Close()
	defer wg.Done()

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

func download(t *api.Thread, c *api.Client) {
	visited := make(map[int]void) // Keeps track of what posts in a thread have been seen
	done := make(chan void)       // Signals the program to break the for-select loop
	first := make(chan void, 1)   // To get the timer to "tick" instantly then this channel is used
	wg := &sync.WaitGroup{}       // Ensures all downloads are finished before thread exit

	// Ticker to check the thread at intervals
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()
	// lastCall keeps track of how long it has been seen the thread has last been modified
	var lastCall time.Time

	// Check the thread at intervals until it 404s or is archived
	first <- exists
	for {
		select {
		case <-done:
			wg.Wait()
			return
		case <-first:
		case <-ticker.C:
		}

		// Get the newest version of the thread
		t, mod, err := c.RefreshThread(t, true, "http")
		if err == api.ErrNotFound {
			log.Println("Thread 404d")
			close(done)
			continue
		} else if err != nil {
			log.Printf("Error when refreshing thread /%s/%d: %s\n", t.Board, t.No, err)
			continue
		} else if !mod {
			// Check if the thread is archived if it's been 1 hour since last modified,
			// if it has then stop archiving the thread
			if time.Since(lastCall) > 1*time.Hour {
				t, _, err = c.RefreshThread(t, false, "http")
				if err != nil && t.Archived {
					close(done)
				}
			}
			continue
		}

		// Thread has changed to so update lastCall
		lastCall = time.Now()

		// Download all new files in the thread
		for _, p := range t.Posts {
			_, found := visited[p.No]
			if !found && p.HasFile {
				// Mark the post as visited
				visited[p.No] = exists

				// Download the file
				m, err := c.GetFileFromPost(p, "http")
				if err != nil {
					log.Printf("Failed to download file: %s\n", p.Filename+p.Ext)
					continue
				}

				// Ensure it has a valid MD5 Base64 encoded hash
				if *validateMD5 && !api.VerifyMD5(p, m) {
					fmt.Printf("MD5 hash of download does not match api supplied MD5: %s, retrying...\n", m.FilenameExt)
					m, err = c.GetFileFromPost(p, "http")
					if (err != nil) || (err == nil && !api.VerifyMD5(p, m)) {
						log.Printf("Retry failed, download failed for: %s\n", p.Filename+p.Ext)
						continue
					}
				}

				// Save the file
				fmt.Printf("%s Saving: '%s'\n", time.Now().Format("15:04:05"), *outputDir + m.FilenameExt)
				wg.Add(1)
				go save(m, wg)
			}
		}

		if *runOnce {
			close(done)
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
	thread, _, err := c.GetThread(*id, *board, false, "http")
	if err != nil {
		log.Fatalf("Failed to get the thread: %s\n", err)
	}

	// Start watching the thread
	fmt.Println("Watching...")
	download(thread, c)
	fmt.Println("Done!")
}