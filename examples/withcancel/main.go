package main

import (
	"log"
	"net/http"
	"time"

	"github.com/myself659/tracehelper"
)

func httpreq(method string, url string) {
	_, tracecancel := tracehelper.WithCancel("baidu")
	defer tracecancel()
	req, _ := http.NewRequest(method, url, nil)

	_, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	go httpreq("GET", "https://www.google.com")

	<-time.After(30 * time.Second)
}
