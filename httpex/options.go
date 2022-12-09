package ghttpex

import (
	"log"
	"os"
	"time"
)

const (
	defaultTimeout    = 30 * time.Second
	defaultRetryDelay = 400 * time.Millisecond
)

// Options includes useful client setting options
type Options struct {
	Logger       Logger
	Timeout      time.Duration // timeout for http.Client
	Retries      int           // 0 means no retried; -1 means retried forever; others means retried times.
	RetryDelay   time.Duration // the delay between each request when set Retries
	FilterChains []FilterChain // set ordered filters to intercept each http request to do some custom things (like AOP)
	UserAgent    string        // default value for http header "User-Agent"
	ShowDebug    bool          // can get dump info by Request.DumpRequest after send request if set true.
	DumpBody     bool          // set whether need to dump the request body
}

// init add some default values for Options
func (opt *Options) init() {
	if opt.Logger == nil {
		opt.Logger = log.New(os.Stderr, "httpex: ", log.LstdFlags|log.Lshortfile)
	}
	if opt.Timeout <= 0 {
		opt.Timeout = defaultTimeout
	}
	if opt.Retries < -1 {
		opt.Retries = -1
	}
	if opt.Retries != 0 && opt.RetryDelay <= 0 {
		opt.RetryDelay = defaultRetryDelay
	}
}
