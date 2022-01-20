package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	LANGUAGE = "python"
	N        = 1000
	M        = 50
	V1       = "1-0"
	V2       = "2-0"
)

var (
	LOAD_BALANCER = os.Getenv("LOAD_BALANCER")
	LINTER        = os.Getenv("LINTER")
	PROMETHEUS    = os.Getenv("PROMETHEUS")
	LINT_URL      = "http://" + LOAD_BALANCER + "/lint?lang=" + LANGUAGE
)

type Result struct {
	Value []interface{} `json: value`
}

type Data struct {
	Result []Result `json: result`
}

type Prom struct {
	Data Data `json: data`
}

func lints_for_version(version string) int {
	res, err := http.Get(fmt.Sprintf(`http://%s/api/v1/query?query=sum(python_linter_lints{container="linter-python-%s"})`, PROMETHEUS, version))
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var prom Prom
	json.Unmarshal(body, &prom)
	i, _ := strconv.Atoi(prom.Data.Result[0].Value[1].(string))
	return i
}

func main() {
	finished_chan := make(chan error, N)

	v1_start := lints_for_version(V1)
	v2_start := lints_for_version(V2)

	client := &http.Client{Transport: &http.Transport{
		MaxConnsPerHost: 50,
	}}
	for i := 0; i < M; i++ {
		go func() {
			for j := 0; j < N/M; j++ {
				resp, err := client.Post(LINT_URL, "text/plain", strings.NewReader("x =23"))
				if resp != nil {
					defer resp.Body.Close()
				}
				ioutil.ReadAll(resp.Body)
				finished_chan <- err
			}
		}()
	}

	for i := 0; i < N; i++ {
		err := <-finished_chan
		if err != nil {
			fmt.Println("some error posting:" + err.Error())
		}
	}

	// Waiting for prometheus
	time.Sleep(1500 * time.Millisecond)

	v1_lints := lints_for_version(V1) - v1_start
	v2_lints := lints_for_version(V2) - v2_start

	fmt.Printf("requests to linter %s: %d\nrequests to linter %s: %d\n", V1, v1_lints, V2, v2_lints)
}
