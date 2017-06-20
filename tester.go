package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Test is used for running a benchmarking test and storing the results
type Test struct {
	SuccessfulRequests uint64
	FailedRequests     uint64
	Errors             map[string]uint64
	AvgResponse        time.Duration
	TotalTime          time.Duration
	results            chan time.Duration
	done               chan bool
	requests           uint64
	mux                *sync.Mutex
}

// TestConfig is required to create a new Test instance
type TestConfig struct {
	HostAddr string
	Requests uint64
	Delay    time.Duration
	Message  string
}

var tester *Test

// NewTest creates a new test instance
func (t *Test) NewTest(requests uint64, mux *sync.Mutex) Test {
	return Test{
		Errors:   make(map[string]uint64),
		results:  make(chan time.Duration, requests),
		done:     make(chan bool),
		requests: requests,
		mux:      mux,
	}
}

// MakeRequest makes a request to the server and saves the response time
func (t *Test) MakeRequest(conn net.Conn, message string) {
	// send message to server
	fmt.Fprintln(conn, message)
	// save start time
	start := time.Now()
	// wait for response and save response time
	_, err := bufio.NewReader(conn).ReadString('\n')
	responseTime := time.Since(start)

	go func() {
		// log results
		if err != nil {
			// save error message for summary if request unsuccessful
			errorArray := strings.SplitAfter(err.Error(), ": ")
			t.mux.Lock()
			t.Errors[errorArray[len(errorArray)-1]]++
			t.mux.Unlock()
			atomic.AddUint64(&t.FailedRequests, 1)
		} else {
			// log successful test and response time
			atomic.AddUint64(&t.SuccessfulRequests, 1)
			t.results <- responseTime
		}
		t.done <- true
	}()
}

// Wait waits until all requests are complete before calculating stats
func (t *Test) Wait() {
	// set start time and wait for requests to complete
	start := time.Now()
	for i := 0; uint64(i) < t.requests; i++ {
		<-t.done
	}
	t.TotalTime = time.Since(start)
	close(t.results)

	// calculate average response time
	t.AvgResponse = t.TotalTime / time.Duration(t.SuccessfulRequests)
}

// TestMult runs test by opening multiple sockets with a delay between each request
func TestMult(config TestConfig) Test {
	fmt.Print("Testing with multiple sockets...\n\n")
	var mux sync.Mutex
	test := tester.NewTest(config.Requests, &mux)

	for i := 0; i < int(config.Requests); i++ {
		go func(conn net.Conn, err error) {
			if err != nil {
				panic(err)
			}
			// start request
			test.MakeRequest(conn, config.Message)
			// close connection
			conn.Close()
		}(net.Dial("tcp", config.HostAddr))
		time.Sleep(config.Delay)
	}
	test.Wait()
	return test
}

// TestSingle runs all requests consecutively instead of concurrently with no delay
func TestSingle(config TestConfig) Test {
	fmt.Print("Testing with single socket...\n\n")
	var mux sync.Mutex
	test := tester.NewTest(config.Requests, &mux)

	// initialize connection to server
	conn, err := net.Dial("tcp", config.HostAddr)
	if err != nil {
		panic(err)
	}

	go func() {
		for i := 0; i < int(config.Requests); i++ {
			test.MakeRequest(conn, config.Message)
		}
	}()
	test.Wait()
	conn.Close()

	return test
}
