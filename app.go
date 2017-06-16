package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"strings"
	"sync"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Port    int
	Host    string
	Clients int
	Delay   string
}

var mux sync.Mutex
var config Config
var successfulTests = 0
var failedTests = 0
var errors = make(map[string]int64)

func runTest(message string) {
	conn, err := net.Dial("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(conn, message)
	start := time.Now()

	_, err = bufio.NewReader(conn).ReadString('\n')
	responseTime := time.Since(start)

	if err != nil {
		errorArray := strings.SplitAfter(err.Error(), ": ")
		mux.Lock()
		errors[errorArray[len(errorArray)-1]]++
		mux.Unlock()
		failedTests++
	} else {
		successfulTests++
		results <- responseTime
	}
	conn.Close()

	testDone <- true
}

var results chan time.Duration
var testDone = make(chan bool)

func init() {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	configFile, err := ioutil.ReadFile(path.Join(home, ".msgpetrc"))
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		panic(err)
	}

	results = make(chan time.Duration, config.Clients)
}

func main() {
	if len(os.Args) != 2 {
		if len(os.Args) < 2 {
			fmt.Println("please provide a message for test clients to send")
		} else {
			fmt.Println("too many command line arguments")
		}
		return
	} else if config.Clients < 1 {
		fmt.Println("number of clients must be greater than zero")
		return
	}

	// set delay between client requests
	delay, err := time.ParseDuration(config.Delay)

	if err != nil {
		fmt.Println("invalid delay, please input a duration in the format <int><unit>. (ex. 200ms)")
	}

	// run tests
	start := time.Now()
	for i := 0; i < config.Clients; i++ {
		go runTest(os.Args[1])
		time.Sleep(delay)
	}

	// wait for all tests to complete
	for i := 0; i < config.Clients; i++ {
		<-testDone
	}
	stopTime := time.Since(start)

	// close the results channel
	close(results)

	// calculate sum of all valid response times
	var totalResponseTime time.Duration = 0
	for result := range results {
		totalResponseTime += result
	}

	// print summary
	fmt.Println("Sum of successful response times:", totalResponseTime)
	fmt.Println("Average successful response time:", totalResponseTime/time.Duration(successfulTests))
	fmt.Println("Successful tests:", successfulTests)
	fmt.Println("Failed tests:", failedTests)
	if failedTests > 0 {
		fmt.Println("\nErrors (reason -> frequency):")
	} else {
		fmt.Printf("\nServer was able to successfully handle %d requests in %s\n", config.Clients, stopTime.String())
	}
	for err, count := range errors {
		fmt.Printf("\t%s -> %d\n", err, count)
	}
}
