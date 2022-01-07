package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
)

var (
	endpoint   = "http://manager:80"
	get_path   = "/get"
	bump_path  = "/bump"
	reset_path = "/reset"
)

func request(t *testing.T, url string) map[string]string {
	resp, err := http.Get(url)
	if err != nil {
		t.Error(err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return nil
	}
	data := make(map[string]string)
	json.Unmarshal(body, &data)

	return data
}

func TestMachineManager(t *testing.T) {
	request(t, endpoint+reset_path)

	get1 := request(t, endpoint+get_path)
	assert.Equal(t, strconv.FormatInt(0, 10), get1["value"])

	bump1 := request(t, endpoint+bump_path)
	_, failed := bump1["error"]
	assert.Equal(t, false, failed)

	get2 := request(t, endpoint+get_path)
	assert.Equal(t, strconv.FormatInt(1, 10), get2["value"])
}

func bump(t *testing.T, ch chan<- bool) {
	bmp := request(t, endpoint+bump_path)
	failed := false
	if bmp["error"] != "" {
		failed = true
	}
	ch <- failed
}

func TestConditionalUpdate(t *testing.T) {
	request(t, endpoint+reset_path)

	n := 10
	ch := make(chan bool, n)

	for i := 0; i < n; i++ {
		go bump(t, ch)
	}

	errors := 0
	for i := 0; i < n; i++ {
		if <-ch {
			errors += 1
		}
	}
	assert.NotEqual(t, 0, errors)
}
