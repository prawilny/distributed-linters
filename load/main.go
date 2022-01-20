package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const PORT = 40000

var (
	lintCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "load_lints",
		Help: "The total number of lints performed by the program so far",
	})
	LOAD_BALANCER_HOST = os.Getenv("LOADBALANCER_SERVICE_HOST")
	LOAD_BALANCER_PORT = os.Getenv("LOADBALANCER_SERVICE_PORT_DATA")
)

func main() {
	tr := &http.Transport{
		MaxConnsPerHost: 50,
	}
	client := &http.Client{Transport: tr}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(res http.ResponseWriter, req *http.Request) {
		promhttp.Handler().ServeHTTP(res, req)
	})
	s := http.Server{
		Addr:    fmt.Sprintf(":%d", PORT),
		Handler: mux,
	}
	go func() {
		log.Fatal(s.ListenAndServe())
	}()

	query := fmt.Sprintf("http://%s:%s/lint?lang=python", LOAD_BALANCER_HOST, LOAD_BALANCER_PORT)
	var wg sync.WaitGroup
	for i := 0; i < 50; i += 1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				req, err := http.NewRequest("POST", query, bytes.NewBuffer([]byte(`a=b
c= d
e =f
g = h`)))
				if err != nil {
					log.Fatalf("Could not create the request %s\n", err)
				}
				resp, err := client.Do(req)
				if err != nil {
					panic(err.Error())
				}
				ioutil.ReadAll(resp.Body)
				lintCounter.Inc()
			}
		}()
	}
	wg.Wait()
}
