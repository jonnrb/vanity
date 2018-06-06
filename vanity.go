package vanity // import "go.jonnrb.io/vanity"

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type config struct {
	importTag *string
	sourceTag *string
	redir     Redirector
}

// Configures the Handler. The only required option is WithImport.
type Option func(*config)

// Instructs the go tool where to fetch the repo at vcsRoot and the importPath
// that tree should be rooted at.
func WithImport(importPath, vcs, vcsRoot string) Option {
	importTag := "meta name=\"go-import\" content=\"" + importPath + " " +
		vcs + " " + vcsRoot + "\">"
	return func(cfg *config) {
		if cfg.importTag != nil {
			panic(fmt.Sprintf("vanity: existing import tag: %s", *cfg.importTag))
		}
		cfg.importTag = &importTag
	}
}

// Instructs gddo (godoc.org) how to direct browsers to browsable source code
// for packages and their contents rooted at prefix.
//
// home specifies the home page of prefix, directory gives a format for how to
// browse a directory, and file gives a format for how to view a file and go to
// specific lines within it.
//
// More information can be found at https://github.com/golang/gddo/wiki/Source-Code-Links.
//
func WithSource(prefix, home, directory, file string) Option {
	sourceTag := "meta name=\"go-source\" content=\"" + prefix + " " +
		home + " " + directory + " " + file + "\">"
	return func(cfg *config) {
		if cfg.sourceTag != nil {
			panic(fmt.Sprintf("vanity: existing source tag: %s", *cfg.importTag))
		}
		cfg.sourceTag = &sourceTag
	}
}

// When a browser navigates to the vanity URL of pkg, this function rewrites
// pkg to a browsable URL.
type Redirector func(pkg string) (url string)

func WithRedirector(redir Redirector) Option {
	return func(cfg *config) {
		if cfg.redir != nil {
			panic("vanity: existing Redirector")
		}
		cfg.redir = redir
	}
}

// Returns an http.Handler that serves the vanity URL information for a single
// repository. Each Option gives additional information to agents about the
// repository or provides help to browsers that may have navigated to the vanity// URL. The WithImport Option is mandatory since the go tool requires it to
// fetch the repository.
func Handler(opts ...Option) http.Handler {
	var redir Redirector

	tpl := func() *template.Template {
		// Process options.
		var cfg config
		for _, opt := range opts {
			opt(&cfg)
		}

		// A WithImport is required.
		if cfg.importTag == nil {
			panic("vanity: WithImport is required")
		}

		tags := []string{*cfg.importTag}
		if cfg.sourceTag != nil {
			tags = append(tags, *cfg.sourceTag)
		}
		tagBlk := strings.Join(tags, "\n")

		h := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
%s
<meta http-equiv="refresh" content="0; url={{ . }}">
</head>
<body>
Nothing to see here; <a href="{{ . }}">move along</a>.
</body>
</html>
`, tagBlk)

		redir = cfg.redir
		return template.Must(template.New("").Parse(h))
	}()

	if redir == nil {
		redir = func(pkg string) string {
			return "https://godoc.org/" + pkg
		}
	}

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

		pkg := r.Host + r.URL.Path
		redirURL := redir(pkg)

		// Issue an HTTP redirect if this is definitely a browser.
		if r.FormValue("go-get") != "1" {
			http.Redirect(w, r, redirURL, http.StatusTemporaryRedirect)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=300")
		if err := tpl.ExecuteTemplate(w, "", redirURL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

// Helpers for common VCSs.

// Redirects gddo to browsable source files for GitHub hosted repositories.
func WithGitHubStyleSource(importPath, repoPath, ref string) Option {
	directory := repoPath + "/tree/" + ref + "{/dir}"
	file := repoPath + "/blob/" + ref + "{/dir}/{file}#L{line}"

	return WithSource(importPath, repoPath, directory, file)
}

// Redirects gddo to browsable source files for Gogs hosted repositories.
func WithGogsStyleSource(importPath, repoPath, ref string) Option {
	directory := repoPath + "/src/" + ref + "{/dir}"
	file := repoPath + "/src/" + ref + "{/dir}/{file}#L{line}"

	return WithSource(importPath, repoPath, directory, file)
}

// Creates a Handler that serves a GitHub repository at a specific importPath.
func GitHubHandler(importPath, user, repo, vcsScheme string) http.Handler {
	ghImportPath := "github.com/" + user + "/" + repo
	return Handler(
		WithImport(importPath, "git", vcsScheme+"://"+ghImportPath),
		WithGitHubStyleSource(importPath, "https://"+ghImportPath, "master"),
	)
}

// Creates a Handler that serves a repository hosted with Gogs at host at a
// specific importPath.
func GogsHandler(importPath, host, user, repo, vcsScheme string) http.Handler {
	gogsImportPath := host + "/" + user + "/" + repo
	return Handler(
		WithImport(importPath, "git", vcsScheme+"://"+gogsImportPath),
		WithGogsStyleSource(importPath, "https://"+gogsImportPath, "master"),
	)
}
