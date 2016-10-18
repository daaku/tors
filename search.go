// +build ignore

package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/daaku/tors"
	"github.com/pkg/errors"
)

func linuxTrackerOrg(q string) *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   "linuxtracker.org",
		Path:   "/index.php",
		RawQuery: (url.Values{
			"page":   []string{"torrents"},
			"search": []string{q},
		}).Encode(),
	}
}

func Main() error {
	query := flag.String("query", "", "query string to search for")
	flag.Parse()

	client, err := tors.NewClient(tors.ClientURL(linuxTrackerOrg))
	if err != nil {
		return errors.Wrap(err, "invalid client")
	}
	uri, err := client.Search(*query)
	if err != nil {
		return errors.Wrap(err, "search failed")
	}

	fmt.Println(uri)
	return nil
}

func main() {
	if err := Main(); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}
