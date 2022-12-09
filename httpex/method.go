package ghttpex

type Method string

// Common HTTP methods.
//
// Unless otherwise noted, these are defined in RFC 7231 section 4.3.
const (
	GET     = Method("GET")
	HEAD    = Method("HEAD")
	POST    = Method("POST")
	PUT     = Method("PUT")
	PATCH   = Method("PATCH") // RFC 5789
	DELETE  = Method("DELETE")
	CONNECT = Method("CONNECT")
	OPTIONS = Method("OPTIONS")
	TRACE   = Method("TRACE")
)
