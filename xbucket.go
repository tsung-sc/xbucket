package xbucket

import (
	"context"
	"errors"
	"sync"
	"time"
)

var bucketMap = new(sync.Map)
var wait = new(sync.WaitGroup)

type bucket struct {
	limit   int64
	mu      *sync.RWMutex
	cancel  context.CancelFunc
	storage chan string
}

// check if can do next step
func (b *bucket) Allow() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	cache := int64(len(b.storage))
	if cache == b.limit {
		return false
	}
	return true
}

// put a data into current xbucket channel,if current channel is full,will return false,otherwise return ture
func (b *bucket) Store() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.storage <- ""
	return true
}

// register a new xbucket with id,if this id is exist,will return a new one
func NewBucket(id string, limit int, releasePeriod time.Duration) *bucket {
	bucketInterface, exist := bucketMap.Load(id)
	if exist {
		bucket, _ := bucketInterface.(*bucket)
		bucket.cancel()
	}
	newBucket := new(bucket)
	newBucket.storage = make(chan string, limit)
	ctx, cancel := context.WithCancel(context.Background())
	newBucket.cancel = cancel
	newBucket.mu = &sync.RWMutex{}
	newBucket.limit = int64(limit)
	bucketMap.Store(id, newBucket)
	wait.Add(1)
	go func(xCtx context.Context, bucket *bucket) {
		wait.Done()
		ticker := time.NewTicker(releasePeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				length := len(bucket.storage)
				bucket.mu.Lock()
				for i := 0; i < length; i++ {
					<-bucket.storage
				}
				bucket.mu.Unlock()
			case <-xCtx.Done():
				close(bucket.storage)
				return
			}
		}
	}(ctx, newBucket)
	wait.Wait()
	return newBucket
}

// get current xbucket with this id. if xbucket is not exist,will return false
func GetBucket(id string) (*bucket, error) {
	bucketInterface, exist := bucketMap.Load(id)
	if !exist {
		return nil, errors.New("xbucket not exist")
	}
	bucket, ok := bucketInterface.(*bucket)
	if !ok {
		return nil, errors.New("not *Bucket")
	}
	return bucket, nil
}
