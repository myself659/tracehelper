package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/myself659/tracehelper"
)

func httpreq(method string, url string) {
	filter := func() bool {
		if strings.Contains(url, "https") {

			return true
		}
		return false
	}

	stoptrace := tracehelper.WithFilter("ipds", filter)
	defer stoptrace()

	//starttrace() //  call twice but do once
	req, _ := http.NewRequest(method, url, nil)

	_, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	httpreq("GET", "http://blog.ipds.top/")
	go httpreq("GET", "https://blog.ipds.top/")
	<-time.After(30 * time.Second)
}
