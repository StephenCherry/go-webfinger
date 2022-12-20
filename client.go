// Package webfinger provides a simple client implementation of the WebFinger
// protocol.
//
// It is a work in progress, the API is not frozen.
// We're trying to catchup with the last draft of the protocol:
// http://tools.ietf.org/html/draft-ietf-appsawg-webfinger-14
//
// Example:
//
//	package main
//
//	import (
//	        "fmt"
//	        "os"
//
//	        "webfinger.net/go/webfinger"
//	)
//
//	func main() {
//	        email := os.Args[1]
//
//	        client := webfinger.NewClient(nil)
//	        client.AllowHTTP = true
//
//	        jrd, err := client.Lookup(email, nil)
//	        if err != nil {
//	                fmt.Println(err)
//	                return
//	        }
//
//	        fmt.Printf("JRD: %+v", jrd)
//	}
package webfinger

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Resource is a resource for which a WebFinger query can be issued.
type Resource url.URL

// Parse parses rawurl into a WebFinger Resource.  The rawurl should be an
// absolute URL, or an email-like identifier (e.g. "bob@example.com").
func Parse(rawurl string) (*Resource, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	// if parsed URL has no scheme but is email-like, treat it as an acct: URL.
	if u.Scheme == "" {
		parts := strings.SplitN(rawurl, "@", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("URL must be absolute, or an email address: %v", rawurl)
		}
		return Parse("acct:" + rawurl)
	}

	r := Resource(*u)
	return &r, nil
}

// WebFingerHost returns the default host for issuing WebFinger queries for
// this resource.  For Resource URLs with a host component, that value is used.
// For URLs that do not have a host component, the host is determined by other
// mains if possible (for example, the domain in the addr-spec of a mailto
// URL).  If the host cannot be determined from the URL, this value will be an
// empty string.
func (r *Resource) WebFingerHost() string {
	if r.Host != "" {
		return r.Host
	} else if r.Scheme == "acct" || r.Scheme == "mailto" {
		parts := strings.SplitN(r.Opaque, "@", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return ""
}

// String reassembles the Resource into a valid URL string.
func (r *Resource) String() string {
	u := url.URL(*r)
	return u.String()
}

// JRDURL returns the WebFinger query URL for this resource. If rels is
// specified, it will be included in the query URL.
func (r *Resource) JRDURL(rels []string) *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   r.WebFingerHost(),
		Path:   "/.well-known/webfinger",
		RawQuery: url.Values{
			"resource": []string{r.String()},
			"rel":      rels,
		}.Encode(),
	}
}

// A Client is a WebFinger client.
type Client struct {
	// HTTP client used to perform WebFinger lookups.
	client *http.Client

	// Allow the use of HTTP endoints for lookups.  The WebFinger spec requires
	// all lookups be performed over HTTPS, so this should only ever be enabled
	// for development.
	AllowHTTP bool

	// Logger used during webfinger fetching.
	Logger *log.Logger
}

// DefaultClient is the default Client and is used by Lookup.
var DefaultClient = &Client{
	client: http.DefaultClient,
}

// Lookup returns the JRD for the specified identifier.
//
// Lookup is a wrapper around DefaultClient.Lookup.
func Lookup(identifier string, rels []string) (*JRD, error) {
	return DefaultClient.Lookup(identifier, rels)
}

// NewClient returns a new WebFinger Client.  If a nil http.Client is provied,
// http.DefaultClient will be used.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		client: httpClient,
	}
}

// Lookup returns the JRD for the specified identifier.  If provided, only the
// specified rel values will be requested, though WebFinger servers are not
// obligated to respect that request.
func (c *Client) Lookup(identifier string, rels []string) (*JRD, error) {
	resource, err := Parse(identifier)
	if err != nil {
		return nil, err
	}

	return c.LookupResource(resource, rels)
}

// LookupResource returns the JRD for the specified Resource.  If provided,
// only the specified rel values will be requested, though WebFinger servers
// are not obligated to respect that request.
func (c *Client) LookupResource(resource *Resource, rels []string) (*JRD, error) {
	c.logf("Looking up WebFinger data for %s", resource)

	resourceJRD, err := c.fetchJRD(resource.JRDURL(rels))
	if err != nil {
		return nil, err
	}

	return resourceJRD, nil
}

func (c *Client) fetchJRD(jrdURL *url.URL) (*JRD, error) {
	// TODO verify signature if not https
	// TODO extract http cache info

	// Get follows up to 10 redirects
	c.logf("GET %s", jrdURL.String())
	res, err := c.client.Get(jrdURL.String())
	if err != nil {
		errString := strings.ToLower(err.Error())
		// For some crazy reason, App Engine returns a "ssl_certificate_error" when
		// unable to connect to an HTTPS URL, so we check for that as well here.
		if (strings.Contains(errString, "connection refused") ||
			strings.Contains(errString, "ssl_certificate_error")) && c.AllowHTTP {
			jrdURL.Scheme = "http"
			c.logf("GET %s", jrdURL.String())
			res, err = c.client.Get(jrdURL.String())
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if !(200 <= res.StatusCode && res.StatusCode < 300) {
		return nil, errors.New(res.Status)
	}

	content, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	return ParseJRD(content)
}

func (c *Client) logf(format string, v ...interface{}) {
	if c.Logger != nil {
		c.Logger.Printf(format, v...)
	} else {
		log.Printf(format, v...)
	}
}
