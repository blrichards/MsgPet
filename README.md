# MsgPet
Server benchmarking tool in Go

## Installation

2 options are available

1. install go SDK for your platform and compile code from source
2. download the precompiled executable for your platform located in the 'bin' folder

## Usage

MsgPET is very simple to use. Configuration for the tool relies mainly on a yaml config file ('.msgpetrc') that should be placed in the user's home directory

### Config File

These four key value pairs are required in `~/.msgpetrc` for the tool to run

* `port`:`int` -> port number of server to test
* `host`:`string` -> hostname of server to test
* `clients`:`int` -> number of tests to simulate during test
* `delay`:`string` -> delay between client requests, specified by \[number\]\[unit\] (example: `0.5ms`)

##### Example `.msgpetrc`

```
# .msgpetrc

port: 8080
host: 192.168.1.1
clients: 100
delay: 100ms
```

### Command Line Args

The only command line argument for MsgPET is the size of the message to be generated and used during the test. There are two methods of declaring the size of the message to be sent

1. an `int` representing the the size of the message in bytes
2. an animal! after all it is called MsgPET

  `mouse` -> 8 bytes   
`chicken` -> 16 bytes   
`pig` -> 32 bytes  
`goat` -> 64 bytes  
`zebra` -> 128 bytes  
`rhino` -> 256 bytes  
`hippo` -> 512 bytes  
`elephant` -> 1024 bytes  
`whale` -> 2048 bytes

## Output

Output of the tool is a summary consisting of the message used in the test, sum of response times, average response time, number of successful responses, number of failed responses, and either...

1. the speed of the server in clients/sec if all requests were successfully handled or
2. a list of error messages if any requests were unsuccessful

##### Example Output

```
$ ./MsgPET chicken
message: ySmUAPHdzzZLzDdL

Sum of successful response times: 22.007701ms
Average successful response time: 220.077µs
Successful tests: 100
Failed tests: 0

Server was able to successfully handle 100 requests in 57.973972ms (1724.912000 clients/sec)
```
