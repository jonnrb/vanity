from quay.io/jonnrb/go as build
add . /src
run cd /src && go get -v ./cmd/vanityserver

from gcr.io/distroless/base
expose 8080
copy --from=build /go/bin/vanityserver /vanityserver
entrypoint ["/vanityserver"]
