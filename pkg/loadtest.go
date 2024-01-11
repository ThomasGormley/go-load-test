package loadtest

import (
	"fmt"
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
	idleClients       []Requester
	clientMu          sync.Mutex
	requestsPerSecond int
	numberOfRequests  int
	concurrency       int
	url               string

	requestCount   int
	requestCountMu sync.Mutex
}

func (p *pool) addClient() Requester {
	client := httpClient{
		url: p.url,
	}
	p.clients = append(p.clients, client)
	p.idleClients = append(p.clients, client)
	return client
}

type httpClient struct {
	url string
}

func (c httpClient) Request(url string) error {
	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	// defer resp.Body.Close()
	// _, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	slog.Info("Request complete", "status", resp.StatusCode)
	return nil
}

func (p *pool) startRequests(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		if p.clientPoolExhausted() {
			p.clientMu.Unlock()
			continue
		}

		p.clientMu.Lock()
		client := p.idleClients[0]
		p.idleClients = p.idleClients[1:]
		p.clientMu.Unlock()

		if err := client.Request(p.url); err != nil {
			fmt.Println("Error during request: ", err)
		}

		p.incrementRequestCount()
		shouldBreak := p.numberOfRequests > 0 && p.requestCount >= p.numberOfRequests
		if shouldBreak {
			break
		}
		p.clientMu.Lock()
		p.idleClients = append(p.idleClients, client)
		p.clientMu.Unlock()
	}

}

func (p *pool) incrementRequestCount() int {

	p.requestCountMu.Lock()
	defer p.requestCountMu.Unlock()
	p.requestCount++

	return p.requestCount
}

func (p *pool) clientPoolExhausted() bool {
	p.clientMu.Lock()
	defer p.clientMu.Unlock()
	c := len(p.idleClients)

	return c == 0
}

func (p *pool) start() {

	var wg sync.WaitGroup

	for i := 0; i < p.concurrency; i++ {
		fmt.Printf("%d < %d\n", i, p.concurrency)
		wg.Add(1)
		p.addClient()
		go p.startRequests(&wg)
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
		requestsPerSecond: 1,
		concurrency:       lt.concurrency,
		url:               lt.url,
		numberOfRequests:  lt.numberOfRequests,
	}

	p.start()
	return nil
}
