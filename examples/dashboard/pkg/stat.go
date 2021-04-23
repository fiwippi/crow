package metrics

import "time"

type stat struct {
	Time     time.Time // Time stat was calculated
	Duration string
	Name     string //
	Board    string //
	Count    int    //
}
