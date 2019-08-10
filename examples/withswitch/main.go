package main

import (
	"log"
	"net/http"
	"time"

	"github.com/myself659/tracehelper"
)

func httpreq(method string, url string) {
	_, stoptrace := tracehelper.WithSwitch("ipds", true)
	defer stoptrace()
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
