package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
)

func r(wg *sync.WaitGroup, m map[string]int) {
	for i := 0; i < 100000; i++ {
		json.Marshal(m)
		/*
			    v := m["a"]
					if i%10000 == 0 {
						fmt.Println("a:", v)
					}
		*/
		//time.Sleep(1 * time.Microsecond)
	}
	wg.Done()
}

func w(wg *sync.WaitGroup, m map[string]int, f string) {
	for i := 0; i < 100000; i++ {
		m[f] = i
		if i%10000 == 0 {
			fmt.Printf("%s: %d\n", f, m[f])
		}
		//time.Sleep(1 * time.Microsecond)
	}
	wg.Done()
}

func main() {
	runtime.GOMAXPROCS(2)
	fmt.Println("Hello, playground")
	x := map[string]int{"a": 1}
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go r(wg, x)
	go w(wg, x, "b")
	go w(wg, x, "c")
	wg.Wait()
}
