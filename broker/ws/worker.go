package ws

import "log"

var (
	works chan func()
)

// InitWorker init workers.
func InitWorker(count int, bufSize int) {
	works = make(chan func(), bufSize)
	for i := 0; i < count; i++ {
		go worker(i, works)
	}
}

func worker(id int, works <-chan func()) {
	log.Printf("worker#%d started.", id)
	for work := range works {
		work()
	}
}
