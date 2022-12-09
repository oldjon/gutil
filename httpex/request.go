package ghttpex

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

// Request provides more useful methods for requesting one url than http.Request.
type Request struct {
	*http.Request

	client *Client

	url       string
	command   string
	paramKeys []string
	params    map[string][]string
	files     map[string]string
	dump      []byte
	resp      *http.Response
	respBody  []byte
	reqBody   []byte
	signFunc  func(*Request) error
}

// Command sets command in request manually
func (r *Request) Command(command string) *Request {
	r.command = command
	return r
}

// Header sets header item string in request
func (r *Request) Header(key, value string) *Request {
	r.Request.Header.Set(key, value)
	return r
}

// ContentType sets Content-Type header field
func (r *Request) ContentType(value string) *Request {
	r.Request.Header.Set("Content-Type", value)
	return r
}

// BasicAuth sets the request's Authorization header to use HTTP Basic Authentication with the provided username and password.
func (r *Request) BasicAuth(username, password string) *Request {
	r.Request.SetBasicAuth(username, password)
	return r
}

// BearerAuth sets the request's Authorization header to use bearer token
func (r *Request) BearerAuth(token string) *Request {
	r.Request.Header.Set("Authorization", "Bearer "+token)
	return r
}

// Param adds query param in to request.
// params build query string as ?key1=value1&key2=value2...
func (r *Request) Param(key, value string) *Request {
	if param, ok := r.params[key]; ok {
		r.params[key] = append(param, value)
	} else {
		r.paramKeys = append(r.paramKeys, key)
		r.params[key] = []string{value}
	}
	return r
}

// PostFile adds a post file to the request
func (r *Request) PostFile(formname, filename string) *Request {
	r.files[formname] = filename
	return r
}

// SignFunc adds a sign function to the request (can modify Request before sent such as set header)
func (r *Request) SignFunc(signFunc func(*Request) error) *Request {
	r.signFunc = signFunc
	return r
}

// BytesBody adds request raw body by []byte
func (r *Request) BytesBody(data []byte) *Request {
	r.reqBody = data
	return r
}

// StringBody adds request raw body by string
func (r *Request) StringBody(data string) *Request {
	r.reqBody = []byte(data)
	return r
}

// JSONBody adds request raw body encoding by JSON
func (r *Request) JSONBody(obj interface{}) (*Request, error) {
	if obj == nil {
		return r, nil
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return r, err
	}
	r.reqBody = data
	return r, nil
}

// SetBody adds request body independently
// copied from net/http/request.go#NewRequestWithContext
func (r *Request) SetBody(body io.Reader) *Request {
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}
	r.Body = rc

	switch v := body.(type) {
	case *bytes.Buffer:
		r.ContentLength = int64(v.Len())
		buf := v.Bytes()
		r.GetBody = func() (io.ReadCloser, error) {
			r := bytes.NewReader(buf)
			return ioutil.NopCloser(r), nil
		}
	case *bytes.Reader:
		r.ContentLength = int64(v.Len())
		snapshot := *v
		r.GetBody = func() (io.ReadCloser, error) {
			r := snapshot
			return ioutil.NopCloser(&r), nil
		}
	case *strings.Reader:
		r.ContentLength = int64(v.Len())
		snapshot := *v
		r.GetBody = func() (io.ReadCloser, error) {
			r := snapshot
			return ioutil.NopCloser(&r), nil
		}
	default:
		// This is where we'd set it to -1 (at least
		// if body != NoBody) to mean unknown, but
		// that broke people during the Go 1.8 testing
		// period. People depend on it being 0 I
		// guess. Maybe retry later. See Issue 18117.
	}
	// For client requests, Request.ContentLength of 0
	// means either actually 0, or unknown. The only way
	// to explicitly say that the ContentLength is zero is
	// to set the Body to nil. But turns out too much code
	// depends on NewRequest returning a non-nil Body,
	// so we use a well-known ReadCloser variable instead
	// and have the http package also treat that sentinel
	// variable to mean explicitly zero.
	if r.GetBody != nil && r.ContentLength == 0 {
		r.Body = http.NoBody
		r.GetBody = func() (io.ReadCloser, error) { return http.NoBody, nil }
	}

	return r
}

