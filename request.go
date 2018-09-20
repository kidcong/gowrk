package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//USERAGENT for UA
const USERAGENT = "wrk"

var mutex sync.Mutex

//EstimateHTTPHeadersSize had to create this because headers size was not counted
func estimateHTTPHeadersSize(headers http.Header) (result int64) {
	result = 0
	for k, v := range headers {
		result += int64(len(k) + len(": \r\n"))
		for _, s := range v {
			result += int64(len(s))
		}
	}
	result += int64(len("\r\n"))
	return
}

func maxDuration(d1 time.Duration, d2 time.Duration) time.Duration {
	if d1 > d2 {
		return d1
	}
	return d2
}

func minDuration(d1 time.Duration, d2 time.Duration) time.Duration {
	if d1 < d2 {
		return d1
	}
	return d2
}

//httprequest function refer to github.com/tsliwowicz/
func httprequest(httpClient *http.Client, header map[string]string, method string, url string) (respSize int, stat int) {
	duration = -1
	respSize = -1
	stat = -1
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println("An error occured doing request", err)
		return
	}
	mutex.Lock()
	for hk, hv := range header {
		if hk == "Host" {
			req.Host = hv
		}
		req.Header.Add(hk, hv)
	}
	mutex.Unlock()
	req.Header.Add("User-Agent", USERAGENT)

	resp, err := httpClient.Do(req)
	if err != nil {
		//fmt.Println("An error occured doing request", err)
		errstr := fmt.Sprintf("%s", err)
		if strings.Contains(errstr, "connection refused") {
			stat = -502
		} else if strings.Contains(errstr, "timeout awaiting response headers") {
			stat = -504
		}
		return
	}

	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("An error occured reading body", err)
	} else {
		respSize = len(body) + int(estimateHTTPHeadersSize(resp.Header))
		stat = resp.StatusCode
	}
	return
}

func (wc *wrkconf) runwrkroutines() {
	httpClient := &http.Client{}
	httpClient.Transport = &http.Transport{
		ResponseHeaderTimeout: time.Millisecond * time.Duration(timeout),
	}

	stats := requeststats{MinRequestTime: time.Minute, Status: map[int]int{}}
	start := time.Now()

	var reqestmp reqconf
	for time.Since(start).Seconds() <= float64(wc.Duration) && atomic.LoadInt32(&wc.Interrupted) == 0 {
		startq := time.Now()
		reqestmp = wc.Reqest[rand.Intn(len(wc.Reqest))]
		if len(proxy) != 0 && configFile {
			mutex.Lock()
			reqestmp.URL = doProxy(reqestmp.URL, reqestmp.Header)
			mutex.Unlock()
		}
		respSize, statusCode := httprequest(httpClient, reqestmp.Header, reqestmp.Method, reqestmp.URL)
		if statusCode > 0 && statusCode < 600 {
			stats.Status[statusCode]++
		} else {
			switch statusCode {
			case -502:
				stats.NumConnectErrs++
			case -504:
				stats.NumTimeoutErrs++
			default:
				stats.NumRequestErrs++
			}
		}
		reqDur := time.Since(startq)
		stats.NumRequests++
		stats.TotRespSize += int64(respSize)
		stats.TotDuration += reqDur
		stats.MaxRequestTime = maxDuration(reqDur, stats.MaxRequestTime)
		stats.MinRequestTime = minDuration(reqDur, stats.MinRequestTime)
	}
	wc.Statschan <- stats
}

func (wc *wrkconf) stopwrkroutines() {
	atomic.StoreInt32(&wc.Interrupted, 1)
	os.Exit(0)
}
