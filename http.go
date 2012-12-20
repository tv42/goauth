package oauth

import (
	"bufio"
	//"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"

	"net/url"
	"strings"
)

type badStringError struct {
	what string
	str  string
}

func (e *badStringError) Error() string {
	return fmt.Sprintf("%s %q", e.what, e.str)
}

type readClose struct {
	io.Reader
	io.Closer
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func hasPort(s string) bool {
	return strings.LastIndex(s, ":") > strings.LastIndex(s, "]")
}

func send(req *http.Request) (resp *http.Response, err error) {
	//dump, _ := http.DumpRequest(req, true)
	//fmt.Fprintf(os.Stderr, "%s", dump)
	//fmt.Fprintf(os.Stderr, "\n--- body:\n%s\n---", bodyString(req.Body))
	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		return nil, &badStringError{"unsupported protocol scheme", req.URL.Scheme}
	}

	addr := req.URL.Host
	var conn net.Conn
	switch req.URL.Scheme {
	case "http":
		if !hasPort(addr) {
			addr += ":http"
		}

		conn, err = net.Dial("tcp", addr)
	case "https":
		if !hasPort(addr) {
			addr += ":https"
		}

		conn, err = tls.Dial("tcp", addr, nil)
	}
	if err != nil {
		return nil, err
	}

	err = req.Write(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	reader := bufio.NewReader(conn)
	resp, err = http.ReadResponse(reader, req)
	if err != nil {
		conn.Close()
		return nil, err
	}

	resp.Body = readClose{resp.Body, conn}

	return
}

func post(url_ string, oauthHeaders map[string]string) (r *http.Response, err error) {
	var req http.Request
	req.Method = "POST"
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.Close = true
	req.Header = map[string][]string{
		"Authorization": {"OAuth "},
	}
	req.TransferEncoding = []string{"chunked"}

	first := true
	for k, v := range oauthHeaders {
		if first {
			first = false
		} else {
			req.Header["Authorization"][0] += ",\n    "
		}
		req.Header["Authorization"][0] += k + "=\"" + v + "\""
	}

	req.URL, err = url.Parse(url_)
	if err != nil {
		return nil, err
	}

	return send(&req)
}

func get(url_ string, oauthHeaders map[string]string) (r *http.Response, err error) {
	var req http.Request
	req.Method = "GET"
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.Close = true
	req.Header = map[string][]string{
		"Authorization": {"OAuth "},
	}
	req.TransferEncoding = []string{"chunked"}

	first := true
	for k, v := range oauthHeaders {
		if first {
			first = false
		} else {
			req.Header["Authorization"][0] += ",\n    "
		}
		req.Header["Authorization"][0] += k + "=\"" + v + "\""
	}

	req.URL, err = url.Parse(url_)
	if err != nil {
		return nil, err
	}

	return send(&req)
}


func (o *OAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	if !o.Authorized() {
		return nil, &danceError{
			What:  "Not authorized",
			Where: "OAuth\u00b7PostParams()",
		}
	}

	o.UpdateRequest(req)
 	oParams := o.params()
 	r, err = o.makeRequest("POST", url, oParams, params)

 	t := o.Transport
	if t == nil {
		t = http.DefaultTransport
	}

}
