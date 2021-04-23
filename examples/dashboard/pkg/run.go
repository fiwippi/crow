package metrics

import (
	"github.com/fiwippi/crow/pkg/api"
	"sync"
	"time"
)

var silentLog = false
var wg = &sync.WaitGroup{}

func Run(boards []string, silent bool) {
	// Create the api client
	c := api.DefaultClient()
	silentLog = silent

	// Watch for changes
	wg.Add(2)
	go watchPostsPerDur(boards, c, 1*time.Minute)
	go watchThreadsPerDur(boards, c, 1*time.Hour)
	wg.Wait()
}
