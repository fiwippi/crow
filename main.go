package main

import (
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fiwippi/crow/internal/archiver"
	"github.com/fiwippi/crow/internal/log"
	"github.com/fiwippi/crow/pkg/api"
)

func main() {
	dst := flag.String("dst", "./", "Destination dir")
	overwrite := flag.Bool("overwrite", false, "Whether to overwrite files which already exist")
	validateMD5 := flag.Bool("validate-md5", true, "Whether to validate the MD5 hash of files")
	runOnce := flag.Bool("run-once", false, "Download the thread once and exit without checking for updates")
	filesOnly := flag.Bool("files-only", false, "Whether to archive only the files and not the html page of the thread")
	interval := flag.Duration("interval", 5*time.Minute, "How often to check if the thread updated")

	flag.Usage = func() {
		fmt.Println("Usage:")
		fmt.Println("  ./crow po 570368")
		fmt.Println("  ./crow po/thread/570368")
		fmt.Println("  ./crow https://boards.4channel.org/po/thread/570368")
		fmt.Println()
		flag.PrintDefaults()
	}
	flag.Parse()

	var thread int
	var board string
	if len(flag.Args()) == 1 {
		// Attempt to parse a URL if one argument
		link, err := url.Parse(flag.Args()[0])
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse 4chan url")
		}

		parts := strings.Split(strings.Trim(link.Path, "/"), "/")
		if len(parts) != 3 {
			log.Fatal().Err(fmt.Errorf("could not split url into 3 parts: %s", parts)).Msg("failed to parse 4chan url")
		}

		board = parts[0]
		t, err := strconv.Atoi(parts[2])
		if err != nil {
			log.Fatal().Err(fmt.Errorf("could not parse thread id as int: %s", parts[2])).Msg("failed to parse 4chan url")
		}
		thread = t
	} else if len(flag.Args()) == 2 {
		// Attempt to parse "board thread" if two arguments,
		// e.g. "po 570368"
		board = flag.Args()[0]
		t, err := strconv.Atoi(flag.Args()[1])
		if err != nil {
			log.Fatal().Err(fmt.Errorf("could not parse thread id as int: %s", flag.Args()[1])).Msg("failed to parse 4chan url")
		}
		thread = t
	} else {
		flag.Usage()
		os.Exit(1)
	}

	// Create the client and retrieve the thread
	c := api.DefaultClient()
	cache, modified, err := c.GetThread(board, thread)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get thread")
	} else if !modified {
		log.Fatal().Err(err).Msg("retrieved thread is not modified")
	}

	// Watch the thread
	done := make(chan struct{})     // Signals the program to break the for-select loop
	first := make(chan struct{}, 1) // To get the timer to "tick" instantly then this channel is used
	wg := &sync.WaitGroup{}         // Ensures all downloads are finished before thread exit

	// Ticker to check the thread at intervals
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()
	// lastCall keeps track of how long it has been seen the thread has last been modified
	var lastCall = time.Now()

	// Check the thread at intervals until it 404s or is archived
	first <- struct{}{}
	for {
		select {
		case <-done:
			wg.Wait()
			return
		case <-first:
			cache.ClearLastModified()
		case <-ticker.C:
		}

		// Get the newest version of the thread
		t, mod, err := c.RefreshThread(cache)

		if err == api.ErrNotFound {
			log.Info().Msg("Thread 404d")
			close(done)
			continue
		} else if err != nil {
			log.Error().Err(err).Int("no", t.No).Str("board", t.Board).Msg("error refreshing thread")
			continue
		} else if !mod {
			// Check if the thread is archived if it's been 1 hour since last modified,
			// if it has then stop archiving the thread. If it's been more than 3 days
			// then hard close the watching
			if time.Since(lastCall) > 72*time.Hour {
				close(done)
			} else if time.Since(lastCall) > 1*time.Hour {
				t, _, err = c.RefreshThread(cache)
				if err != nil && t.Archived {
					close(done)
				}
			}
			continue
		}

		// Thread has changed so update lastCall
		lastCall = time.Now()

		// Archive the thread
		if *filesOnly {
			for _, p := range t.Posts {
				if p.HasFile {
					m, err := c.GetFile(p)
					if err != nil {
						log.Error().Err(err).Str("filename", p.Filename+p.Ext).Msg("failed to download file")
						continue
					}
					log.Info().Str("filepath", *dst+p.Filename+p.Ext).Msg("saving file")
					go save(m, *dst, t)
				}
			}
		} else {
			err = archiver.Archive(c, t, *dst, *overwrite, *validateMD5)
			if err != nil {
				log.Error().Err(err).Msg("error archiving thread")
			}
		}

		if *runOnce {
			close(done)
		}
	}
}

func save(m *api.Media, dst string, t *api.Thread) {
	defer m.Body.Close()

	if !strings.HasSuffix(dst, "/") {
		dst += "/"
	}
	dst += fmt.Sprintf("4chan/%s/%d/", t.Board, t.No)

	err := os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Str("dir", dst).Msg("failed to make folder directory")
		return
	}

	// Create the file on the host
	out, err := os.Create(dst + m.Filename + m.Ext)
	if err != nil {
		log.Error().Err(err).Str("filename", m.Filename+m.Ext).Msg("failed to create file")
		return
	}
	defer out.Close()

	// Copy the contents to the file
	_, err = io.Copy(out, m.Body)
	if err != nil {
		log.Error().Err(err).Str("filename", m.Filename+m.Ext).Msg("failed to write to file")
		return
	}
}
