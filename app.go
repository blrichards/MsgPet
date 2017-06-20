package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"reflect"
	"strconv"
	"time"

	"github.com/iot-dsa-v2/MsgPET/transforms"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

var animals = map[string]int{
	"mouse":    8,
	"chicken":  16,
	"pig":      32,
	"goat":     64,
	"zebra":    128,
	"rhino":    256,
	"hippo":    512,
	"elephant": 1024,
	"whale":    2048,
}

type appConfig struct {
	Port         int    `yaml:"port"`
	Host         string `yaml:"host"`
	Requests     uint64 `yaml:"requests"`
	Delay        string `yaml:"delay"`
	SingleSocket bool   `yaml:"single-socket"`
	MessageSize  string `yaml:"message-size"`
}

var defaults appConfig
var testDone = make(chan bool)

const helpString = `name:         MsgPET
description:  server stress testing tool
version:      v0.0
author:       Ben Richards
project page: https://github.com/iot-dsa-v2/MsgPET

OPTION              | ARG                    | DESCRIPTION
====================|========================|==============================
-m, -size           | int or animal name     | message size
-t, -tests          | int                    | number of tests
-h, -host           | ip address             | hostname of server to test
-p, -port           | port number            | port number of server to test
-d, -delay          | duration (ex. 100ms)   | delay between requests
-s, -single-socket  | none                   | use one socket for test (ignores delay)`

func init() {
	// parse config file at ~/.msgpetrc for default args
	home, err := homedir.Dir()
	if err != nil {
		fatal("%s", err)
	}
	configFile, err := ioutil.ReadFile(path.Join(home, ".msgpetrc"))
	if err != nil {
		fatal("%s", err)
	}
	err = yaml.Unmarshal(configFile, &defaults)
	if err != nil {
		fatal("%s", err)
	}
}

func main() {
	var config appConfig
	parseArgs(&config, os.Args[1:])

	testConfig := TestConfig{
		HostAddr: fmt.Sprintf("%s:%d", config.Host, config.Port),
		Requests: config.Requests,
		Delay:    handleError(time.ParseDuration(config.Delay)).(time.Duration),
		Message:  genMessage(config.MessageSize),
	}

	var results Test
	if config.SingleSocket {
		results = TestSingle(testConfig)
	} else {
		results = TestMult(testConfig)
	}

	// print summary
	fmt.Println("Total test time:", results.TotalTime)
	fmt.Println("Average response time:")
	fmt.Println("Successful requests:", results.SuccessfulRequests)
	fmt.Println("Failed requests:", results.FailedRequests)
	if results.FailedRequests > 0 {
		fmt.Println("\nErrors (reason -> frequency):")
	} else {
		fmt.Printf("\nServer was able to successfully handle %d requests in %s (%f clients/sec)\n",
			config.Requests,
			results.TotalTime.String(),
			float64(config.Requests)/(float64(results.TotalTime)*float64(1e-9)))
	}
	for err, count := range results.Errors {
		fmt.Printf("\t%s -> %d\n", err, count)
	}
}

func fatal(format string, args ...interface{}) {
	fmt.Print("\nAn unexpected error occured:\n\t")
	fmt.Println(fmt.Sprintf(format, args...))
	fmt.Println("\nTry `MsgPET help` for a list of valid command line arguments")
	fmt.Println()
	os.Exit(1)
}

func parseArgs(config *appConfig, args []string) {
	var i int
	// setup wrapper functions
	handleConv := func(value interface{}, err error) interface{} {
		if err != nil {
			fatal("invalid option, %s, for argument %s", args[i+1], args[i])
		}
		return value
	}
	needsVal := func(setter func()) {
		if i+1 >= len(args) {
			fatal("invalid option, null, for argument " + args[i])
		}
		setter()
		i++
	}

	// check for help call
	if len(args) == 1 && args[0] == "help" {
		fmt.Println(helpString)
		os.Exit(0)
	}

	// parse args
	for i = 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		default:
			fatal("invalid command line argument " + args[i])
		case "-m", "-size":
			needsVal(func() { config.MessageSize = args[i+1] })
		case "-t", "-tests":
			needsVal(func() { config.Requests = uint64(handleConv(strconv.Atoi(args[i+1])).(int)) })
		case "-h", "-host":
			needsVal(func() { config.Host = args[i+1] })
		case "-p", "-port":
			needsVal(func() { config.Port = handleConv(strconv.Atoi(args[i+1])).(int) })
		case "-d", "-delay":
			needsVal(func() { config.Delay = args[i+1] })
		case "-s", "-single-socket":
			config.SingleSocket = true
		}
	}

	// set defaults
	v := reflect.ValueOf(config).Elem()
	d := reflect.ValueOf(defaults)
	for i := 0; i < v.NumField(); i++ {
		if isZero(v.Field(i)) && v.Type().Field(i).Name != "SingleSocket" {
			if isZero(d.Field(i)) {
				fatal("The `%s` value was not set from command line and no default value exists. "+
					"Please set a default value in ~/.msgpetrc or specify value in command line.",
					transforms.Underscore(v.Type().Field(i).Name))
			}
			v.Field(i).Set(d.Field(i))
		}
	}

}

func isZero(f reflect.Value) bool {
	return f.Interface() == reflect.Zero(f.Type()).Interface()
}

func handleError(val interface{}, err error) interface{} {
	if err != nil {
		fatal("%s", err)
	}
	return val
}

func genMessage(sizeString string) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	nextByte := func() byte {
		const chars = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"
		return chars[random.Intn(len(chars))]
	}

	// figure out message size
	var size int
	if val, ok := animals[sizeString]; ok {
		size = val
	} else if val, err := strconv.Atoi(sizeString); err == nil && val > 0 {
		size = val
	} else {
		fatal("invalid message size, " + sizeString)
	}

	// generate random message
	byteArray := make([]byte, size)
	for i := 0; i < size; i++ {
		byteArray[i] = nextByte()
	}

	message := string(byteArray)
	fmt.Print("message: ", message, "\n\n")
	return message
}
