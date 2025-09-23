package datastore

import (
	"backend/internal/config"
	"time"
)

func (s *Datastore) Set(key, val string, ttl time.Duration) {
	e := Entry{val: val}
	if ttl > 0 {
		expireAt := time.Now().Add(ttl)
		e.expireAt = &expireAt
	}
	s.m[key] = e
}

func (s *Datastore) Get(key string) (string, bool, error) {
	e, ok := s.getEntry(key)
	if !ok {
		return "", false, nil
	}

	if data, ok := e.val.(string); ok {
		return data, true, nil
	}
	return "", false, config.ErrWrongType
}

func (s *Datastore) TTL(key string) int64 {
	e, ok := s.getEntry(key)
	if !ok {
		return -2
	}

	if e.expireAt == nil {
		return -1
	}
	return int64(time.Until(*e.expireAt).Seconds())
}

func (s *Datastore) PTTL(key string) int64 {
	e, ok := s.getEntry(key)
	if !ok {
		return -2
	}

	if e.expireAt == nil {
		return -1
	}
	return time.Until(*e.expireAt).Milliseconds()
}

func (s *Datastore) Expire(key string, seconds int64) bool {
	e, ok := s.getEntry(key)
	if !ok {
		return false
	}
	expireAt := time.Now().Add(time.Duration(seconds) * time.Second)
	e.expireAt = &expireAt
	s.m[key] = e
	return true
}

func (s *Datastore) PExpire(key string, ms int64) bool {
	e, ok := s.getEntry(key)
	if !ok {
		return false
	}
	expireAt := time.Now().Add(time.Duration(ms) * time.Millisecond)
	e.expireAt = &expireAt
	s.m[key] = e
	return true
}

func (s *Datastore) Persist(key string) bool {
	e, ok := s.getEntry(key)
	if !ok || e.expireAt == nil {
		return false
	}
	e.expireAt = nil
	s.m[key] = e
	return true
}

func (s *Datastore) Exists(keys []string) int {
	count := 0
	for _, key := range keys {
		if _, ok := s.getEntry(key); ok {
			count++
		}
	}
	return count
}

func (s *Datastore) Del(keys []string) int {
	count := 0
	for _, key := range keys {
		if _, ok := s.getEntry(key); ok {
			delete(s.m, key)
			count++
		}
	}
	return count
}
