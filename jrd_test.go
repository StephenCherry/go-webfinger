package webfinger

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestParseJRD(t *testing.T) {
	// Example JRD from http://tools.ietf.org/html/rfc6415#appendix-A
	blob := `
    {
      "subject":"http://blog.example.com/article/id/314",
      "expires":"2010-01-30T09:30:00Z",

      "aliases":[
        "http://blog.example.com/cool_new_thing",
        "http://blog.example.com/steve/article/7"],

      "properties":{
        "http://blgx.example.net/ns/version":"1.3",
        "http://blgx.example.net/ns/ext":null
      },

      "links":[
        {
          "rel":"author",
          "type":"text/html",
          "href":"http://blog.example.com/author/steve",
          "titles":{
            "default":"About the Author",
            "en-us":"Author Information"
          },
          "properties":{
            "http://example.com/role":"editor"
          }
        },
        {
          "rel":"author",
          "href":"http://example.com/author/john",
          "titles":{
            "default":"The other author"
          }
        },
        {
          "rel":"copyright",
          "template":"http://example.com/copyright?id={uri}"
        }
      ]
    }
        `
	obj, err := ParseJRD([]byte(blob))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := obj.Subject, "http://blog.example.com/article/id/314"; got != want {
		t.Errorf("JRD.Subject is %q, want %q", got, want)
	}
	expires := time.Date(2010, 01, 30, 9, 30, 0, 0, time.UTC)
	if got, want := obj.Expires, &expires; !cmp.Equal(got, want) {
		t.Errorf("JRD.Expires is %q, want %q", got, want)
	}

	// Properties
	if got, want := obj.GetProperty("http://blgx.example.net/ns/version"), "1.3"; got != want {
		t.Errorf("obj.GetProperty('http://blgx.example.net/ns/version') returned %q, want %q", got, want)
	}
	if got, want := obj.GetProperty("http://blgx.example.net/ns/ext"), ""; got != want {
		t.Errorf("obj.GetProperty('http://blgx.example.net/ns/ext') returned %q, want %q", got, want)
	}
	if got, want := obj.GetProperty("does-not-exist"), ""; got != want {
		t.Errorf("obj.GetProperty('does-not-exist') returned %q, want %q", got, want)
	}

	// Links
	if obj.GetLinkByRel("copyright") == nil {
		t.Error("obj.GetLinkByRel('copyright') returned nil, want non-nil value")
	}
	if got, want := obj.GetLinkByRel("copyright").Template, "http://example.com/copyright?id={uri}"; got != want {
		t.Errorf("obj.GetLinkByRel('copyright').Template returned %q, want %q", got, want)
	}
	if got, want := obj.GetLinkByRel("author").Titles["default"], "About the Author"; got != want {
		t.Errorf("obj.GetLinkByRel('author').Titles['default'] returned %q, want %q", got, want)
	}
	if got, want := obj.GetLinkByRel("author").GetProperty("http://example.com/role"), "editor"; got != want {
		t.Errorf("obj.GetLinkByRel('author').GetProperty('http://example.com/role') returned %q, want %q", got, want)
	}
	if got, want := obj.GetLinkByRel("does-not-exist"), (*Link)(nil); got != want {
		t.Errorf("obj.GetLinkByRel('does-not-exist') returned %q, want %q", got, want)
	}
	if got, want := obj.GetLinkByRel("author").GetProperty("does-not-exist"), ""; got != want {
		t.Errorf("obj.GetLinkByRel('author').GetProperty('does-not-exist') returned %q, want %q", got, want)
	}
}

func TestParseJRD_error(t *testing.T) {
	_, err := ParseJRD([]byte("`"))
	if err == nil {
		t.Errorf("ParseJRD(`) did not return expected error")
	}
}
