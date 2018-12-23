/*
Runs a barebones vanity server over HTTP.

Usage

  ./vanityserver [-index] [-nohealthz] fqdn [repo file]

The "-index" flag enables an index page at "/" that lists all repos hosted on
this server.

The "-nohealthz" flag disables the "/healthz" endpoint that returns a 200 OK
when everything is OK.

The "-watch" flag watches the repo file for changes. When it is updated, the
updated version will be used for serving.

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
	noHealthz = flag.Bool("nohealthz", false, "Disable healthcheck endpoint at /healthz")
	watch     = flag.Bool("watch", false, "Watch repos file for changes and reload")
)

var (
	host      string           // param 1
	reposPath string = "repos" // param 2
)

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

func addRepoHandlers(mux *http.ServeMux, r io.Reader) error {
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
			return fmt.Errorf("expected line of form \"path vcsScheme://vcsHost/user/repo\" but got %q", sc.Text())
		}

		if *showIndex {
			indexMap[fields[0]] = fields[1]
		}

		path := fields[0]
		u, err := url.Parse(fields[1])
		if err != nil {
			return fmt.Errorf("repo was not a valid URL: %q", fields[1])
		}

		serveRepo(mux, path, u)
	}

	if !*showIndex {
		return nil
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
		return fmt.Errorf("couldn't create index page: %v", err)
	}
	buf := b.Bytes()

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		io.Copy(w, bytes.NewReader(buf))
	}))
	return nil
}

func registerHealthz(mux *http.ServeMux, isHealthy func() bool) {
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isHealthy() {
			io.WriteString(w, "OK\r\n")
		} else {
			http.Error(w, "internal error\r\n", http.StatusInternalServerError)
		}
	}))
}

var healthcheck = func() bool {
	return true
}

func generateHandler() (http.Handler, error) {
	mux := http.NewServeMux()

	f, err := os.Open(reposPath)
	if err != nil {
		return nil, fmt.Errorf("error opening %q: %v", reposPath, err)
	}
	if err := addRepoHandlers(mux, f); err != nil {
		return nil, err
	}

	if !*noHealthz {
		registerHealthz(mux, healthcheck)
	}
	return mux, nil
}

func buildServer(h http.Handler) *http.Server {
	return &http.Server{
		// This should be sufficient.
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,

		Addr:    ":8080",
		Handler: h,
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s fqdn [repos file]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	host = flag.Arg(0)
	if host == "" {
		flag.Usage()
		os.Exit(-1)
	}

	if override := flag.Arg(1); override != "" {
		reposPath = override
	}

	var h http.Handler
	if *watch {
		dh := newDynamicHandler(reposPath, generateHandler)
		healthcheck = dh.IsHealthy
		defer dh.Close()
		h = dh
	} else {
		var err error
		h, err = generateHandler()
		if err != nil {
			log.Printf("Error generating handler: %v", err)
		}
	}

	srv := buildServer(h)

	log.Println(srv.ListenAndServe())
}
