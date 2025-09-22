package datastore

import (
	"errors"
	"math"
	"sync"
	"sync/atomic"

	"github.com/spaolacci/murmur3"
)

// entryCMS implements the Count-Min Sketch for one "key".
type entryCMS struct {
	width   uint32
	depth   uint32
	counter [][]uint32
}

// CMS manages multiple Count-Min Sketch instances (like RedisBloom CMS key -> entryCMS).
type CMS struct {
	mu sync.RWMutex
	m  map[string]*entryCMS
}

// NewCMS initializes a new CMS container.
func NewCMS() *CMS {
	return &CMS{
		m: make(map[string]*entryCMS),
	}
}

// CreateEntryCMS initializes a new Count-Min Sketch with given width and depth.
func CreateEntryCMS(w uint32, d uint32) *entryCMS {
	e := &entryCMS{
		width: w,
		depth: d,
	}
	e.counter = make([][]uint32, d)
	for i := uint32(0); i < d; i++ {
		e.counter[i] = make([]uint32, w)
	}
	return e
}

// CalcCMSDim calculates the width and depth for CMS
// given error rate (ε) and error probability (δ).
//
// ε = error rate (e.g., 0.01)
// δ = probability of error exceeding ε (e.g., 1e-5)
//
// width = ceil(e / ε)
// depth = ceil(ln(1/δ))
func CalcCMSDim(errRate float64, errProb float64) (uint32, uint32) {
	w := uint32(math.Ceil(math.E / errRate))
	d := uint32(math.Ceil(math.Log(1.0 / errProb)))
	return w, d
}

// calcHash generates a hash value for a string with a given seed.
func calcHash(item string, seed uint32) uint32 {
	return murmur3.Sum32WithSeed([]byte(item), seed)
}

// EnsureKey ensures a key has an entryCMS (like CMS.INITBYDIM in RedisBloom).
func (c *CMS) CreateCMS(key string, w, d uint32) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.m[key]; exists {
		return false
	}

	c.m[key] = CreateEntryCMS(w, d)
	return true
}

func (c *CMS) CreateCMSByProb(key string, errRate float64, errProb float64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	w, d := CalcCMSDim(errRate, errProb)
	if _, exists := c.m[key]; exists {
		return false
	}

	c.m[key] = CreateEntryCMS(w, d)
	return true
}

// IncrBy increments the estimated count of an item by 'value'.
// It returns the new estimated count after increment.
func (c *CMS) IncrBy(key, item string, value uint32) (uint32, error) {
	c.mu.RLock()
	entry, ok := c.m[key]
	c.mu.RUnlock()
	if !ok {
		return 0, errors.New("CMS: key does not exist")
	}

	minCount := ^uint32(0) // Max uint32

	for i := uint32(0); i < entry.depth; i++ {
		hash := calcHash(item, i)
		j := hash % entry.width

		// concurrency-safe update
		newVal := atomic.AddUint32(&entry.counter[i][j], value)

		// track minimum
		if newVal < minCount {
			minCount = newVal
		}
	}
	return minCount, nil
}

// Count estimates the frequency of an item.
func (c *CMS) Count(key, item string) uint32 {
	c.mu.RLock()
	entry, ok := c.m[key]
	c.mu.RUnlock()
	if !ok {
		return 0
	}

	minCount := ^uint32(0) // Max uint32

	for i := uint32(0); i < entry.depth; i++ {
		hash := calcHash(item, i)
		j := hash % entry.width
		val := atomic.LoadUint32(&entry.counter[i][j])
		if val < minCount {
			minCount = val
		}
	}
	return minCount
}

func (c *CMS) Query(key string, listItem []string) ([]uint32, error) {
	c.mu.RLock()
	entry, ok := c.m[key]
	c.mu.RUnlock()
	if !ok {
		return nil, errors.New("CMS: key does not exist")
	}

	res := []uint32{}
	for _, item := range listItem {
		minCount := ^uint32(0) // Max uint32
		for i := uint32(0); i < entry.depth; i++ {
			hash := calcHash(item, i)
			j := hash % entry.width
			val := atomic.LoadUint32(&entry.counter[i][j])
			if val < minCount {
				minCount = val
			}
		}
		res = append(res, minCount)
	}

	return res, nil
}

// Reset clears all counters for a given key.
func (c *CMS) Reset(key string) {
	c.mu.RLock()
	entry, ok := c.m[key]
	c.mu.RUnlock()
	if !ok {
		return
	}

	for i := range entry.counter {
		for j := range entry.counter[i] {
			atomic.StoreUint32(&entry.counter[i][j], 0)
		}
	}
}

// Merge combines another entryCMS into this one.
// Both must have the same width and depth.
func (c *CMS) Merge(key string, other *entryCMS) {
	c.mu.RLock()
	entry, ok := c.m[key]
	c.mu.RUnlock()
	if !ok {
		panic("entryCMS does not exist for key: " + key)
	}

	if entry.width != other.width || entry.depth != other.depth {
		panic("CMS dimensions do not match")
	}

	for i := uint32(0); i < entry.depth; i++ {
		for j := uint32(0); j < entry.width; j++ {
			ov := atomic.LoadUint32(&other.counter[i][j])
			if ov > 0 {
				atomic.AddUint32(&entry.counter[i][j], ov)
			}
		}
	}
}

func (c *CMS) Info(key string) (uint32, uint32, error) {
	c.mu.RLock()
	entry, ok := c.m[key]
	c.mu.RUnlock()
	if !ok {
		return 0, 0, errors.New("CMS: key does not exist")
	}

	return entry.width, entry.depth, nil
}
