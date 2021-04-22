package api

type HTTPScheme string // Make API requests using the same protocol as the app. Only use SSL when a user is accessing your app over HTTPS.

const (
	HTTP  HTTPScheme = "http"
	HTTPS HTTPScheme = "https"
)

func validScheme(scheme HTTPScheme) bool {
	return scheme == HTTP || scheme == HTTPS
}
