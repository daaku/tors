// Package tors provides torrent search functionality.
package tors

import (
	"context"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

type errNoResults struct{}

func (errNoResults) Error() string {
	return "tors: no results found"
}

// IsErrNoResults returns true if the error indicates that no results were
// found.
func IsErrNoResults(err error) bool {
	_, ok := errors.Cause(err).(errNoResults)
	return ok
}

// Client performs searches against the configured URLs.
type Client struct {
	transport http.RoundTripper
	urls      []func(string) *url.URL
}

// Search performs a search with the given query and returns a single magnet
// URL.
func (c *Client) Search(ctx context.Context, q string) (string, error) {
	if q == "" {
		return "", errors.New("tors: invalid empty query search")
	}

	for _, uf := range c.urls {
		req := (&http.Request{
			Method: "GET",
			URL:    uf(q),
			Header: http.Header{},
		}).WithContext(ctx)
		res, err := c.transport.RoundTrip(req)
		if err != nil {
			return "", errors.Wrap(err, "http error")
		}
		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return "", errors.Wrap(err, "html parsing error")
		}

		uri, found := doc.Find("a[href*=magnet]").First().Attr("href")
		if found {
			return uri, nil
		}
	}

	return "", errors.Wrap(errNoResults{}, "")
}

// ClientOption allows configuring various aspects of the Client.
type ClientOption func(*Client)

// ClientTransport configures the Transport for the Client. If not specified
// http.DefaultTransport is used.
func ClientTransport(t http.RoundTripper) ClientOption {
	return func(c *Client) {
		c.transport = t
	}
}

// ClientURL configures an additional URL generator for a query.
func ClientURL(f func(query string) *url.URL) ClientOption {
	return func(c *Client) {
		c.urls = append(c.urls, f)
	}
}

// NewClient creates a new client with the given options.
func NewClient(options ...ClientOption) (*Client, error) {
	var c Client
	for _, o := range options {
		o(&c)
	}
	if len(c.urls) == 0 {
		return nil, errors.New("tors: no URLs configured")
	}
	if c.transport == nil {
		c.transport = http.DefaultTransport
	}
	return &c, nil
}
