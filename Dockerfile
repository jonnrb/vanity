from golang:1.10.3 as build
add . /go/src/go.jonnrb.io/vanity
workdir /go/src/go.jonnrb.io/vanity
run go install ./cmd/vanityserver

from gcr.io/distroless/base
expose 8080
copy --from=build /go/bin/vanityserver /vanityserver
entrypoint ["/vanityserver"]
