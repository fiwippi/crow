package archiver

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/fiwippi/crow/internal/log"
	"github.com/fiwippi/crow/pkg/api"
)

// Saves a media file to a specified output directory
func (a *archiver) saveFile(m *api.Media, dir string, count, total int, item string) {
	defer m.Body.Close()
	defer a.wg.Done()

	// Info
	if item != "" {
		item = " " + item
	}
	if total != 0 {
		log.Debug().Str("file", dir+m.ID+m.Ext).Msg(fmt.Sprintf("saving%s... [%d/%d]", item, count, total))
	} else {
		log.Debug().Str("file", dir+m.ID+m.Ext).Msg(fmt.Sprintf("saving%s...", item))
	}

	// Ensure the directory exists
	x := strings.Split(dir+m.ID+m.Ext, "/")
	fullDir := strings.Join(x[:len(x)-1], "/")
	if _, err := os.Stat(fullDir); os.IsNotExist(err) {
		err := os.MkdirAll(fullDir, os.ModePerm)
		if err != nil {
			log.Error().Err(err).Str("dir", fullDir).Msg("failed to create directory")
			return
		}
	}

	// Create the file on the host
	out, err := os.Create(dir + m.ID + m.Ext)
	if err != nil {
		log.Error().Err(err).Str("file", m.ID+m.Ext).Msg("failed to create file")
		return
	}
	defer out.Close()

	// Copy the contents to the file
	_, err = io.Copy(out, m.Body)
	if err != nil {
		log.Error().Err(err).Str("file", m.ID+m.Ext).Msg("failed to write file")
		return
	}
}

// Downloads all files and thumbnails from a thread
func (a *archiver) dlThreadFiles(t *api.Thread) {
	defer a.wg.Done()

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
		_, found := a.downloaded[strconv.Itoa(p.No)]
		// For every unvisited post with a file
		if !found && p.HasFile {
			// Always download thumbnails, cannot verify MD5 of thumbnail so always save
			m, err := a.c.GetThumbnail(p)
			if err != nil {
				log.Error().Err(err).Str("file", p.Filename+"s.jpg").Msg("failed to download file thumbnail")
				count += 1
			} else {
				a.wg.Add(1)
				go a.saveFile(m, a.thumbDir, count, total, "images")
				count += 1
			}

			// Download images if they dont exist or if overwriting true
			a.downloaded[strconv.Itoa(t.No)] = struct{}{}
			if !a.overwrite && fileExists(a.imgDir+p.ImageID.String()+p.Ext) {
				log.Debug().Str("file", a.imgDir+p.ID+p.Ext).Msg("file already exists, not overwriting")
				count += 1
				continue
			}
			m, err = a.c.GetFile(p)
			if err != nil {
				log.Error().Err(err).Str("file", p.ImageID.String()+p.Ext).Msg("failed to download file")
				count += 1
				continue
			}

			// Ensure it has a valid MD5 Base64 encoded hash
			if a.md5 && !api.VerifyMD5(p, m) {
				log.Error().Str("file", p.ImageID.String()+p.Ext).Msg("MD5 hash of download does not match api supplied MD5, retrying...")
				m, err := a.c.GetFile(p)
				if err != nil || !api.VerifyMD5(p, m) {
					log.Error().Err(err).Str("file", p.ImageID.String()+p.Ext).Msg("retry download failed, download failed for: %s")
					count += 1
					continue
				}
			}

			// Download the file
			a.wg.Add(1)
			go a.saveFile(m, a.imgDir, count, total, "images")
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
