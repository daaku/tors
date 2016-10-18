package tors

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/facebookgo/ensure"
)

type fTransport func(*http.Request) (*http.Response, error)

func (f fTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

var dummyURL = ClientURL(func(string) *url.URL { return &url.URL{} })

func TestNoResults(t *testing.T) {
	c, err := NewClient(
		dummyURL,
		ClientTransport(fTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				Body: ioutil.NopCloser(strings.NewReader("<html>")),
			}, nil
		})),
	)
	ensure.Nil(t, err)
	uri, err := c.Search("unimportant")
	ensure.True(t, IsErrNoResults(err), err)
	ensure.Err(t, err, regexp.MustCompile("no results found"))
	ensure.DeepEqual(t, uri, "")
}

func TestTransportError(t *testing.T) {
	givenErr := errors.New("foobar")
	c, err := NewClient(
		dummyURL,
		ClientTransport(fTransport(func(*http.Request) (*http.Response, error) {
			return nil, givenErr
		})),
	)
	ensure.Nil(t, err)
	uri, err := c.Search("unimportant")
	ensure.Err(t, err, regexp.MustCompile(givenErr.Error()))
	ensure.DeepEqual(t, uri, "")
}

func TestBodyReadError(t *testing.T) {
	givenErr := errors.New("foobar")
	r, w := io.Pipe()
	w.CloseWithError(givenErr)
	c, err := NewClient(
		dummyURL,
		ClientTransport(fTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{Body: ioutil.NopCloser(r)}, nil
		})),
	)
	ensure.Nil(t, err)
	uri, err := c.Search("unimportant")
	ensure.Err(t, err, regexp.MustCompile(givenErr.Error()))
	ensure.DeepEqual(t, uri, "")
}

func TestNormalResults(t *testing.T) {
	c, err := NewClient(
		dummyURL,
		ClientTransport(fTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(`
				<body>
					<table>
						<tr>
							<td>
								<a href="magnet:a">l</a>
							</td>
						</tr>
					</table>
				</body>
			`)),
			}, nil
		})),
	)
	ensure.Nil(t, err)
	uri, err := c.Search("unimportant")
	ensure.Nil(t, err)
	ensure.DeepEqual(t, uri, "magnet:a")
}

func TestNewClientNoURLsError(t *testing.T) {
	c, err := NewClient()
	ensure.True(t, c == nil)
	ensure.Err(t, err, regexp.MustCompile("no URLs configured"))
}

func TestEmptyQuerySearch(t *testing.T) {
	c, err := NewClient(dummyURL)
	ensure.Nil(t, err)
	uri, err := c.Search("")
	ensure.Err(t, err, regexp.MustCompile("empty query search"))
	ensure.DeepEqual(t, uri, "")
}
