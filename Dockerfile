from quay.io/jonnrb/go as build
add . /src
run cd /src && CGO_ENABLED=0 go get ./cmd/vanityserver

from gcr.io/distroless/static
expose 8080
copy --from=build /go/bin/vanityserver /vanityserver
entrypoint ["/vanityserver"]
