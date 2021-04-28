package main

import (
	"github.com/tsung-sc/xbucket"
	"log"
	"sync/atomic"
	"time"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}
func main() {
	xbucket.NewBucket("first", 10, 20*time.Second)
	time.Sleep(time.Second)
	var count int64
	var successCount int64
	for i := 0; i < 100; i++ {
		bucket, err := xbucket.GetBucket("first")
		if err != nil {
			log.Printf("%+v", err)
			return
		}
		count := atomic.AddInt64(&count, 1)
		go func() {
			if bucket.Allow() {
				successCount := atomic.AddInt64(&successCount, 1)
				log.Println("count: ", count, bucket.Store(), "successCount: ", successCount)
			} else {
				log.Println("count: ", count, "allow false")
			}
		}()
	}
	select {}
}
