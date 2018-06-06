# vanityserver

Runs a barebones vanity server over HTTP.

## Usage

```
./vanityserver fqdn [repo file]
```

If repo file is not given, "./repos" is used. The file has the following format:

```
pkgroot  vcsScheme://vcsHost/user/repo
pkgroot2 vcsScheme://vcsHost/user/repo2
```

vcsHost is either a Gogs server (that's what I use) or github.com. I'm open to
supporting other VCSs but I'm not sure what that would look like.
