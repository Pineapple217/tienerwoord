package main

import (
	"bytes"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/brianvoe/gofakeit"
)

const (
	apiURL      = "https://vrtnws-api.vrt.be/nwsnwsnws/teenager-word-election/_vote"
	contentType = "application/json"
)

func main() {
	var (
		totalVotes  int64
		delta       int64
		errors      int64
		errorsTotal int64
		wg          sync.WaitGroup
	)

	word := flag.String("word", "pookie", "word to vote")
	workers := flag.Int("c", 5, "worker count")
	deltaS := flag.Int("s", 1, "delta in seconds for log")
	flag.Parse()
	wordPayload := fmt.Sprintf(`{"word":"%s"}`, *word)

	slog.Info("running!", "word", *word, "delta_s", *deltaS)

	done := make(chan bool)

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{}
			for {
				select {
				case <-done:
					return
				default:
					req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer([]byte(wordPayload)))
					if err != nil {
						atomic.AddInt64(&errors, 1)
						atomic.AddInt64(&errorsTotal, 1)
						continue
					}
					req.Header.Set("Content-Type", contentType)
					req.Header.Set("User-Agent", gofakeit.UserAgent())
					req.Header.Set("Accept-Encoding", "gzip")
					req.Header.Set("Cache-Control", "no-cache")
					req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")

					resp, err := client.Do(req)
					if err != nil {
						atomic.AddInt64(&errors, 1)
						atomic.AddInt64(&errorsTotal, 1)
						continue
					}
					resp.Body.Close()

					if resp.StatusCode == http.StatusOK {
						atomic.AddInt64(&totalVotes, 1)
						atomic.AddInt64(&delta, 1)
					} else {
						atomic.AddInt64(&errors, 1)
						atomic.AddInt64(&errorsTotal, 1)
					}
				}
			}
		}()
	}

	go func() {
		d := *deltaS
		for {
			time.Sleep(time.Duration(d) * time.Millisecond * 1000)
			errRate := math.Floor(float64(errorsTotal)/float64(totalVotes+errorsTotal)*100) / 100
			slog.Info("vote stats", "votes", delta, "errors", errors, "total", totalVotes, "error rate", errRate)
			atomic.SwapInt64(&delta, 0)
			atomic.SwapInt64(&errors, 0)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	slog.Info("Received an interrupt signal, exiting...")
	done <- true
	slog.Info("Exiting... Goodbye!")
}
