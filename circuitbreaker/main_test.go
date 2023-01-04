package main

import (
	"context"
	"fmt"
	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var (
	non200CodeResponseError = fmt.Errorf("non 200 response code")
)

type HTTP struct {
	client *http.Client
	cb     *gobreaker.CircuitBreaker
}

func NewCB(client *http.Client) (*HTTP, chan gobreaker.State) {
	ch := make(chan gobreaker.State, 1)

	h := &HTTP{
		client: client,
		cb: gobreaker.NewCircuitBreaker(
			gobreaker.Settings{
				// When to flush counters int the Closed state
				Interval: 50 * time.Millisecond,
				// Time to switch from Open to Half-open
				Timeout: 70 * time.Millisecond,
				// Function with check when to switch from Closed to Open
				ReadyToTrip: func(counts gobreaker.Counts) bool {
					failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
					log.Println(fmt.Sprintf("failureRatio: %.2f counts.Requests: %d", failureRatio, counts.Requests))
					return counts.Requests >= 3 && failureRatio >= 0.6
				},
				MaxRequests: 3,
				OnStateChange: func(name string, from, to gobreaker.State) {
					log.Println(fmt.Sprintf("State changed from %v to %v", from.String(), to.String()))
					ch <- to
				},
			},
		),
	}
	return h, ch
}

func NewCBWithSettings(client *http.Client, settings gobreaker.Settings) *HTTP {
	return &HTTP{
		client: client,
		cb:     gobreaker.NewCircuitBreaker(settings),
	}
}

func (h *HTTP) Get(req *http.Request) (*http.Response, error) {
	if _, ok := req.Context().Deadline(); !ok {
		return nil, fmt.Errorf("all requests must have a Context deadline set")
	}

	r, err := h.cb.Execute(
		func() (interface{}, error) {
			resp, err := h.client.Do(req)
			if err != nil {
				return nil, err
			}
			if resp.StatusCode != 200 {
				return nil, non200CodeResponseError
			}
			return resp, nil
		},
	)
	if err != nil {
		return nil, err
	}
	return r.(*http.Response), nil
}

type testhandler struct {
	returnOkResponse *bool
}

func (t testhandler) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	if *(t.returnOkResponse) {
		rw.WriteHeader(http.StatusOK)
		_, err := rw.Write([]byte("ok"))
		if err != nil {
			fmt.Println(err)
		}
	} else {
		rw.WriteHeader(http.StatusServiceUnavailable)
	}
}

func TestCircuitBreakerWith3MaxRequests(t *testing.T) {
	tempValue := false
	handler := testhandler{&tempValue}
	server := httptest.NewServer(handler)
	log.Println(fmt.Sprintf("Internal server running at %s", server.URL))
	defer server.Close()

	c, c1 := NewCB(&http.Client{})

	loopBreak := false
	for i := 0; i < 100; i++ {
		select {
		case msg1 := <-c1:
			log.Println(fmt.Sprintf("Received from channel < %v > at iternation: %d", msg1.String(), i))
			*(handler.returnOkResponse) = true
			if msg1.String() == "closed" {
				loopBreak = true
			}
		default:
			if loopBreak {
				break
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
			if err != nil {
				panic(err)
			}
			resp, err := c.Get(req)
			if err != nil {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			defer resp.Body.Close()
			all, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			log.Println(fmt.Sprintf("Received value from circuit: [ %v ] at iteration %d", string(all), i))
			time.Sleep(5 * time.Millisecond)
		}
	}
	assert.True(t, loopBreak)
}

func TestCircuitBreakerWith20MaxRequests(t *testing.T) {
	tempValue := false
	handler := testhandler{&tempValue}
	server := httptest.NewServer(handler)
	log.Println(fmt.Sprintf("Internal server running at %s", server.URL))
	defer server.Close()
	ch := make(chan gobreaker.State, 1)

	settings := gobreaker.Settings{
		// When to flush counters int the Closed state
		Interval: 50 * time.Millisecond,
		// Time to switch from Open to Half-open
		Timeout: 70 * time.Millisecond,
		// Function with check when to switch from Closed to Open
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			log.Println(fmt.Sprintf("failureRatio: %.2f counts.Requests: %d", failureRatio, counts.Requests))
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		MaxRequests: 20,
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Println(fmt.Sprintf("State changed from %v to %v", from.String(), to.String()))
			ch <- to
		},
	}

	c := NewCBWithSettings(&http.Client{}, settings)

	loopBreak := false
	for i := 0; i < 100; i++ {
		select {
		case msg1 := <-ch:
			log.Println(fmt.Sprintf("Received from channel < %v > at iternation: %d", msg1.String(), i))
			*(handler.returnOkResponse) = true
			if msg1.String() == "closed" {
				loopBreak = true
			}
		default:
			if loopBreak {
				break
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
			if err != nil {
				panic(err)
			}
			resp, err := c.Get(req)
			if err != nil {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			defer resp.Body.Close()
			all, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			log.Println(fmt.Sprintf("Received value from circuit: [ %v ] at iteration %d", string(all), i))
			time.Sleep(5 * time.Millisecond)
		}
	}
	assert.True(t, loopBreak)
}

func TestCircuitBreakerWith0Interval(t *testing.T) {
	tempValue := false
	handler := testhandler{&tempValue}
	server := httptest.NewServer(handler)
	log.Println(fmt.Sprintf("Internal server running at %s", server.URL))
	defer server.Close()
	ch := make(chan gobreaker.State, 1)

	settings := gobreaker.Settings{
		// When to flush counters int the Closed state
		Interval: 0,
		// Time to switch from Open to Half-open
		Timeout: 70 * time.Millisecond,
		// Function with check when to switch from Closed to Open
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			log.Println(fmt.Sprintf("failureRatio: %.2f counts.Requests: %d", failureRatio, counts.Requests))
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		MaxRequests: 3,
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Println(fmt.Sprintf("State changed from %v to %v", from.String(), to.String()))
			ch <- to
		},
	}

	c := NewCBWithSettings(&http.Client{}, settings)

	loopBreak := false
	for i := 0; i < 100; i++ {
		select {
		case msg1 := <-ch:
			log.Println(fmt.Sprintf("Received from channel < %v > at iternation: %d", msg1.String(), i))
			*(handler.returnOkResponse) = true
			if msg1.String() == "closed" {
				loopBreak = true
			}
		default:
			if loopBreak {
				break
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
			if err != nil {
				panic(err)
			}
			resp, err := c.Get(req)
			if err != nil {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			defer resp.Body.Close()
			all, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			log.Println(fmt.Sprintf("Received value from circuit: [ %v ] at iteration %d", string(all), i))
			time.Sleep(5 * time.Millisecond)
		}
	}
	assert.True(t, loopBreak)
}
