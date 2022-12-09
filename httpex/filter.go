package ghttpex

import (
	"net/http"
)

type FilterChain func(next Filter) Filter

type Filter func(req *Request) (*http.Response, error)
