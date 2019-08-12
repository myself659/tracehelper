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
	//starttrace() // even call twice but only do once
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
