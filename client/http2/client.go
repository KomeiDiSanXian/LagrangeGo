package http2

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/http2"

	"github.com/fumiama/terasu/dns"
	trsh2 "github.com/fumiama/terasu/http2"
)

func init() {
	dns.SetTimeout(time.Second)
}

var (
	ErrEmptyHostAddress = errors.New("empty host addr")
)

var defaultDialer = net.Dialer{
	Timeout: time.Minute,
}

func SetDefaultClientTimeout(t time.Duration) {
	defaultDialer.Timeout = t
}

var DefaultClient = http.Client{
	Transport: &http2.Transport{},
}

var TRSClient = &trsh2.DefaultClient

func Get(url string) (resp *http.Response, err error) {
	return DefaultClient.Get(url)
}

func Head(url string) (resp *http.Response, err error) {
	return DefaultClient.Head(url)
}

func Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	return DefaultClient.Post(url, contentType, body)
}

func PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return DefaultClient.PostForm(url, data)
}
