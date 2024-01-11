package loadtest

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

type Requester interface {
	Request(url string) error
}

type requestPool struct {
	createClient      func() Requester
	clients           []Requester
	idleClients       []Requester
	clientMu          sync.Mutex
	requestsPerSecond int
	numberOfRequests  int
	size              int
	url               string

	requestCount   int
	requestCountMu sync.Mutex
}

func (p *requestPool) addClient() Requester {
	client := p.createClient()
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

func (p *requestPool) startRequests(wg *sync.WaitGroup) {
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

func (p *requestPool) incrementRequestCount() int {

	p.requestCountMu.Lock()
	defer p.requestCountMu.Unlock()
	p.requestCount++

	return p.requestCount
}

func (p *requestPool) clientPoolExhausted() bool {
	p.clientMu.Lock()
	defer p.clientMu.Unlock()
	c := len(p.idleClients)

	return c == 0
}

func (p *requestPool) start() {

	var wg sync.WaitGroup

	for i := 0; i < p.size; i++ {
		fmt.Printf("%d < %d\n", i, p.size)
		wg.Add(1)
		p.addClient()
		go p.startRequests(&wg)
	}

	wg.Wait()
}

func createRequestClient(url string) Requester {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return httpClient{
			url: url,
		}
	}

	return httpClient{
		url: url,
	}
}

func Run(url string, concurrency, numberOfRequests int) error {
	fmt.Printf("Running load test on %s with concurrency %d and %d requests\n", url, concurrency, numberOfRequests)

	createClient := func() Requester {
		return createRequestClient(url)
	}

	p := requestPool{
		createClient:      createClient,
		requestsPerSecond: 1,
		size:              concurrency,
		numberOfRequests:  numberOfRequests,
		url:               url,
	}

	p.start()
	return nil
}
