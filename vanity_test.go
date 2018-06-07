package vanity

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGoTool(t *testing.T) {
	tests := []struct {
		path   string
		result string
	}{
		{"/pkg?go-get=1", "go.jonnrb.io/pkg git https://github.com/jonnrb/pkg"},
		{"/pkg/?go-get=1", "go.jonnrb.io/pkg git https://github.com/jonnrb/pkg"},
		{"/pkg/subpkg?go-get=1", "go.jonnrb.io/pkg git https://github.com/jonnrb/pkg"},
	}
	for _, test := range tests {
		res := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "go.jonnrb.io"+test.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		h := GitHubHandler("go.jonnrb.io/pkg", "jonnrb", "pkg", "https")
		h.ServeHTTP(res, req)

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("reading response body failed with error: %v", err)
		}

		expected := `<meta name="go-import" content="` + test.result + `">`
		if !strings.Contains(string(body), expected) {
			t.Fatalf("Expecting url '%v' body to contain html meta tag: '%v', but got:\n'%v'", test.path, expected, string(body))
		}

		expected = "text/html; charset=utf-8"
		if res.HeaderMap.Get("content-type") != expected {
			t.Fatalf("Expecting content type '%v', but got '%v'", expected, res.HeaderMap.Get("content-type"))
		}

		if res.Code != http.StatusOK {
			t.Fatalf("Expected response status 200, but got %v", res.Code)
		}
	}
}

func TestBrowserGoDoc(t *testing.T) {
	tests := []struct {
		path   string
		result string
	}{
		{"/pkg", "https://godoc.org/go.jonnrb.io/pkg"},
		{"/pkg/", "https://godoc.org/go.jonnrb.io/pkg"},
		{"/pkg/sub/foo", "https://godoc.org/go.jonnrb.io/pkg/sub/foo"},
	}
	for _, test := range tests {
		res := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "go.jonnrb.io"+test.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		srv := GitHubHandler("go.jonnrb.io/pkg", "jonnrb", "pkg", "https")
		srv.ServeHTTP(res, req)

		if res.Code != http.StatusTemporaryRedirect {
			t.Fatalf("Expected response status %v, but got %v", http.StatusTemporaryRedirect, res.Code)
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("reading response body failed with error: %v", err)
		}
		if !strings.Contains(string(body), test.result) {
			t.Fatalf("Expecting '%v' be contained in '%v'", test.result, string(body))
		}
	}
}

func ExampleGitHubHandler() {
	// Redirects the vanity import path "go.jonnrb.io/vanity" to the code hosted
	// on GitHub by user (or organization) "jonnrb" in repo "vanity" using the
	// git over "https".
	h := GitHubHandler("go.jonnrb.io/vanity", "jonnrb", "vanity", "https")

	http.Handle("/vanity", h)
	http.Handle("/vanity/", h) // to handle requests for subpackages.

	http.ListenAndServe(":http", nil)
}
