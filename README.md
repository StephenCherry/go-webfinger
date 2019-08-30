# go-webfinger

[![GoDoc](https://godoc.org/webfinger.net/go/webfinger?status.svg)](https://godoc.org/webfinger.net/go/webfinger)
[![Build Status](https://travis-ci.org/webfinger/go-webfinger.svg?branch=master)](https://travis-ci.org/webfinger/go-webfinger)
[![Test Coverage](https://codecov.io/gh/webfinger/go-webfinger/branch/master/graph/badge.svg)](https://codecov.io/gh/webfinger/go-webfinger)

go-webfinger is a Go client for the [Webfinger protocol](https://webfinger.net).

## Usage

Install using:

    go get webfinger.net/go/webfinger


A simple example of using the package:

``` go
package main

import (
    "fmt"
    "os"

    "webfinger.net/go/webfinger"
)

func main() {
    email := os.Args[1]

    client := webfinger.NewClient(nil)

    jrd, err := client.Lookup(email, nil)
    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Printf("JRD: %+v", jrd)
}
```
