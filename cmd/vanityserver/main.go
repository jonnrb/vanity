/*
Runs a barebones vanity server over HTTP.
*/
package main // go.jonnrb.io/vanity

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go.jonnrb.io/vanity"
)

var host string

func serveRepo(mux *http.ServeMux, root string, u *url.URL) {
	vcsScheme, vcsHost := u.Scheme, u.Host

	// Get ["", "user", "repo"].
	pathParts := strings.Split(u.Path, "/")
	if len(pathParts) != 3 {
		log.Fatalf("Repo URL must be of the form vcsScheme://vcsHost/user/repo but got %q", u.String())
	}
	user, repo := pathParts[1], pathParts[2]

	importPath := host + "/" + root
	var h http.Handler
	if vcsHost == "github.com" {
		h = vanity.GitHubHandler(importPath, user, repo, vcsScheme)
	} else {
		h = vanity.GogsHandler(importPath, vcsHost, user, repo, vcsScheme)
	}
	mux.Handle("/"+root, h)
	mux.Handle("/"+root+"/", h)
}

func buildMux(mux *http.ServeMux, r io.Reader) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		switch len(fields) {
		case 0:
			continue
		case 2:
			// Pass
		default:
			log.Fatalf("Expected line of form \"path vcsScheme://vcsHost/user/repo\" but got %q", sc.Text())
		}

		path := fields[0]
		u, err := url.Parse(fields[1])
		if err != nil {
			log.Fatalf("Repo was not a valid URL: %q", fields[1])
		}

		serveRepo(mux, path, u)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s fqdn [repos file]", os.Args[0])
		os.Exit(-1)
	}
	host = os.Args[1]

	reposPath := "repos"
	if len(os.Args) > 2 {
		reposPath = os.Args[2]
	}

	mux := http.NewServeMux()

	if f, err := os.Open(reposPath); err != nil {
		log.Fatalf("Error opening repos path: %v", err)
	} else {
		buildMux(mux, f)
	}

	srv := &http.Server{
		// This should be sufficient.
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,

		Addr:    ":8080",
		Handler: mux,
	}

	log.Println(srv.ListenAndServe())
}
