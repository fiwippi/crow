package archiver

import (
	"io"
	"strings"

	"golang.org/x/net/html"

	"github.com/fiwippi/crow/internal/log"
	"github.com/fiwippi/crow/pkg/api"
)

// Removes the advert divs from the doc
func removeUnwanted(n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "div" {
			for _, v := range c.Attr {
				if v.Key == "class" {
					if strings.Contains(v.Val, "adg-rects") {
						prevSibling := c.PrevSibling
						n.RemoveChild(c)
						c = prevSibling
					}
				}
			}
		}
		removeUnwanted(c)
	}
}

// Downloads assets and redirects assets and images to local counterparts
func redirect(n *html.Node, a *archiver, t *api.Thread) {
	switch n.Data {
	case "a":
		redirectA(n, t)
	case "link":
		redirectLink(n, a)
	case "script":
		redirectScript(n, a)
	case "img":
		redirectImage(n, a, t)
	case "div":
		redirectDiv(n, a)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		redirect(c, a, t)
	}
}

// Formats the downloaded HTML page to redirect links to static assets and remove
// unwanted javascript which loads ads
func (a *archiver) formatHTML(data io.Reader, errChan chan error, htmlChan chan *html.Node, t *api.Thread) {
	log.Info().Int("no", t.No).Str("board", t.Board).Msg("formatting HTML data")
	defer a.wg.Done()

	// Parse the html data into a doc
	doc, err := html.Parse(data)
	if err != nil {
		errChan <- err
	}

	// Downloads all assets and removes unwanted html elements in the page
	redirect(doc, a, t)
	removeUnwanted(doc)

	errChan <- nil
	htmlChan <- doc

	log.Info().Int("no", t.No).Str("board", t.Board).Msg("done formatting HTML data")
}
