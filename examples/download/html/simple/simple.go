package main

import (
	"flag"
	"github.com/fiwippi/crow/pkg/api"
	"github.com/fiwippi/crow/pkg/archiver"
	"log"
)

var id        = flag.String("thread", "", "The id of the thread, e.g. \"570368\"")
var board     = flag.String("board", "", "The board directory, e.g. \"po\"")
var outputDir = flag.String("output-dir", "./", "Directory to output the archived thread, slashes of importance")
var overwrite = flag.Bool("overwrite", false, "Whether to overwrite files which already exist")
var validateMD5 = flag.Bool("validate-md5", true, "Whether to validate the MD5 hash of files")

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
	err = a.ArchiveThread(t, *overwrite, *validateMD5)
	if err != nil {
		log.Fatal(err)
	}
}
