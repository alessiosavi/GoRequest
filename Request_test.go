package request

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_Cookie(t *testing.T) {

	// create a listener with the desired port.
	var l net.Listener
	var err error
	var url string
	var resp *http.Response
	var body string
	var first = "First Call!"
	var second = "Second Call!"

	url = "127.0.0.1:8082"
	if l, err = net.Listen("tcp", url); err != nil {
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

	r := InitRequest("http://" + url)
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

	ts.CloseClientConnections()
	ts.Close()

}
