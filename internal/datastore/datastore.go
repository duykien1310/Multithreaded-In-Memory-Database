package datastore

import "time"

type Entry struct {
	val      any
	expireAt *time.Time
}

type Datastore struct {
	m map[string]Entry
}

func NewDataStore() *Datastore {
	return &Datastore{
		m: make(map[string]Entry),
	}
}

func (s *Datastore) isExpired(e Entry) bool {
	return e.expireAt != nil && time.Now().After(*e.expireAt)
}

func (s *Datastore) getEntry(key string) (Entry, bool) {
	e, ok := s.m[key]
	if !ok {
		return Entry{}, false
	}
	if s.isExpired(e) {
		delete(s.m, key)
		return Entry{}, false
	}
	return e, true
}
