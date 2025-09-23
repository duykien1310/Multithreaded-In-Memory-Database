package datastore

import (
	"errors"
	"math"
	"sync/atomic"

	"github.com/spaolacci/murmur3"
)

type entryCMS struct {
	width   uint32
	depth   uint32
	counter [][]uint32
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

func CalcCMSDim(errRate float64, errProb float64) (uint32, uint32) {
	w := uint32(math.Ceil(math.E / errRate))
	d := uint32(math.Ceil(math.Log(1.0 / errProb)))
	return w, d
}

func calcHash(item string, seed uint32) uint32 {
	return murmur3.Sum32WithSeed([]byte(item), seed)
}

func (ds *Datastore) CreateCMS(key string, w, d uint32) (bool, error) {
	if e, ok := ds.m[key]; ok {
		if _, ok := e.val.(*entryCMS); ok {
			return false, nil
		}
		return false, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	ds.m[key] = Entry{val: CreateEntryCMS(w, d)}
	return true, nil
}

func (ds *Datastore) CreateCMSByProb(key string, errRate, errProb float64) (bool, error) {
	w, d := CalcCMSDim(errRate, errProb)
	return ds.CreateCMS(key, w, d)
}

// IncrBy increments the estimated count of an item by 'value'.
func (ds *Datastore) IncrBy(key, item string, value uint32) (uint32, error) {

	e, ok := ds.m[key]
	if !ok {
		return 0, errors.New("CMS: key does not exist")
	}
	cms, ok := e.val.(*entryCMS)
	if !ok {
		return 0, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	minCount := ^uint32(0) // Max uint32
	for i := uint32(0); i < cms.depth; i++ {
		hash := calcHash(item, i)
		j := hash % cms.width
		newVal := atomic.AddUint32(&cms.counter[i][j], value)
		if newVal < minCount {
			minCount = newVal
		}
	}
	return minCount, nil
}

// Count estimates the frequency of an item.
func (ds *Datastore) Count(key, item string) (uint32, error) {

	e, ok := ds.m[key]
	if !ok {
		return 0, errors.New("CMS: key does not exist")
	}
	cms, ok := e.val.(*entryCMS)
	if !ok {
		return 0, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	minCount := ^uint32(0)
	for i := uint32(0); i < cms.depth; i++ {
		hash := calcHash(item, i)
		j := hash % cms.width
		val := atomic.LoadUint32(&cms.counter[i][j])
		if val < minCount {
			minCount = val
		}
	}
	return minCount, nil
}

// Query multiple items
func (ds *Datastore) Query(key string, listItem []string) ([]uint32, error) {

	e, ok := ds.m[key]
	if !ok {
		return nil, errors.New("CMS: key does not exist")
	}
	cms, ok := e.val.(*entryCMS)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	res := make([]uint32, len(listItem))
	for idx, item := range listItem {
		minCount := ^uint32(0)
		for i := uint32(0); i < cms.depth; i++ {
			hash := calcHash(item, i)
			j := hash % cms.width
			val := atomic.LoadUint32(&cms.counter[i][j])
			if val < minCount {
				minCount = val
			}
		}
		res[idx] = minCount
	}
	return res, nil
}

// Reset clears all counters for a given key.
func (ds *Datastore) Reset(key string) error {

	e, ok := ds.m[key]
	if !ok {
		return errors.New("CMS: key does not exist")
	}
	cms, ok := e.val.(*entryCMS)
	if !ok {
		return errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	for i := range cms.counter {
		for j := range cms.counter[i] {
			atomic.StoreUint32(&cms.counter[i][j], 0)
		}
	}
	return nil
}

// Merge combines another entryCMS into this one.
// Both must have the same width and depth.
func (ds *Datastore) Merge(key string, other *entryCMS) error {

	e, ok := ds.m[key]
	if !ok {
		return errors.New("CMS: key does not exist")
	}
	cms, ok := e.val.(*entryCMS)
	if !ok {
		return errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	if cms.width != other.width || cms.depth != other.depth {
		return errors.New("CMS: dimensions do not match")
	}

	for i := uint32(0); i < cms.depth; i++ {
		for j := uint32(0); j < cms.width; j++ {
			ov := atomic.LoadUint32(&other.counter[i][j])
			if ov > 0 {
				atomic.AddUint32(&cms.counter[i][j], ov)
			}
		}
	}
	return nil
}

// Info returns the CMS dimensions.
func (ds *Datastore) Info(key string) (uint32, uint32, error) {

	e, ok := ds.m[key]
	if !ok {
		return 0, 0, errors.New("CMS: key does not exist")
	}
	cms, ok := e.val.(*entryCMS)
	if !ok {
		return 0, 0, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return cms.width, cms.depth, nil
}
