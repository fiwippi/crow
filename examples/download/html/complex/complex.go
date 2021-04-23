package main

import (
	"flag"
	"fmt"
	"github.com/fiwippi/crow/pkg/api"
	"github.com/fiwippi/crow/pkg/archiver"
	"log"
	"sync"
	"time"
)

// Key for maps to take less space than bool
type void struct{}

var exists void

var id = flag.String("thread", "", "The id of the thread, e.g. \"570368\"")
var board = flag.String("board", "", "The board directory, e.g. \"po\"")
var outputDir = flag.String("output-dir", "./", "Directory to output the archived thread, slashes of importance")
var overwrite = flag.Bool("overwrite", false, "Whether to overwrite files which already exist")
var validateMD5 = flag.Bool("validate-md5", true, "Whether to validate the MD5 hash of files")
var runOnce = flag.Bool("run-once", false, "Download the thread once and exit without checking for updates")
var interval = flag.Duration("interval", 5*time.Minute, "How often to check if the thread updated")

func watch(c *api.Client, t *api.Thread, a *archiver.Archiver) {
	done := make(chan void)     // Signals the program to break the for-select loop
	first := make(chan void, 1) // To get the timer to "tick" instantly then this channel is used
	wg := &sync.WaitGroup{}     // Ensures all downloads are finished before thread exit

	// We can't refresh a nil thread so we keep a copy of it with the data used for refreshing
	cache := &api.Thread{
		Board: t.Board,
		No:    t.No,
	}

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
		t, mod, err := c.RefreshThread(cache, true, "http")
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
				t, _, err = c.RefreshThread(cache, false, "http")
				if err != nil && t.Archived {
					close(done)
				}
			}
			continue
		}

		// Thread has changed to so update lastCall
		lastCall = time.Now()

		// Archive the thread
		err = a.ArchiveThread(t, *overwrite, *validateMD5)
		if err != nil {
			log.Printf("Error archiving thread, err: %s\n", err)
		}

		if *runOnce {
			close(done)
		}
	}
}

func main() {
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

	// Archive the thread
	a := archiver.NewArchiver(c, *outputDir)
	fmt.Println("Watching...")
	watch(c, t, a)
	fmt.Println("Done...")
}
