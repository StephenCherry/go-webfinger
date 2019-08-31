// The webfinger tool is a command line client for performing webfinger lookups.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"webfinger.net/go/webfinger"
)

var (
	verbose = flag.Bool("v", false, "print details about the resolution")
)

func usage() {
	fmt.Println("webfinger [-v] <resource uri>")
	flag.PrintDefaults()
	fmt.Println("\nexample: webfinger -v bob@example.com") // same Bob as in the draft
}

func main() {
	flag.Usage = usage
	flag.Parse()

	resource := flag.Arg(0)
	if resource == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.SetFlags(0)
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}

	client := webfinger.NewClient(nil)
	client.AllowHTTP = true

	jrd, err := client.Lookup(resource, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(jrd)
}
