# Vanity [![GoDoc](https://godoc.org/go.jonnrb.io/vanity?status.svg)](https://godoc.org/go.jonnrb.io/vanity)

A vanity import path is any import path that can be downloaded with
`go get` but isn't otherwise blessed by the `go` tool (e.g. GitHub,
BitBucket, etc.). A commonly used vanity import path is
"golang.org/x/...". This package attempts to mimic the behavior of
"golang.org/x/..." as closely as possible.

## Features
 - Redirects browsers to godoc.org
 - Redirects Go tool to VCS
 - Redirects godoc.org to browsable files
 - Redirects HTTP to HTTPS

## Installation
```bash
go get go.jonnrb.io/vanity
```

## Specification
- [Remote Import Paths](https://golang.org/cmd/go/#hdr-Remote_import_paths)
- [GDDO Source Code Links](https://github.com/golang/gddo/wiki/Source-Code-Links)
- [Custom Import Path Checking](https://docs.google.com/document/d/1jVFkZTcYbNLaTxXD9OcGfn7vYv5hWtPx9--lTx1gPMs/edit)
