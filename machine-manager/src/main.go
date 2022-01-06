package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"
)

var (
	dialTimeout    = 2 * time.Second
	requestTimeout = 10 * time.Second
	counterKey     = "counter"
)

type Counter struct {
	etcd *clientV3.Client
}

func (c *Counter) respond(w http.ResponseWriter, statusCode int, params map[string]string) {
	json, _ := json.Marshal(params)
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func (c *Counter) getValue() (string, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()

	get, err := clientV3.NewKV(c.etcd).Get(ctx, counterKey)
	if err != nil {
		return "", err
	}
	kvs := get.Kvs
	value := "0"
	if len(kvs) != 0 {
		value = string(kvs[0].Value)
	}
	return value, nil
}

func (c *Counter) get(w http.ResponseWriter, r *http.Request) {
	value, err := c.getValue()
	if err != nil {
		c.respond(w, 500, map[string]string{
			"error": err.Error(),
		})
		return
	}
	c.respond(w, 200, map[string]string{
		"value": value,
	})
}

func (c *Counter) bump(w http.ResponseWriter, r *http.Request) {
	valueString, err := c.getValue()
	if err != nil {
		c.respond(w, 500, map[string]string{
			"error": err.Error(),
		})
		return
	}
	value, err := strconv.ParseInt(valueString, 10, 8)

	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()

	txn, err := clientV3.NewKV(c.etcd).Txn(ctx).If(
		clientV3.Compare(clientV3.Value(counterKey), "=", valueString),
	).Then(clientV3.OpPut(counterKey, strconv.FormatInt(value+1, 10))).Commit()
	if err != nil {
		c.respond(w, 500, map[string]string{
			"error": err.Error(),
		})
		return
	}
	params := make(map[string]string)
	if !txn.Succeeded {
		params["message"] = "bump failed"
	}
	c.respond(w, 200, params)
}

func (c *Counter) Set(value int64) error {
	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()

	_, err := clientV3.NewKV(c.etcd).Put(ctx, counterKey, strconv.FormatInt(value, 10))
	if err != nil {
		return err
	}
	return nil
}

func main() {
	cli, err := clientV3.New(clientV3.Config{
		DialTimeout: dialTimeout,
		Endpoints:   []string{"etcd:2379"},
	})
	if err != nil {
		fmt.Errorf("error creating etcd connection: %s", err.Error())
		os.Exit(1)
	}
	counter := &Counter{etcd: cli}
	if counter.Set(0) != nil {
		os.Exit(2)
	}
	http.HandleFunc("/get", counter.get)
	http.HandleFunc("/bump", counter.bump)
	http.ListenAndServe(":80", nil)
}
