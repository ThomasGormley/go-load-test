package loadtest

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
)

type loadTest struct {
	url              string
	concurrency      int
	numberOfRequests int
}

type Requester interface {
	Request(url string) error
}

type pool struct {
	clients           []Requester
	requestsPerSecond int
	concurrency       int
	url               string
}

type httpClient struct {
}

func (c httpClient) Request(url string) error {
	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	slog.Info("Request complete", "status", resp.StatusCode, "body", body)
	return nil
}

func (p *pool) start() {

	var wg sync.WaitGroup

	for i := 0; i < cap(p.clients); i++ {
		wg.Add(1)
		client := httpClient{}
		p.clients = append(p.clients, client)

		go func() {
			defer wg.Done()
			client.Request(p.url)
		}()
	}

	wg.Wait()
}

func New(url string, c, n int) *loadTest {
	return &loadTest{
		url:              url,
		concurrency:      c,
		numberOfRequests: n,
	}
}

func (lt loadTest) Run() error {
	fmt.Printf("Running load test on %s with concurrency %d and %d requests\n", lt.url, lt.concurrency, lt.numberOfRequests)

	p := pool{
		clients:           make([]Requester, lt.concurrency),
		requestsPerSecond: 1,
		concurrency:       lt.concurrency,
		url:               lt.url,
	}

	p.start()
	// for i := 0; i < lt.concurrency; i++ {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		for j := 0; j < lt.numberOfRequests; j++ {
	// 			if err := lt.makeRequest(); err != nil {
	// 				fmt.Println("Error during request: ", err)
	// 			}
	// 		}

	// 	}()
	// }

	// wg.Wait()
	return nil
}

func (lt loadTest) makeRequest() error {
	resp, err := http.Get(lt.url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	slog.Info("Request complete", "status", resp.StatusCode, "body", body)
	return nil
}
