package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"
)

// VERSION WRK
const VERSION = "0.1"
const jsonfile = "/etc/request.json"

var help bool
var goroutines int
var duration int
var timeout int
var numThreads int
var configFile bool
var method string
var proxy string
var headerstr string

type reqconf struct {
	URL    string
	Method string
	Header map[string]string
}

type requeststats struct {
	TotRespSize    int64
	TotDuration    time.Duration
	MinRequestTime time.Duration
	MaxRequestTime time.Duration
	NumRequests    int
	NumRequestErrs int
	NumConnectErrs int
	NumTimeoutErrs int
	Status         map[int]int
}

type wrkconf struct {
	Goroutines  int
	Duration    int
	NumThreads  int
	Timeout     int
	Interrupted int32
	Reqest      []reqconf
	Statschan   chan requeststats
}

func init() {
	flag.BoolVar(&help, "h", false, "help")
	flag.IntVar(&goroutines, "c", 10, "Connections to keep open")
	flag.IntVar(&duration, "d", 10, "Duration of test in seconds")
	flag.IntVar(&numThreads, "t", 10, "Number of threads to use")
	flag.IntVar(&timeout, "T", 5000, "Socket/request timeout in ms")
	flag.BoolVar(&configFile, "F", false, "Run Wrk From Json config file")
	flag.StringVar(&proxy, "x", "", "proxy host[:port] Use this proxy")
	flag.StringVar(&method, "M", "GET", "HTTP method")
	flag.StringVar(&headerstr, "H", "", "Request header")
	flag.Usage = flag.PrintDefaults
}

func parseJSON(path string) (requests []reqconf) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = json.Unmarshal(data, &requests)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return
}

func doProxy(requrl string, header map[string]string) (newurl string) {
	typeURL, err := url.Parse(requrl)
	if err != nil && !configFile {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(proxy) == 0 {
		newurl = requrl
	} else {
		header["Host"] = typeURL.Host
		typeURL.Host = proxy
		newurl = typeURL.String()
	}
	return
}

func newWrkConf() (wc *wrkconf) {
	headert := make(map[string]string)
	if len(headerstr) != 0 {
		headerPairs := strings.Split(headerstr, ";")
		for _, hdr := range headerPairs {
			hp := strings.Split(hdr, ":")
			headert[hp[0]] = hp[1]
		}
	}
	url := doProxy(flag.Arg(0), headert)
	var reqconftmp []reqconf
	if configFile {
		reqconftmp = parseJSON(jsonfile)
	} else {
		reqconftmp = []reqconf{
			{url, method, headert},
		}
	}
	fmt.Println("\nGo-Wrk Version:", VERSION, " Running Test Begin ...")
	fmt.Println("Go-Wrk Threads:", numThreads, "   Connections:", goroutines, "   Duration:", duration, "s", "   Timeout:", timeout, "ms")
	statsChan := make(chan requeststats, goroutines)
	wc = &wrkconf{goroutines, duration, numThreads, timeout, 0, reqconftmp, statsChan}
	return
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(numThreads)

	if help {
		flag.Usage()
		return
	} else if flag.Arg(0) == "" && !configFile {
		fmt.Println("Need request url")
		return
	}

	conf := newWrkConf()
	for i := 0; i < goroutines; i++ {
		go conf.runwrkroutines()
	}
	out := requeststats{MinRequestTime: time.Minute, Status: map[int]int{}}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	responders := 0
	for responders < goroutines {
		select {
		case <-sigChan:
			conf.stopwrkroutines()
		case stats := <-conf.Statschan:
			resultCollection(&stats, &out)
			responders++
		}
	}
	templateOut(out, responders)
	fmt.Println("")
}
