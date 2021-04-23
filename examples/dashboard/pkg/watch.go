package metrics

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/fiwippi/crow/pkg/api"
	"log"
	"time"
)

func watchPostsPerDur(boards []string, c *api.Client, dur time.Duration) {
	defer wg.Done()

	// Keeps track of all the unique threads on a board
	uniqueThreads := make(map[string]mapset.Set)
	for _, b := range boards {
		uniqueThreads[b] = mapset.NewSet()
	}

	// How many replies does a thread have
	replies := make(map[int]int)

	// Ensures update check every minute
	ticker := time.NewTicker(dur)
	defer ticker.Stop()

	// Stops writing to DB on first attempt since
	// it counts all threads which exist instead of
	// only new threads
	save := false

	done := make(chan void)
	first := make(chan void, 1)
	first <- exists
	for {
		select {
		case <-first:
		case <-done:
			return
		case <-ticker.C:
		}

		for _, b := range boards {
			ctl, _, err := c.GetCatalog(b, false, "http")
			if err != nil {
				log.Printf("Failed to get board catalog: %s, err: %s\n", b, err)
				continue
			}

			count := 0
			for _, p := range ctl.Pages {
				for _, t := range p.Threads {
					// + 1 to include the op
					count += t.Replies + 1 - replies[t.No]
					replies[t.No] = t.Replies + 1
				}
			}

			s := stat{
				Time:     time.Now().In(gmt),
				Name:     "new_posts",
				Duration: dur.String(),
				Board:    b,
				Count:    count,
			}

			if save && !silentLog {
				s.writeDB()
				fmt.Printf("%+v\n", s)
			}
		}

		if !save {
			save = true
		}
	}
}

func watchThreadsPerDur(boards []string, c *api.Client, dur time.Duration) {
	defer wg.Done()

	// Keeps track of all the unique threads on a board
	uniqueThreads := make(map[string]mapset.Set) // Keeps track of all the unique threads on a board
	for _, b := range boards {
		uniqueThreads[b] = mapset.NewSet()
	}

	// Ensures update check every hour
	ticker := time.NewTicker(dur)
	defer ticker.Stop()

	// Stops writing to DB on first attempt since
	// it counts all threads which exist instead of
	// only new threads
	save := false

	done := make(chan void)
	first := make(chan void, 1)
	first <- exists
	for {
		select {
		case <-first:
		case <-done:
			return
		case <-ticker.C:
		}

		for _, b := range boards {
			ctl, _, err := c.GetCatalog(b, false, "http")
			if err != nil {
				log.Printf("Failed to get board catalog: %s, err: %s\n", b, err)
				continue
			}

			set := mapset.NewSet()
			for _, p := range ctl.Pages {
				for _, t := range p.Threads {
					set.Add(t.No)
				}
			}
			unique := uniqueThreads[b].Difference(set)
			uniqueThreads[b] = set

			s := stat{
				Time:     time.Now().In(gmt),
				Name:     "new_threads",
				Duration: dur.String(),
				Board:    b,
				Count:    unique.Cardinality(),
			}

			if save && !silentLog {
				fmt.Printf("%+v\n", s)
				s.writeDB()
			}
		}

		if !save {
			save = true
		}
	}
}
