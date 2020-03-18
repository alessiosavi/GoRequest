package request

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// Test that a cookie is saved for consecutive request
func Test_Cookie(t *testing.T) {

	// create a listener with the desired port.
	var (
		resp   *http.Response
		URL    = "127.0.0.1:8082"
		first  = "First Call!"
		second = "Second Call!"
		body   string
		err    error
		l      net.Listener
	)

	if l, err = net.Listen("tcp", URL); err != nil {
		panic(err)
	}

	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err2 := r.Cookie("myCustomCookie")
		if err2 != nil && err2 != http.ErrNoCookie {
			t.Error(err)
		}
		if cookie != nil && cookie.Value == "1" {
			w.Header().Add("Set-Cookie", "mySecondCookie=1")
			_, _ = fmt.Fprintf(w, second)
		} else {
			w.Header().Add("Set-Cookie", "myCustomCookie=1")
			_, _ = fmt.Fprintf(w, first)
		}

	})
	ts := httptest.NewUnstartedServer(f)
	_ = ts.Listener.Close()
	ts.Listener = l

	// Start the server.
	ts.Start()
	time.Sleep(10 * time.Millisecond)

	r := InitRequest("http://" + URL)
	if resp, err = r.Post("", ""); err != nil {
		t.Error(err)
	}
	if body, err = GetBody(resp); err != nil {
		t.Error(err)
	}
	if body != first {
		t.Errorf("Expected: %s | Found: %s", first, body)
	}

	if resp, err = r.Post("", ""); err != nil {
		t.Error(err)
	}
	if body, err = GetBody(resp); err != nil {
		t.Error(err)
	}
	if body != second {
		t.Errorf("Expected: %s | Found: %s", second, body)
	}

	var _url *url.URL
	if _url, err = url.Parse("http://" + URL); err != nil {
		panic(err)
	}
	t.Logf("Cookies: %+v", r.cookies)
	t.Logf("Cookies: %+v", r.cookieJar)
	t.Logf("Cookies: %+v", client.Jar.Cookies(_url))
	ts.CloseClientConnections()
	ts.Close()

}
