package metrics

import (
	"log"
	"time"
)

var gmt *time.Location

func init() {
	var err error
	gmt, err = time.LoadLocation("GMT")
	if err != nil {
		log.Fatalf("Failed to load GMT time: %s\n", err)
	}
}
