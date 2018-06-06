/*
Runs a barebones vanity server over HTTP.

Usage

  ./vanityserver [-index] fqdn [repo file]

The "-index" flag enables an index page at "/" that lists all repos hosted on
this server.

If repo file is not given, "./repos" is used. The file has the following format:

  pkgroot  vcsScheme://vcsHost/user/repo
  pkgroot2 vcsScheme://vcsHost/user/repo2

vcsHost is either a Gogs server (that's what I use) or github.com. I'm open to
supporting other VCSs but I'm not sure what that would look like.
*/
package main // go.jonnrb.io/vanity

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go.jonnrb.io/vanity"
)

var (
	showIndex = flag.Bool("index", false, "Show a list of repos at /")
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
	indexMap := map[string]string{}

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

		if *showIndex {
			indexMap[fields[0]] = fields[1]
		}

		path := fields[0]
		u, err := url.Parse(fields[1])
		if err != nil {
			log.Fatalf("Repo was not a valid URL: %q", fields[1])
		}

		serveRepo(mux, path, u)
	}

	if !*showIndex {
		return
	}

	var b bytes.Buffer
	err := template.Must(template.New("").Parse(`<!DOCTYPE html>
<table>
{{ $host := .Host }}
<h1>{{ html $host }}</h1>
{{ range $root, $repo := .IndexMap }}
<tr>
<td><a href="https://{{ html $host }}/{{ html $root }}">{{ html $root }}</a></td>
<td><a href="{{ html $repo }}">{{ html $repo }}</a></td>
{{ else }}
Nothing here.
{{ end }}
</table>
`)).Execute(&b, struct {
		IndexMap map[string]string
		Host     string
	}{
		IndexMap: indexMap,
		Host:     host,
	})
	if err != nil {
		log.Fatalf("Couldn't create index page: %v", err)
	}
	buf := b.Bytes()

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		io.Copy(w, bytes.NewReader(buf))
	}))
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s fqdn [repos file]", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	host = flag.Arg(0)
	if host == "" {
		flag.Usage()
		os.Exit(-1)
	}

	reposPath := "repos"
	if override := flag.Arg(1); override != "" {
		reposPath = override
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
