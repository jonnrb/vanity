package vanity // import "go.jonnrb.io/vanity"

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type tag string

func ImportTag(importPath, vcs, vcsRoot string) tag {
	return tag("<meta name=\"go-import\" content=\"" + importPath + " " + vcs +
		" " + vcsRoot + "\">")
}

func SourceTag(prefix, home, directory, file string) tag {
	return tag("<meta name=\"go-source\" content=\"" + prefix + " " + home +
		" " + directory + " " + file + "\">")
}

// Returns an http.Handler that serves the vanity URL information for a single
// repository. Each tag gives additional information to agents about the
// repository and the packages it contains. An ImportTag is basically mandatory
// since the go tool requires it to fetch the repository.
func Handler(tags ...tag) http.Handler {
	tpl := func() *template.Template {
		s := make([]string, len(tags))
		for i, t := range tags {
			s[i] = string(t)
		}
		tagBlk := strings.Join(s, "\n")

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

		return template.Must(template.New("").Parse(h))
	}()

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

		// Redirect browsers to gddo.
		if r.FormValue("go-get") != "1" {
			url := "https://godoc.org/" + r.Host + r.URL.Path
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=300")
		if err := tpl.ExecuteTemplate(w, "", pkg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func GitHubStyleSourceTag(importPath, repoPath, ref string) tag {
	directory := repoPath + "/tree/" + ref + "{/dir}"
	file := repoPath + "/blob/" + ref + "{/dir}/{file}#L{line}"

	return SourceTag(importPath, repoPath, directory, file)
}

func GogsStyleSourceTag(importPath, repoPath, ref string) tag {
	directory := repoPath + "/src/" + ref + "{/dir}"
	file := repoPath + "/src/" + ref + "{/dir}/{file}#L{line}"

	return SourceTag(importPath, repoPath, directory, file)
}

// Creates a Handler that serves a GitHub repository at a specific importPath.
func GitHubHandler(importPath, user, repo, vcsScheme string) http.Handler {
	ghImportPath := "github.com/" + user + "/" + repo
	return Handler(
		ImportTag(importPath, "git", vcsScheme+"://"+ghImportPath),
		GitHubStyleSourceTag(importPath, "https://"+ghImportPath, "master"),
	)
}

// Creates a Handler that serves a repository hosted with Gogs at host at a
// specific importPath.
func GogsHandler(importPath, host, user, repo, vcsScheme string) http.Handler {
	gogsImportPath := host + "/" + user + "/" + repo
	return Handler(
		ImportTag(importPath, "git", vcsScheme+"://"+gogsImportPath),
		GogsStyleSourceTag(importPath, "https://"+gogsImportPath, "master"),
	)
}
