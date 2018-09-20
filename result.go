package main

import (
	"fmt"
	"os"
	"text/template"
	"time"
)

type templStatus struct {
	StatusCode int
	Num        int
	Ratio      float64
}

func mergeMap(now map[int]int, old map[int]int) map[int]int {
	for k, kv := range now {
		if yv, ok := old[k]; ok {
			old[k] = yv + kv
		} else {
			old[k] = kv
		}
	}
	return old
}

func resultCollection(in *requeststats, out *requeststats) {
	out.Status = mergeMap(in.Status, out.Status)
	out.NumRequestErrs += in.NumRequestErrs
	out.NumConnectErrs += in.NumConnectErrs
	out.NumTimeoutErrs += in.NumTimeoutErrs
	out.NumRequests += in.NumRequests
	out.TotRespSize += in.TotRespSize
	out.TotDuration += in.TotDuration
	out.MaxRequestTime = maxDuration(out.MaxRequestTime, in.MaxRequestTime)
	out.MinRequestTime = minDuration(out.MinRequestTime, in.MinRequestTime)
	return
}

func outStatus(statstruct map[int]int, total int) []templStatus {
	statMap := make([]templStatus, len(statstruct))
	var tmp templStatus
	for k, kv := range statstruct {
		tmp = templStatus{k, kv, 100 * float64(kv) / float64(total)}
		statMap = append(statMap, tmp)
	}
	return statMap
}

func templateOut(out requeststats, responders int) {
	const outtemplate = `
	------------------ wrk out -------------------
	
	{{.NumRequests}} requests in {{.AvgThreadDurSec | printf "%.2f" }} s, total {{.TotalSize | printf "%.2f" }} MB read
	{{range .Status}}{{if .StatusCode}} 
	Status: {{.StatusCode}}  num: {{.Num}}  ratio: {{.Ratio | printf "%.2f" }}%{{end}}{{end}}

	Requests/sec: {{.ReqRate | printf "%.2f" }}
	Transfer/sec: {{.BytesRate | printf "%.2f" }} MB/s
	Avg Req Time: {{.ReqTime | printf "%.2f" }} ms
	Fastest Request: {{.MinRequestTime | printf "%.4f" }} ms
	Slowest Request:: {{.MaxRequestTime | printf "%.4f" }} ms

	Number of Request Errors: {{.NumRequestErrs}}
	Number of Connect Errors: {{.NumConnectErrs}}
	Number of Timeout Errors: {{.NumTimeoutErrs}}
	----------------------------------------------
	`
	type Stout struct {
		NumRequests, NumRequestErrs, NumConnectErrs, NumTimeoutErrs int
		AvgThreadDurSec, TotalSize, ReqRate, BytesRate              float64
		AvgTime, ReqTime, MaxRequestTime, MinRequestTime            float64
		Status                                                      []templStatus
	}

	outt := Stout{
		out.NumRequests,
		out.NumRequestErrs,
		out.NumConnectErrs,
		out.NumTimeoutErrs,
		out.TotDuration.Seconds() / float64(responders),
		float64(out.TotRespSize / 1024 / 1024),
		float64(out.NumRequests) / (out.TotDuration.Seconds() / float64(responders)),
		float64(out.TotRespSize/1024/1024) / (out.TotDuration.Seconds() / float64(responders)),
		float64(out.TotDuration / time.Duration(out.NumRequests)),
		out.TotDuration.Seconds() * 1000 / float64(out.NumRequests),
		float64(out.MaxRequestTime / time.Millisecond),
		float64(out.MinRequestTime / time.Millisecond),
		outStatus(out.Status, out.NumRequests)}

	tmpl := template.Must(template.New("wrkout").Parse(outtemplate))
	err := tmpl.Execute(os.Stdout, outt)
	if err != nil {
		fmt.Println("template execution :", err)
	}
}
