# tracehelper 


[![GoDoc](https://godoc.org/github.com/myself659/tracehelper?status.svg)](https://godoc.org/github.com/myself659/tracehelper)
**make trace convenient for golang** 

Support a variety of different trace modes:

- http
- signal 
- context scope 
- switch control
- filter control 

# API

[Refer to godoc](https://godoc.org/github.com/myself659/tracehelper)

# example  

```
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/myself659/tracehelper"
)

func httpreq(method string, url string) {
	starttrace, stoptrace := tracehelper.WithSwitch("ipds", false)
	defer stoptrace()
	starttrace()
	req, _ := http.NewRequest(method, url, nil)

	_, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	go httpreq("GET", "http://blog.ipds.top/")

	<-time.After(30 * time.Second)
}

```

