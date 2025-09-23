package datastore

import (
	"backend/internal/config"
	"errors"
	"math"
	"sync/atomic"

	"github.com/spaolacci/murmur3"
)

type EntryCMS struct {
	width   uint32
	depth   uint32
	counter [][]uint32
}

// CreateEntryCMS initializes a new Count-Min Sketch with given width and depth.
func CreateEntryCMS(w uint32, d uint32) *EntryCMS {
	e := &EntryCMS{
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

func (s *Datastore) getCMS(key string) (*EntryCMS, error) {
	e, ok := s.getEntry(key)
	if !ok {
		return nil, config.ErrKeyNotExist
	}
	cms, ok := e.val.(*EntryCMS)
	if !ok {
		return nil, config.ErrWrongType
	}
	return cms, nil
}

func (s *Datastore) CreateCMS(key string, w, d uint32) (bool, error) {
	if e, ok := s.m[key]; ok {
		if _, ok := e.val.(*EntryCMS); ok {
			return false, nil
		}
		return false, config.ErrWrongType
	}
	s.m[key] = Entry{val: CreateEntryCMS(w, d)}
	return true, nil
}

func (s *Datastore) CreateCMSByProb(key string, errRate, errProb float64) (bool, error) {
	w, d := CalcCMSDim(errRate, errProb)
	return s.CreateCMS(key, w, d)
}

// IncrBy increments the estimated count of an item by 'value'.
func (s *Datastore) IncrBy(key, item string, value uint32) (uint32, error) {
	cms, err := s.getCMS(key)
	if err != nil {
		return 0, err
	}

	minCount := ^uint32(0)
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
func (s *Datastore) Count(key, item string) (uint32, error) {
	cms, err := s.getCMS(key)
	if err != nil {
		return 0, err
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
func (s *Datastore) Query(key string, listItem []string) ([]uint32, error) {
	cms, err := s.getCMS(key)
	if err != nil {
		return nil, err
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
func (s *Datastore) Reset(key string) error {
	cms, err := s.getCMS(key)
	if err != nil {
		return err
	}

	for i := range cms.counter {
		for j := range cms.counter[i] {
			atomic.StoreUint32(&cms.counter[i][j], 0)
		}
	}
	return nil
}

func (s *Datastore) Merge(key string, other *EntryCMS) error {
	cms, err := s.getCMS(key)
	if err != nil {
		return err
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
func (s *Datastore) Info(key string) (uint32, uint32, error) {
	cms, err := s.getCMS(key)
	if err != nil {
		return 0, 0, err
	}
	return cms.width, cms.depth, nil
}
