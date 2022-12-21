package webfinger

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// setup a local HTTP server for testing
func setup() (client *Client, mux *http.ServeMux, host string, teardown func()) {
	// test server
	mux = http.NewServeMux()
	server := httptest.NewTLSServer(mux)
	u, _ := url.Parse(server.URL)

	// for testing, use an HTTP client which doesn't check certs
	client = NewClient(&http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})
	return client, mux, u.Host, server.Close
}

func TestNewClient_EmptyClient(t *testing.T) {
	client := NewClient(nil)
	if client.client != http.DefaultClient {
		t.Errorf("NewClient(nil) did not use http.DefaultClient")
	}
}

func TestResource_Parse(t *testing.T) {
	tests := []struct {
		input string
		want  *Resource
	}{
		// URL with host
		{"http://example.com/", &Resource{Scheme: "http", Host: "example.com", Path: "/"}},
		// email-like identifier
		{"bob@example.com", &Resource{Scheme: "acct", Opaque: "bob@example.com"}},
	}

	for _, tt := range tests {
		got, err := Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", tt.input, err)
		}
		if !cmp.Equal(got, tt.want) {
			t.Errorf("Parse(%q) returned %#v, want %#v", tt.input, got, tt.want)
		}
	}
}

func TestResource_Parse_error(t *testing.T) {
	_, err := Parse("example.com")
	if err == nil {
		t.Error("Expected parse error")
	}

	_, err = Parse("%")
	if err == nil {
		t.Error("Expected parse error")
	}
}

func TestResource_WebFingerHost(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// URL with host
		{"http://example.com/", "example.com"},
		// emai-like identifier
		{"bob@example.com", "example.com"},
		// mailto URL
		{"mailto:bob@example.com", "example.com"},
		// URL with no host
		{"file:///example", ""},
		// Email style local account
		{"acct:juliet%40capulet.example@shoppingsite.example", "shoppingsite.example"},
		{"acct:juliet@capulet.example@shoppingsite.example", "shoppingsite.example"},
	}

	for _, tt := range tests {
		r, _ := Parse(tt.input)
		got := r.WebFingerHost()
		if !cmp.Equal(got, tt.want) {
			t.Errorf("WebFingerHost(%q) returned %#v, want %#v", tt.input, got, tt.want)
		}
	}
}

func TestResource_JRDURL(t *testing.T) {
	r, _ := Parse("bob@example.com")
	got := r.JRDURL([]string{"a", "b"})
	want, _ := url.Parse("https://example.com/.well-known/webfinger?" +
		"rel=a&rel=b&resource=acct%3Abob%40example.com")
	if !cmp.Equal(got, want) {
		t.Errorf("JRDURL() returned: %#v, want %#v", got, want)
	}
}

func TestResource_String(t *testing.T) {
	r, _ := Parse("bob@example.com")
	if got, want := r.String(), "acct:bob@example.com"; got != want {
		t.Errorf("String() returned: %#v, want %#v", got, want)
	}
}

func TestLookup(t *testing.T) {
	client, mux, host, teardown := setup()
	defer teardown()

	mux.HandleFunc("/.well-known/webfinger", func(w http.ResponseWriter, r *http.Request) {
		resource := r.FormValue("resource")
		if want := "acct:bob@" + host; resource != want {
			t.Errorf("Requested resource: %v, want %v", resource, want)
		}
		w.Header().Add("content-type", "application/jrd+json")
		fmt.Fprint(w, `{"subject":"bob@example.com"}`)
	})

	jrd, err := client.Lookup("acct:bob@"+host, nil)
	if err != nil {
		t.Errorf("Unexpected error lookup up webfinger: %v", err)
	}
	want := &JRD{Subject: "bob@example.com"}
	if !cmp.Equal(jrd, want) {
		t.Errorf("Lookup returned %#v, want %#v", jrd, want)
	}
}

func TestLookup_parseError(t *testing.T) {
	// use default client here, just to make sure that gets tested
	_, err := Lookup("bob", nil)
	if err == nil {
		t.Error("Expected parse error")
	}
}

func TestLookup_404(t *testing.T) {
	client, _, host, teardown := setup()
	defer teardown()

	_, err := client.Lookup("bob@"+host, nil)
	if err == nil {
		t.Error("Expected error")
	}
}
