package server

import (
	"errors"
	"sync"
	"time"

	"github.com/juju/ratelimit"
)

type rateLimiter struct {
	buckets      map[string]*bucket
	fillInterval time.Duration
	limit        int64
	mu           *sync.RWMutex
}

type bucket struct {
	lastAccessed time.Time
	*ratelimit.Bucket
}

func NewLimiter() *rateLimiter {
	purgeInterval := 1 * time.Hour
	purgeCheckDuration := 10 * time.Minute

	r := &rateLimiter{
		buckets:      make(map[string]*bucket),
		fillInterval: 60 * time.Second,
		limit:        5,
		mu:           &sync.RWMutex{},
	}

	go func() {
		for _ = range time.Tick(purgeCheckDuration) {
			r.purge(purgeInterval)
		}
	}()

	return r
}

func (r *rateLimiter) purge(purgeInterval time.Duration) {
	r.mu.Lock()
	for address, bucket := range r.buckets {
		if time.Now().Sub(bucket.lastAccessed) >= purgeInterval {
			delete(r.buckets, address)
		}
	}
	r.mu.Unlock()
}

func (r *rateLimiter) addIP(addr string) *bucket {
	r.mu.Lock()
	bucket := bucket{
		time.Now(),
		ratelimit.NewBucketWithQuantum(r.fillInterval, r.limit, r.limit),
	}
	r.buckets[addr] = &bucket
	r.mu.Unlock()
	return &bucket
}

func (r *rateLimiter) getIP(addr string) *bucket {
	r.mu.Lock()
	bucket, ok := r.buckets[addr]
	if ok {
		bucket.lastAccessed = time.Now()
		r.mu.Unlock()
		return bucket
	}

	r.mu.Unlock()
	return r.addIP(addr)
}

func (r *rateLimiter) delIP(addr string) {
	r.mu.Lock()
	delete(r.buckets, addr)
	r.mu.Unlock()
}

func (r *rateLimiter) getLastAccessed(addr string) (time.Time, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.buckets[addr]
	if ok {
		return b.lastAccessed, nil
	} else {
		return time.Time{}, errors.New("Address not found")
	}
}
