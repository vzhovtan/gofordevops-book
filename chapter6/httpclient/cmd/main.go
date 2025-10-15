package main

import "github.com/vzhovtan/gofordevops/chapter6/httpclient"

func main() {
	httpclient.HttpGet("http://www.google.com")
}
