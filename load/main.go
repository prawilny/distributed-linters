package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
	"fmt"
	"flag"
)

var (
	target_addr = flag.String("target-addr", "", "The target loadbalancer data port")
)

func main() {
    flag.Parse()
    if len(*target_addr) == 0 {
        log.Fatalf("No -target-addr given\n")
    }
	mut := sync.Mutex{}
	counter := 0
	tr := &http.Transport{
		MaxConnsPerHost: 50,
	}
	client := &http.Client{Transport: tr}
	ticker := time.NewTicker(time.Second)
	go func() {
		for _ = range ticker.C {
			mut.Lock()
			x := counter
			counter = 0
			mut.Unlock()
			log.Printf("%d\n", x)
		}
	}()
    query := fmt.Sprintf("http://%s/lint?lang=python", *target_addr)
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
				mut.Lock()
				counter += 1
				mut.Unlock()
			}
		}()
	}
	wg.Wait()
}
