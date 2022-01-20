package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

const (
	LANGUAGE = "python"
	N        = 1000
	V1       = "1.0"
	V2       = "2.0"
)

var (
	LOAD_BALANCER = os.Getenv("LOAD_BALANCER")
	LINT_URL      = "http://" + LOAD_BALANCER + "/lint?lang=python"
	LINTER        = os.Getenv("LINTER")
	PROMETHEUS    = os.Getenv("PROMETHEUS")
	client, err   = api.NewClient(api.Config{
		Address: "http://" + PROMETHEUS,
	})
)

func lints_for_version(version string) int {
	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result, _, err := v1api.Query(ctx, "python_linter_lints", time.Now())
	if err != nil {
		panic("error querying: " + err.Error())
	}

	samples := result.(model.Vector)

	lints := 0
	for _, smpl := range samples {
		if label, _ := smpl.Metric["container"]; strings.Contains(string(label), strings.ReplaceAll(version, ".", "-")) {
			lints += int(smpl.Value)
		}
	}

	return lints
}

func main() {
	if err != nil {
		panic("Error connecting to prometheus: " + err.Error())
	}
	v1_start := lints_for_version(V1)
	v2_start := lints_for_version(V2)

	for i := 0; i < N; i++ {
		resp, err := http.Post(LINT_URL, "text/plain", strings.NewReader("x =23"))
		if resp != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			fmt.Println("some error posting:" + err.Error())
		}
	}

	// Waiting for prometheus
	time.Sleep(1100 * time.Millisecond)

	v1_lints := lints_for_version(V1) - v1_start
	v2_lints := lints_for_version(V2) - v2_start

	fmt.Printf("requests to linter %s: %d\nv2 requests to linter %s: %d\n", V1, v1_lints, V2, v2_lints)
}
