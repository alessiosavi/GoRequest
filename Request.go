package request

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

var Client http.Client
var err error

type Request struct {
	url       *url.URL
	timeout   time.Duration
	headers   http.Header
	cookieJar http.CookieJar
	cookies   []*http.Cookie
	req       *http.Request
}

// InitRequest is delegated to initialize a new empty request
func InitRequest(u string) *Request {
	var URL *url.URL
	if URL, err = url.Parse(u); err != nil {
		panic(err)
	}
	jar := initCookieJar()
	InitClient(false, false, false, 0)
	Client.Jar = jar
	return &Request{
		url:       URL,
		timeout:   time.Duration(0),
		headers:   http.Header{},
		cookieJar: jar,
		cookies:   nil,
		req:       &http.Request{},
	}
}

func initCookieJar() http.CookieJar {
	var jar http.CookieJar
	options := cookiejar.Options{PublicSuffixList: publicsuffix.List}
	if jar, err = cookiejar.New(&options); err != nil {
		panic(err)
	}
	return jar
}

func (r *Request) SetUrl(u string) error {
	if r.url, err = url.Parse(u); err != nil {
		return err
	}
	return nil
}

func (r *Request) SetTimeout(timeout time.Duration) {
	r.timeout = timeout
}

func (r *Request) AddHeader(key, value string) {
	r.headers.Add(key, value)
}

func (r *Request) AddCookie(cookie *http.Cookie) {
	r.cookies = append(r.cookies, cookie)
}

func (r *Request) SetClient(c http.Client) {
	Client = c
	c.Jar = r.cookieJar
}

func InitClient(disableKeepAlive, disableCompression, skipTls bool, timeout time.Duration) {
	var t *http.Transport
	var tlsConfig *tls.Config

	if skipTls {
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	} else {
		tlsConfig = &tls.Config{InsecureSkipVerify: false}
	}

	t = &http.Transport{
		Proxy:                  nil,
		DialContext:            nil,
		Dial:                   nil,
		DialTLS:                nil,
		TLSClientConfig:        tlsConfig,
		TLSHandshakeTimeout:    0,
		DisableKeepAlives:      disableKeepAlive,
		DisableCompression:     disableCompression,
		MaxIdleConns:           0,
		MaxIdleConnsPerHost:    0,
		MaxConnsPerHost:        0,
		IdleConnTimeout:        0,
		ResponseHeaderTimeout:  0,
		ExpectContinueTimeout:  0,
		TLSNextProto:           nil,
		ProxyConnectHeader:     nil,
		MaxResponseHeaderBytes: 0,
		WriteBufferSize:        0,
		ReadBufferSize:         0,
		ForceAttemptHTTP2:      false,
	}

	Client = http.Client{
		Transport:     t,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       timeout,
	}

}

func (r *Request) Post(contentType, body string) (*http.Response, error) {
	var err error
	var resp *http.Response
	if contentType == "" {
		contentType = "text/html; charset=UTF-8"
	}
	if r.req, err = http.NewRequest("POST", r.url.String(), bytes.NewBufferString(body)); err != nil {
		return nil, err
	}

	r.req.Header = r.headers
	for i := range r.cookies {
		r.req.AddCookie(r.cookies[i])
	}

	if resp, err = Client.Do(r.req); err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *Request) Get(params map[string]string) (*http.Response, error) {
	var err error
	var resp *http.Response
	var q url.Values

	if params != nil {
		if q, err = url.ParseQuery(r.url.RawQuery); err != nil {
			return nil, err
		}
		for key := range params {
			q.Add(key, params[key])
		}
		r.url.RawQuery = q.Encode()
	}

	if r.req, err = http.NewRequest("GET", r.url.String(), nil); err != nil {
		return nil, err
	}

	r.req.Header = r.headers
	for i := range r.cookies {
		r.req.AddCookie(r.cookies[i])
	}

	if resp, err = Client.Do(r.req); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetBody is delegated to retrieve the body from the given response
func GetBody(resp *http.Response) (string, error) {
	var sb strings.Builder

	defer resp.Body.Close()
	if _, err = io.Copy(&sb, resp.Body); err != nil {
		return "", nil
	}
	return sb.String(), nil
}

func (r *Request) SetBasicAuth(username, password string) {
	r.AddHeader("Authorization", "Basic "+basicAuth(username, password))
}

func (r *Request) SetBearerAuth(token string) {
	r.AddHeader("Authorization", "Bearer "+token)
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