func (r *Request) buildURL() error {
	var paramBody string
	if len(r.paramKeys) > 0 {
		var buf bytes.Buffer
		for _, k := range r.paramKeys {
			vs := r.params[k]
			keyEscaped := url.QueryEscape(k)
			for _, v := range vs {
				buf.WriteString(keyEscaped)
				buf.WriteByte('=')
				buf.WriteString(url.QueryEscape(v))
				buf.WriteByte('&')
			}
		}
		paramBody = buf.String()
		paramBody = paramBody[0 : len(paramBody)-1]
	}

	// build GET url with query string
	if len(paramBody) > 0 && r.GetMethod() == GET {
		if strings.Contains(r.url, "?") {
			r.url += "&" + paramBody
		} else {
			r.url += "?" + paramBody
		}

		urlParsed, err := url.Parse(r.url)
		if err != nil {
			r.client.logger.Println("Httpex parse url err:", err, "url:", r.url)
			return err
		}
		r.Request.URL = urlParsed
		return nil
	}

	// build POST/PUT/PATCH/DELETE url and body
	if (r.GetMethod() == POST || r.GetMethod() == PUT || r.GetMethod() == PATCH || r.GetMethod() == DELETE) && r.Request.Body == nil {
		// with files
		if len(r.files) > 0 {
			pr, pw := io.Pipe()
			bodyWriter := multipart.NewWriter(pw)
			go func() {
				for formname, filename := range r.files {
					fileWriter, err := bodyWriter.CreateFormFile(formname, filename)
					if err != nil {
						r.client.logger.Println("Httpex err:", err, "url:", r.url)
					}
					fh, err := os.Open(filename)
					if err != nil {
						r.client.logger.Println("Httpex err:", err, "url:", r.url)
					}
					// iocopy
					_, err = io.Copy(fileWriter, fh)
					fh.Close()
					if err != nil {
						r.client.logger.Println("Httpex err:", err, "url:", r.url)
					}
				}
				for k, v := range r.params {
					for _, vv := range v {
						if err := bodyWriter.WriteField(k, vv); err != nil {
							r.client.logger.Println("Httpex err:", err, "url:", r.url)
						}
					}
				}
				bodyWriter.Close()
				pw.Close()
			}()
			r.Request.Header.Set("Content-Type", bodyWriter.FormDataContentType())
			r.Request.Body = ioutil.NopCloser(pr)
			r.Request.Header.Set("Transfer-Encoding", "chunked")
			return nil
		}

		// with params
		if len(paramBody) > 0 {
			r.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r.reqBody = []byte(paramBody)
		}
	}

	return nil
}

// DoRequest executes client.Do with extra options
func (r *Request) DoRequest(ctx context.Context) (*http.Response, error) {
	if ctx == nil {
		ctx = context.TODO()
	}
	if r.client == nil {
		r.client = DefaultClient
	}

	root := (*Request).doRequest

	fcs := r.client.filterChains
	if len(fcs) > 0 {
		for i := len(fcs) - 1; i >= 0; i-- {
			root = fcs[i](root)
		}
	}

	r.Request = r.Request.WithContext(ctx)
	urlParsed, err := url.Parse(r.url)
	if err != nil {
		r.client.logger.Println("Httpex parse url err:", err, "url:", r.url)
		return nil, err
	}
	r.Request.URL = urlParsed
	if r.command == "" {
		r.command = strings.TrimPrefix(urlParsed.Path, "/")
	}

	return root(r)
}

