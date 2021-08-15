package archiver

import (
	"bytes"
	_ "embed"
	"io"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/fiwippi/crow/pkg/api"
)

//go:embed assets/icons/arrow_down.png
var iconArrowDown []byte

//go:embed assets/icons/arrow_down2.png
var iconArrowDown2 []byte

//go:embed assets/icons/arrow_right.png
var iconArrowRight []byte

//go:embed assets/icons/arrow_up.png
var iconArrowUp []byte

//go:embed assets/icons/cross.png
var iconCross []byte

//go:embed assets/icons/post_expand_minus.png
var iconPostExpandMinus []byte

//go:embed assets/icons/post_expand_plus.png
var iconPostExpandPlus []byte

//go:embed assets/icons/post_expand_rotate.gif
var iconPostExpandRotate []byte

//go:embed assets/icons/refresh.png
var iconRefresh []byte

//go:embed assets/icons/report.png
var iconReport []byte

// Saves all the embedded icons to the assets dir
func (a *Archiver) saveIcons(t *api.Thread) {
	defer a.d[t.No].wg.Done()

	path := a.d[t.No].assetDir + "image/buttons/burichan/"
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Str("dir", "image/buttons/burichan/").Msg("Failed to create icons dir")
		return
	}

	saveIcon(path+"arrow_down.png", iconArrowDown)
	saveIcon(path+"arrow_down2.png", iconArrowDown2)
	saveIcon(path+"arrow_right.png", iconArrowRight)
	saveIcon(path+"arrow_up.png", iconArrowUp)
	saveIcon(path+"cross.png", iconCross)
	saveIcon(path+"post_expand_minus.png", iconPostExpandMinus)
	saveIcon(path+"post_expand_plus.png", iconPostExpandPlus)
	saveIcon(path+"post_expand_rotate.gif", iconPostExpandRotate)
	saveIcon(path+"refresh.png", iconRefresh)
	saveIcon(path+"report.png", iconReport)

	log.Info().Int("no", t.No).Str("board", t.Board).Msg("Finished saving icons")
}

func saveIcon(path string, data []byte) {
	// Create the file on the host
	out, err := os.Create(path)
	if err != nil {
		log.Error().Err(err).Str("file", path).Msg("Failed to create icon")
		return
	}
	defer out.Close()

	// Copy the contents to the file
	_, err = io.Copy(out, bytes.NewReader(data))
	if err != nil {
		log.Error().Err(err).Str("file", path).Msg("Failed to write icon")
		return
	}
}
