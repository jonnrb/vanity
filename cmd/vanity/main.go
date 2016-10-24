package main // import "kkn.fi/vanity/cmd/vanity"

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"kkn.fi/vanity"
)

var (
	domainFlag = flag.String("d", "", "http domain name")
	portFlag   = flag.Int("p", 80, "http server port")
	confFlag   = flag.String("c", "", "configuration file")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: vanity -d domain -c vanity.conf [-p 80]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	log.SetPrefix("vanity: ")
	log.SetFlags(0)

	if *domainFlag == "" || *confFlag == "" {
		usage()
	}

	c, err := os.Open(*confFlag)
	if err != nil {
		log.Fatal(err)
	}
	conf, err := readConfig(c)
	if err != nil {
		log.Fatal(err)
	}
	server := vanity.NewServer(*domainFlag, conf)
	port := fmt.Sprintf(":%v", *portFlag)
	log.Fatal(http.ListenAndServe(port, server))
}

func readConfig(r io.Reader) (map[vanity.Path]vanity.Package, error) {
	conf := make(map[vanity.Path]vanity.Package, 0)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		switch len(fields) {
		case 0:
			continue
		case 3:
			pack := vanity.NewPackage(parsePath(fields[0]), fields[1], fields[2])
			conf[vanity.Path(fields[0])] = *pack
		default:
			return conf, errors.New("configuration error: " + scanner.Text())
		}
	}
	return conf, nil
}

func parsePath(p string) string {
	c := strings.Index(p[1:], "/")
	if c == -1 {
		return p
	}
	return p[:c+1]
}
