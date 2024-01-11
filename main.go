package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	loadtest "github.com/thomasgormley/go-load-test/pkg"
)

func run(args []string) error {

	concurrency := flag.Int("c", 10, "concurrency")
	numberOfRequests := flag.Int("n", 0, "number of requests")

	flag.Parse()

	if len(args) == 0 {
		log.Fatalf("URL cannot be empty")
	}

	url := flag.Arg(0)

	fmt.Println("url:", url)

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		log.Fatalf("URL protocol must match one of: http, https. Got: %s", url)
	}

	return loadtest.Run(url, *concurrency, *numberOfRequests)
}

func main() {
	run(os.Args[1:])
}
