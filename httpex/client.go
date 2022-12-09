package ghttpex

import (
	"net/http"
	"time"
)

// DefaultClient client with default options
var DefaultClient = NewClient(&Options{})

// Client provides more useful features than http.Client
type Client struct {
	*http.Client

	logger       Logger
	retries      int
	retryDelay   time.Duration
	filterChains []FilterChain
	userAgent    string
	showDebug    bool
	dumpBody     bool
}

// NewClient creates an httpex.Client object
func NewClient(opt *Options) *Client {
	opt.init()
	c := Client{
		Client: &http.Client{
			Timeout: opt.Timeout,
		},
		logger:       opt.Logger,
		retries:      opt.Retries,
		retryDelay:   opt.RetryDelay,
		filterChains: opt.FilterChains,
		userAgent:    opt.UserAgent,
		showDebug:    opt.ShowDebug,
		dumpBody:     opt.DumpBody,
	}
	return &c
}

// NewRequest returns *Request with specific method based on the Client
func (c *Client) NewRequest(method Method, rawurl string) *Request {
	req := http.Request{
		Method:     string(method),
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	return &Request{
		Request: &req,
		client:  c,
		url:     rawurl,
		params:  make(map[string][]string),
		files:   make(map[string]string),
	}
}

// Get returns *Request with GET method based on DefaultClient
func Get(url string) *Request {
	return DefaultClient.Get(url)
}

// Get returns *Request with GET method.
func (c *Client) Get(url string) *Request {
	return c.NewRequest(GET, url)
}

// Post returns *Request with POST method based on DefaultClient
func Post(url string) *Request {
	return DefaultClient.Post(url)
}

// Post returns *Request with POST method.
func (c *Client) Post(url string) *Request {
	return c.NewRequest(POST, url)
}

// Put returns *Request with PUT method based on DefaultClient
func Put(url string) *Request {
	return DefaultClient.Put(url)
}

// Put returns *Request with PUT method.
func (c *Client) Put(url string) *Request {
	return c.NewRequest(PUT, url)
}

// Delete returns *Request with DELETE method based on DefaultClient
func Delete(url string) *Request {
	return DefaultClient.Delete(url)
}

// Delete returns *Request with DELETE method.
func (c *Client) Delete(url string) *Request {
	return c.NewRequest(DELETE, url)
}

// Head returns *Request with HEAD method based on DefaultClient
func Head(url string) *Request {
	return DefaultClient.Head(url)
}

// Head returns *Request with HEAD method.
func (c *Client) Head(url string) *Request {
	return c.NewRequest(HEAD, url)
}
