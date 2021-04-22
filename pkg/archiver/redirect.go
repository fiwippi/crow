package archiver

import (
	"bytes"
	"github.com/fiwippi/crow/pkg/api"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
)

func redirectA(n *html.Node, t *api.Thread) {
	for i, v := range n.Attr {
		if v.Key == "href" && strings.Contains(v.Val, api.MediaDomain) {
			endpoint := strings.TrimPrefix(v.Val, "//" + api.MediaDomain + "/" + t.Board + "/")
			if strings.Contains(endpoint, "s") {
				n.Attr[i].Val = "thumbs/"+ endpoint
			} else {
				n.Attr[i].Val = "images/" + endpoint
			}
		}
	}
}

func redirectLink(n *html.Node, a *Archiver, t *api.Thread) {
	for i, v := range n.Attr {
		if v.Key == "href" && strings.Contains(v.Val, api.StaticDomain) {
			// Download the linked static asset
			endpoint := strings.TrimPrefix(v.Val, "//" + api.StaticDomain + "/")
			m, err := a.c.GetStaticAsset(endpoint, "http")
			if err != nil {
				log.Error().Err(err).Str("file", endpoint).Msg("Failed to download file")
				continue
			}

			if strings.HasPrefix(endpoint, "css") {
				// Download assets in the css script
				b, err := ioutil.ReadAll(m.Body)
				m.Body.Close()
				if err != nil {
					log.Error().Err(err).Str("file", endpoint).Msg("Failed to read css script body")
					continue
				}

				oldEndpoints := make([]string, 0)
				newEndpoints := make([]string, 0)

				scriptStr := string(b)
				regexStr := regexp.QuoteMeta("url(")
				urlRegex := regexp.MustCompile(regexStr)
				matches := urlRegex.FindAllStringIndex(scriptStr, -1)
				if matches != nil {
					for _, v := range matches {
						match := scriptStr[v[0]:v[1]+50]
						bracketIndex := strings.Index(match, ")")
						staticURL := scriptStr[v[0]:v[0]+bracketIndex]

						endpoint := strings.TrimPrefix(staticURL, "url(/")
						oldEndpoints = append(oldEndpoints, endpoint)
						endpoint = strings.TrimPrefix(endpoint, "/s.4cdn.org/")
						newEndpoints = append(newEndpoints, "\"" + strings.ReplaceAll(endpoint, "image", "assets") + "\"")

						_, found := a.d[t.No].downloaded[endpoint]
						if !found {
							a.d[t.No].downloaded[endpoint] = exists
							assetM, err := a.c.GetStaticAsset(endpoint, "http")
							if err != nil {
								log.Error().Err(err).Str("file", endpoint).Msg("Failed to download file")
								continue
							}

							a.d[t.No].wg.Add(1)
							go a.saveFile(assetM, a.d[t.No].assetDir, t, 0, 0, "assets")
						}
					}
				}

				for i, str := range oldEndpoints {
					if strings.HasPrefix(str, "/") { // Add the extra "/" for changing "//s.4cdn..."
						str = "/" + str
					}
					regexStr = regexp.QuoteMeta(str)
					urlRegex := regexp.MustCompile(regexStr)
					scriptStr = string(urlRegex.ReplaceAll([]byte(scriptStr), []byte(newEndpoints[i])))
				}
				urlRegex = regexp.MustCompile(`/"assets/`)
				scriptStr = string(urlRegex.ReplaceAll([]byte(scriptStr), []byte(`"assets/`)))

				// Feed the new media to the save function
				reader := io.NopCloser(strings.NewReader(scriptStr))
				m.Body = reader

				a.d[t.No].wg.Add(1)
				go a.saveFile(m, a.d[t.No].cssDir, t, 0, 0, "css")
			} else {
				a.d[t.No].wg.Add(1)
				go a.saveFile(m, a.d[t.No].assetDir, t, 0, 0, "assets")
			}

			// Reflect the change in the HTML document
			n.Attr[i].Val = strings.ReplaceAll(endpoint, "image", "assets")
		}
	}
}

