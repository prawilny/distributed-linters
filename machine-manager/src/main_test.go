package main

import (
	"context"
	"testing"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"
)

func TestEtcd(t *testing.T) {
	dialTimeout := 2 * time.Second
	requestTimeout := 10 * time.Second

	ctx, cancelCtx := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelCtx()
	cli, err := clientV3.New(clientV3.Config{
		DialTimeout: dialTimeout,
		Endpoints:   []string{"etcd:2379"},
	})
	if err != nil {
		t.Errorf("error creating etcd connection: %s", err.Error())
	}
	defer cli.Close()
	kv := clientV3.NewKV(cli)

	key := "key"
	defer kv.Delete(ctx, key, clientV3.WithPrefix())

	value1 := "xD"
	value2 := "xDD"
	value3 := "xDDD"

	_, err = kv.Put(ctx, key, value1)
	if err != nil {
		t.Errorf("error on put: %s", err.Error())
	}

	txnResponse, err := kv.Txn(ctx).If(
		clientV3.Compare(clientV3.Value(key), "=", value2),
	).Then(
		clientV3.OpPut(key, value3),
	).Commit()

	if err != nil {
		t.Errorf("unexpected conditional put error: %s", err.Error())
	}
	if txnResponse.Succeeded {
		t.Errorf("conditional put shouldn't have succeeded")
	}

	_, err = kv.Put(ctx, key, value2)
	if err != nil {
		t.Errorf("error on put: %s", err.Error())
	}

	txnResponse, err = kv.Txn(ctx).If(
		clientV3.Compare(clientV3.Value(key), "=", value2),
	).Then(
		clientV3.OpPut(key, value3),
	).Commit()
	// response, err = kv.Do(ctx, conditionalPut)
	if err != nil {
		t.Errorf("unexpected conditional put error: %s", err.Error())
	}
	if !txnResponse.Succeeded {
		t.Errorf("conditional put should have succeeded")
	}
}