func (r *Request) doRequest() (resp *http.Response, err error) {
	if err = r.buildURL(); err != nil {
		return
	}
	c := r.client

	if c.userAgent != "" && r.Request.Header.Get("User-Agent") == "" {
		r.Request.Header.Set("User-Agent", c.userAgent)
	}

	if c.showDebug {
		dump, err := httputil.DumpRequest(r.Request, c.dumpBody)
		if err != nil {
			c.logger.Println("Httpex dump request err:", err, "url:", r.url)
		}
		r.dump = dump
	}

	// sign function
	if r.signFunc != nil {
		if err = r.signFunc(r); err != nil {
			c.logger.Println("Httpex call signFunc err:", err, "url:", r.url)
			return
		}
	}

	// build body
	if r.Request.Body == nil && r.reqBody != nil {
		r.Request.Body = ioutil.NopCloser(bytes.NewReader(r.reqBody))
		r.Request.ContentLength = int64(len(r.reqBody))
		r.Request.GetBody = func() (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewReader(r.reqBody)), nil
		}
	}

	// retries default value is 0, it will run once.
	// retries equal to -1, it will run forever until success
	// retries is set, it will retries fixed times.
	// Sleeps for a 400ms between calls to reduce spam
	for i := 0; c.retries == -1 || i <= c.retries; i++ {
		if i > 0 {
			c.logger.Println("Httpex doRequest err:", err, "url:", r.url, "ready to retry:", i)
			time.Sleep(c.retryDelay)
			// rewind body
			if r.Request.Body != nil {
				if r.Request.GetBody == nil {
					c.logger.Println("Httpex doRequest retry failed (nil Request.GetBody). url:", r.url)
					break
				}
				if r.Request.Body, err = r.Request.GetBody(); err != nil {
					c.logger.Println("Httpex doRequest retry failed. err:", err, "url:", r.url)
					break
				}
			}
		}
		resp, err = c.Client.Do(r.Request)
		if err == nil {
			break
		}
	}
	return
}

// Response executes request client gets response manually.
// The result (*http.Response) will be saved to the instance.
func (r *Request) Response(ctx context.Context) (*http.Response, error) {
	return r.getResponse(ctx)
}

func (r *Request) getResponse(ctx context.Context) (*http.Response, error) {
	if r.resp != nil {
		return r.resp, nil
	}
	resp, err := r.DoRequest(ctx)
	if err != nil {
		r.client.logger.Println("Httpex getResponse err:", err, "url:", r.url)
		return nil, err
	}
	r.resp = resp
	return resp, nil
}

// StatusCode gets response's status code (should be called after Response)
func (r *Request) StatusCode() int {
	if r.resp == nil {
		return 0
	}
	return r.resp.StatusCode
}

// ToString returns the body string in response.
// it calls getResponse inner.
func (r *Request) String(ctx context.Context) (string, error) {
	data, err := r.Bytes(ctx)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Bytes returns the body []byte in response.
// it calls getResponse inner.
func (r *Request) Bytes(ctx context.Context) ([]byte, error) {
	if r.respBody != nil {
		return r.respBody, nil
	}
	resp, err := r.getResponse(ctx)
	if err != nil {
		return nil, err
	}
	if resp.Body == nil {
		return nil, nil
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			r.client.logger.Println("Httpex response body close failed. err:", err, "url:", r.url)
		}
	}()
	r.respBody, err = ioutil.ReadAll(resp.Body)
	return r.respBody, err
}

// ToJSON returns the map that marshals from the body bytes as json in response.
// it calls getResponse inner.
func (r *Request) ToJSON(ctx context.Context, v interface{}) error {
	data, err := r.Bytes(ctx)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// GetMethod gets method in request
func (r *Request) GetMethod() Method {
	return Method(r.Request.Method)
}

// GetURL gets url in request
func (r *Request) GetURL() string {
	return r.url
}

// GetCommand gets command in request
func (r *Request) GetCommand() string {
	return r.command
}

// GetHeader gets header in request
func (r *Request) GetHeader() http.Header {
	return r.Request.Header
}

// GetParam gets the first param value in request by key
func (r *Request) GetParam(key string) string {
	vs := r.params[key]
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

// GetParams gets params in request
func (r *Request) GetParams() map[string][]string {
	return r.params
}

// GetFile gets file in request by formname
func (r *Request) GetFile(formname string) string {
	return r.files[formname]
}

// GetFiles gets files in request
func (r *Request) GetFiles() map[string]string {
	return r.files
}

// GetReqBody returns the request body
func (r *Request) GetReqBody() []byte {
	return r.reqBody
}

// DumpRequest returns the dump request
func (r *Request) DumpRequest() []byte {
	return r.dump
}
