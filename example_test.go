package vanity

import (
	"net/http"

	"go.jonnrb.io/vanity"
)

func ExampleRedirect() {
	// Redirects the vanity import path "go.jonnrb.io/vanity" to the code hosted
	// on GitHub by user (or organization) "jonnrb" in repo "vanity" using the
	// git over "https".
	h := vanity.GitHubHandler("go.jonnrb.io/vanity", "jonnrb", "vanity", "https")

	http.Handle("/vanity", h)
	http.Handle("/vanity/", h) // to handle requests for subpackages.
}
