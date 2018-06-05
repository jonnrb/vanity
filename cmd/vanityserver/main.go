package main // go.jonnrb.io/vanity

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go.jonnrb.io/vanity"
)

var host = os.Getenv("HOST")

func serveRepo(mux *http.ServeMux, root, vcsScheme, vcsHost, user, repo string) {
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

func main() {
	mux := http.NewServeMux()

	if f, err := os.Open("/repos"); err != nil {
		log.Fatalf("Error opening conf: %v", err)
	} else {
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			fields := strings.Fields(sc.Text())
			switch len(fields) {
			case 0:
				continue
			case 5:
				// Pass
			default:
				log.Fatalf("Expected line of form \"path vcsScheme vcsHost user repo\" but got %q", sc.Text())
			}

			path, vcsScheme, vcsHost, user, repo :=
				fields[0], fields[1], fields[2], fields[3], fields[4]
			serveRepo(mux, path, vcsScheme, vcsHost, user, repo)
		}
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
