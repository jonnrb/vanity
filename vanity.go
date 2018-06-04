package vanity // import "go.jonnrb.io/vanity"

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
)

type importData struct {
	ImportRoot string
	VCS        string
	VCSRoot    string
}

type tag func(r *http.Request) (io.Reader, error)

var goImportTmpl = template.Must(template.New("main").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="{{.ImportRoot}} {{.VCS}} {{.VCSRoot}}">
</head>
</html>
`))

func ImportTag(vcs, importPath, repoRoot string) tag {
	return func(r *http.Request) (io.Reader, error) {
		d := &importData{
			ImportRoot: r.Host + r.URL.Path,
			VCS:        vcs,
			VCSRoot:    repoRoot,
		}
		var buf bytes.Buffer
		return &buf, goImportTmpl.Execute(&buf, d)
	}
}

func Handle(t tag) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Redirect to https.
		if r.URL.Scheme == "http" {
			r.URL.Scheme = "https"
			http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
			return
		}

		// Only method supported is GET.
		if r.Method != http.MethodGet {
			status := http.StatusMethodNotAllowed
			http.Error(w, http.StatusText(status), status)
			return
		}

		// Redirect browsers to gddo.
		if r.FormValue("go-get") != "1" {
			url := "https://godoc.org/" + r.Host + r.URL.Path
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			return
		}

		body, err := t(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else {
			w.Header().Set("Cache-Control", "public, max-age=300")
			io.Copy(w, body)
		}
	})
}

// Redirect is a HTTP middleware that redirects browsers to godoc.org or
// Go tool to VCS repository.
func Redirect(vcs, importPath, repoRoot string) http.Handler {
	return Handle(ImportTag(vcs, importPath, repoRoot))
}
