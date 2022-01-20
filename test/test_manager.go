package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func run(name string, arg ...string) (string, string) {
	cmd := exec.Command(name, arg...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		panic("cmd failed with " + err.Error())
	}
	return string(stdout.Bytes()), string(stderr.Bytes())
}

const (
	LANGUAGE = "python"
	VERSION  = "1.0"
)

var (
	LOAD_BALANCER = os.Getenv("LOAD_BALANCER")
	MANAGER       = os.Getenv("MANAGER")
	REGISTRY      = os.Getenv("REGISTRY")
	LINT_URL      = "http://" + LOAD_BALANCER + "/lint?lang=python"
)

func wait(t time.Duration) {
	fmt.Printf("sleeping for %v \n", t)
	time.Sleep(t)
}

func main() {
	fmt.Println("Trying to lint")
	resp, err := http.Post(LINT_URL, "text/plain", strings.NewReader("body"))
	if err != nil {
		panic(err)
	}
	if resp.StatusCode/100 != 5 {
		panic("loadbalancer shouldn't have been available" + resp.Status)
	}
	fmt.Println("OK: no linters available")

	fmt.Println("Creating a linter version")
	run("./admin", MANAGER, "add_version", "-image_url="+REGISTRY+"/"+LANGUAGE+"_linter", "-language="+LANGUAGE, "-version="+VERSION)
	wait(30 * time.Second)
	fmt.Println("OK: linter created")

	fmt.Println("Checking the list of available linter versions")
	_, stderr := run("./admin", MANAGER, "list_versions")
	if !strings.Contains(stderr, "[]") {
		panic("version list should have been empty")
	}
	fmt.Println("OK: no weights are set")

	fmt.Println("Setting the proportions of requests forwarded to linter versions")
	run("./admin", MANAGER, "set_proportions", LANGUAGE, VERSION, "1.0")
	fmt.Println("Checking the list of available linter versions")
	_, stderr = run("./admin", MANAGER, "list_versions")
	if !strings.Contains(stderr, "["+VERSION+"]") {
		panic("version list should have contained " + VERSION)
	}
	fmt.Println("OK: proportions set")

	fmt.Println("Trying to lint")
	resp, err = http.Post(LINT_URL, "text/plain", strings.NewReader("x =23"))
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic("loadbalancer should have been available")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if !strings.Contains(string(body), "not surrounded by whitespace") {
		panic("the service should've reported error")
	}
	fmt.Println("OK: linter linted the code")

	fmt.Println("Zeroing proportions of requests forwarded to linter versions")
	run("./admin", MANAGER, "set_proportions", LANGUAGE, VERSION, "0.0")
	run("./admin", MANAGER, "remove_version", "-language="+LANGUAGE, "-version="+VERSION)

	_, stderr = run("./admin", MANAGER, "list_versions")
	if !strings.Contains(stderr, "[]") {
		panic("version list should have been empty")
	}
	fmt.Println("OK: linter removed")

	fmt.Println("PASS")
}