func redirectImage(n *html.Node, a *Archiver, t *api.Thread) {
	for i, v := range n.Attr {
		if v.Key == "src" {
			// If the image is media then it's already being downloaded in dlThreadFiles so only redirect url
			if strings.Contains(v.Val, api.MediaDomain) {
				endpoint := strings.TrimPrefix(v.Val, "//" + api.MediaDomain + "/" + t.Board + "/")
				if strings.Contains(endpoint, "s") {
					n.Attr[i].Val = "thumbs/"+ endpoint
				} else {
					n.Attr[i].Val = "images/" + endpoint
				}
			}

			// If it's a static asset then download it
			if strings.Contains(v.Val, api.StaticDomain) {
				endpoint := strings.TrimPrefix(v.Val, "//" + api.StaticDomain + "/")
				_, found := a.d[t.No].downloaded[endpoint]
				if !found {
					a.d[t.No].downloaded[endpoint] = exists
					m, err := a.c.GetStaticAsset(endpoint, "http")
					if err != nil {
						log.Error().Err(err).Str("file", endpoint).Msg("Failed to download file")
						continue
					}

					a.d[t.No].wg.Add(1)
					go a.saveFile(m, a.d[t.No].assetDir, t, 0, 0, "assets")
				}

				// Remove all slashes in the endpoint to get to the filename
				link := strings.ReplaceAll(endpoint, "image", "assets")
				n.Attr[i].Val = link
			}
		}
	}
}

func redirectDiv(n *html.Node, a *Archiver, t *api.Thread) {
	for _, v := range n.Attr {
		// Downloads the title banner
		if v.Key == "data-src" {
			endpoint := "/image/title/" + v.Val
			_, found := a.d[t.No].downloaded[endpoint]
			if !found {
				a.d[t.No].downloaded[endpoint] = exists
				m, err := a.c.GetStaticAsset(endpoint, "http")
				if err != nil {
					log.Error().Err(err).Str("file", endpoint).Msg("Failed to download file")
					continue
				}

				a.d[t.No].wg.Add(1)
				go a.saveFile(m, a.d[t.No].assetDir, t, 0, 0, "assets")
			}
		}
	}
}

func redirectScript(n *html.Node, a *Archiver, t *api.Thread) {
	for i, v := range n.Attr {
		// Remove unwanted advertisement script
		if v.Key == "src" && strings.Contains(v.Val, "bid.glass") {
			n.Attr[i].Val = ""
		}

		// Download the wanted js scripts
		if v.Key == "src" && strings.Contains(v.Val, api.StaticDomain) {
			// Download the scripts
			endpoint := strings.TrimPrefix(v.Val, "//" + api.StaticDomain + "/")
			m, err := a.c.GetStaticAsset(endpoint, "http")
			if err != nil {
				log.Error().Err(err).Str("file", endpoint).Msg("Failed to download file")
				continue
			}

			// Remove js from the scripts related to advertisements
			b, err := ioutil.ReadAll(m.Body)
			m.Body.Close()
			if err != nil {
				log.Error().Err(err).Str("file", endpoint).Msg("Failed to read js script body")
				continue
			}

			var scriptStr string
			end := strings.Index(string(b), "function initAnalytics(){")
			start := strings.Index(string(b), "function applySearch(e){")
			if end != -1 && start != -1 {
				scriptStr = string(b)[:end] + string(b)[start:]
			}

			end = strings.Index(scriptStr, "initAdsAG(),initAdsAT(),initAdsBG(),initAdsLD(),initAdsBGLS()")
			start = strings.Index(scriptStr, "document.post&&(document.post.name.value=get_cookie(\"4chan_name\")")
			if end != -1 && start != -1 {
				scriptStr = scriptStr[:end] + scriptStr[start:]
			}

			end = strings.Index(scriptStr, "initAnalytics()")
			start = strings.Index(scriptStr, "s=(r=location.pathname.split(/\\//))[1],window.passEnabled&&setPassMsg()")
			if end != -1 && start != -1 {
				scriptStr = scriptStr[:end] + scriptStr[start:]
			}

			// Change the link for ui icons
			regexStr := `//s.4cdn.org/`
			urlRegex := regexp.MustCompile(regexStr)
			if scriptStr != "" {
				scriptStr = string(urlRegex.ReplaceAll([]byte(scriptStr), []byte("assets/")))
			} else {
				scriptStr = string(urlRegex.ReplaceAll(b, []byte("assets/")))
			}

			// Feed the new media to the save function
			var reader io.ReadCloser
			if scriptStr != "" {
				reader = io.NopCloser(strings.NewReader(scriptStr))
			} else {
				reader = io.NopCloser(bytes.NewReader(b))
			}
			m.Body = reader

			a.d[t.No].wg.Add(1)
			go a.saveFile(m, a.d[t.No].jsDir, t, 0, 0, "script")

			// Change v.Val
			n.Attr[i].Val = endpoint
		}
	}
}