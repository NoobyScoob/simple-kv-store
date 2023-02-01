package main

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

const ParallelThreads = 20

func TestStartServer(t *testing.T) {
	go startServer()
	fmt.Printf("Starting tcp server...\n")
	time.Sleep(time.Duration(500) * time.Millisecond)
}

func TestTcpServerStart(t *testing.T) {
	_, err := net.Dial("tcp", ADDR)
    if err != nil {
        t.Fatalf("Could not connect to server: %v", err)
    }
}

func TestMemcacheCompaitability(t *testing.T) {
	mc := memcache.New(ADDR)
	key := randString(5)
	value := randString(10)
	err := mc.Set(&memcache.Item{Key: key, Value: []byte(value)})
	if err != nil {
		t.Fatalf("%v", err)
	}
	mc2 := memcache.New(ADDR)
	item, e := mc2.Get(key)
	if e != nil {
		t.Fatalf("%v", e)
	}
	if string(item.Value) != value {
		t.Fatalf("Value Error: %s", item.Value)
	}
}

func TestConcurrentSet(t *testing.T) {
	// errs := make(chan error, 1)
	var wg sync.WaitGroup
	errs := make(chan error, 1)

	for i := 0; i < ParallelThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mc := memcache.New(ADDR)
			key := randString(5)
			value := randString(10)
			err := mc.Set(&memcache.Item{Key: key, Value: []byte(value)})
			if err != nil {
				errs <- err
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestConcurrentGet(t *testing.T) {
	var wg1, wg2 sync.WaitGroup
	keys := make(chan string, 1)

	for i := 0; i < ParallelThreads; i++ {
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			mc := memcache.New(ADDR)
			key := randString(5)
			value := randString(10)
			err := mc.Set(&memcache.Item{Key: key, Value: []byte(value)})
			if err != nil {
				fmt.Println(err)
			} else {
				keys <- key
			}
		}()
	}

	for i := 0; i < ParallelThreads; i++ {
		key := <- keys
		wg2.Add(1)
		go func(key string) {
			defer wg2.Done()
			mc := memcache.New(ADDR)
			_, err := mc.Get(key)
			if err != nil {
				fmt.Println(err)
			}
		}(key)
	}

	wg1.Wait()
	close(keys)
	wg2.Wait()
}

func randString(size int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, size)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}