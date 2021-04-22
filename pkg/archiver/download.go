package archiver

import (
	"fmt"
	"github.com/fiwippi/crow/pkg/api"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"strconv"
	"strings"
)

// Saves a media file to a specified output directory
func (a *Archiver) saveFile(m *api.Media, dir string, t *api.Thread, count, total int, item string) {
	defer m.Body.Close()
	defer a.d[t.No].wg.Done()

	// Info
	if item != "" {
		item = " " + item
	}
	if total != 0 {
		log.Debug().Str("file", dir+m.ID+m.Ext).Msg(fmt.Sprintf("Saving%s... [%d/%d]", item, count, total))
	} else {
		log.Debug().Str("file", dir+m.ID+m.Ext).Msg(fmt.Sprintf("Saving%s...", item))
	}

	// Ensure the directory exists
	x := strings.Split(dir + m.ID + m.Ext, "/")
	fullDir := strings.Join(x[:len(x)-1], "/")
	if _, err := os.Stat(fullDir); os.IsNotExist(err) {
		err := os.MkdirAll(fullDir, os.ModePerm)
		if err != nil {
			log.Error().Err(err).Str("dir", fullDir).Msg("Failed to create directory")
			return
		}
	}

	// Create the file on the host
	out, err := os.Create(dir + m.ID + m.Ext)
	if err != nil {
		log.Error().Err(err).Str("file", m.ID + m.Ext).Msg("Failed to create file")
		return
	}
	defer out.Close()

	// Copy the contents to the file
	_, err = io.Copy(out, m.Body)
	if err != nil {
		log.Error().Err(err).Str("file", m.ID + m.Ext).Msg("Failed to write file")
		return
	}
}

// Downloads all files and thumbnails from a thread
func (a *Archiver) dlThreadFiles(t *api.Thread) {
	defer a.d[t.No].wg.Done()

	// Get total number of files
	total := 0
	for _, p := range t.Posts {
		if p.HasFile {
			total += 1
		}
	}
	total *= 2 // Multiply by 2 because thumbnails

	// Download files
	count := 1
	for _, p := range t.Posts {
		_, found := a.d[t.No].downloaded[strconv.Itoa(p.No)]
		// For every unvisited post with a file
		if !found && p.HasFile {
			// Always download thumbnails, cannot verify MD5 of thumbnail so always save
			m, err := a.c.GetThumbnailFromPost(p, "http")
			if err != nil {
				log.Error().Err(err).Str("file", p.Filename + "s.jpg").Msg("Failed to download file thumbnail")
				count += 1
			} else {
				a.d[t.No].wg.Add(1)
				go a.saveFile(m, a.d[t.No].thumbDir, t, count, total, "images")
				count += 1
			}

			// Download images if they dont exist or if overwriting true
			a.d[t.No].downloaded[strconv.Itoa(t.No)] = exists
			if !a.d[t.No].overwrite && fileExists(a.d[t.No].imgDir+p.ImageID.String()+p.Ext) {
				log.Debug().Str("file", a.d[t.No].imgDir+p.ID+p.Ext).Msg("File already exists, not overwriting")
				count += 1
				continue
			}
			m, err = a.c.GetFileFromPost(p, "http")
			if err != nil {
				log.Error().Err(err).Str("file", p.ImageID.String() + p.Ext).Msg("Failed to download file")
				count += 1
				continue
			}

			// Ensure it has a valid MD5 Base64 encoded hash
			if a.d[t.No].md5 && !api.VerifyMD5(p, m) {
				log.Error().Str("file", p.ImageID.String() + p.Ext).Msg("MD5 hash of download does not match api supplied MD5, retrying...")
				m, err := a.c.GetFileFromPost(p, "http")
				if (err != nil) || (err == nil && !api.VerifyMD5(p, m)) {
					log.Printf("Retry failed, download failed for: %s\n", p.Filename+p.Ext)
					count += 1
					continue
				}
			}

			// Download the file
			a.d[t.No].wg.Add(1)
			go a.saveFile(m, a.d[t.No].imgDir, t, count, total, "images")
			count += 1
		}
	}
}

// Determines whether a file exists on the filesystem with the path
func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}