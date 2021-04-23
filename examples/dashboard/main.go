package main

import (
	"flag"
	metrics "github.com/fiwippi/examples/dashboard/pkg"
)

var boards = []string{"/a/", "/w/", "/g/", "/tv/", "/wg/", "/mu/", "/wsg/", "/biz/", "/b/", "/pol/"}

func main() {
	silent := flag.Bool("silent", false, "Whether to log writing to the db")
	flag.Parse()

	metrics.Run(boards, *silent)
}
