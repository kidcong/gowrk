# gowrk

## usage

```
gowrk [flags] [url]
```

with the flags being
```
  -c    Connections to keep open (default 10)
  -d    Duration of test in seconds (default 10)
  -t    Number of threads to use (default 10)
  -x    proxy host[:port] Use this proxy
  
  -F	Run Wrk From Json config file  (default "/etc/request.json")
  -H    Request header  
  -M    HTTP method (default "GET")
  -T    Socket/request timeout in ms (default 5000)
  
  -h	help
```
## example

gowrk [flags] url

```go
gowrk -c 10 -d 10 -x 127.0.0.1:80 http://www.test.com/
```
gowrk -F

```go
gowrk -c 10 -d 10 -x 127.0.0.1:80 -F
```

JSON格式：
```json
[
    {
        "URL": "http://www.test.com/",
        "Method": "GET",
        "Header": {
          "X-debug" : "1"
        }
      },
      {
        "URL": "http://www.test2.com/Content/jpimg/logo.png",
        "Method": "GET",
        "Header": {
          "TT" : "A"
        }
      }
]

````

## example output
```
Go-Wrk Version: 0.1  Running Test Begin ...
Go-Wrk Threads: 2    Connections: 10    Duration: 10 s   Timeout: 5000 ms

	------------------ wrk out -------------------
	
	4162 requests in 10.02 s, total 2.00 MB read
	 
	Status: 200  num: 4162  ratio: 100.00%

	Requests/sec: 415.44
	Transfer/sec: 24070617.00 MB/s
	Avg Req Time: 24.07 ms
	Fastest Request: 1.0000 ms
	Slowest Request:: 173.0000 ms

	Number of Request Errors: 0
	Number of Connect Errors: 0
	Number of Timeout Errors: 0
	----------------------------------------------

```
