package datastore

import "time"

type entry struct {
	val      []byte
	expireAt *time.Time
}

type KV struct {
	m map[string]entry
}

func NewKV() *KV {
	return &KV{
		m: make(map[string]entry),
	}
}

func (kv *KV) Set(key string, val []byte, ttl time.Duration) {
	e := entry{
		val: val,
	}
	if ttl > 0 {
		expireAt := time.Now().Add(ttl)
		e.expireAt = &expireAt
	}
	kv.m[key] = e
}

func (kv *KV) Get(key string) ([]byte, bool) {
	e, ok := kv.m[key]
	if !ok {
		return nil, false
	}

	if e.expireAt != nil && time.Now().After(*e.expireAt) {
		delete(kv.m, key)
		return nil, false
	}

	return e.val, true
}

func (kv *KV) TTL(key string) int64 {
	e, ok := kv.m[key]
	if !ok {
		return -2
	}

	// No Expiry
	if e.expireAt == nil {
		return -1
	}

	// Expired
	if time.Now().After(*e.expireAt) {
		delete(kv.m, key)
		return -2
	}

	sec := time.Until(*e.expireAt).Seconds()
	if sec < 0 {
		return -2
	}
	return int64(sec)
}

func (kv *KV) PTTL(key string) int64 {
	e, ok := kv.m[key]
	if !ok {
		return -2
	}

	// No Expiry
	if e.expireAt == nil {
		return -1
	}

	// Expired
	if time.Now().After(*e.expireAt) {
		delete(kv.m, key)
		return -2
	}

	milisec := time.Until(*e.expireAt).Milliseconds()
	if milisec < 0 {
		return -2
	}
	return int64(milisec)
}

func (kv *KV) Expire(key string, seconds int64) bool {
	v, ok := kv.m[key]
	if !ok {
		return false
	}
	expireAt := time.Now().Add(time.Duration(seconds) * time.Second)
	v.expireAt = &expireAt
	kv.m[key] = v
	return true
}

func (kv *KV) PExpire(key string, ms int64) bool {
	v, ok := kv.m[key]
	if !ok {
		return false
	}
	expireAt := time.Now().Add(time.Duration(ms) * time.Millisecond)
	v.expireAt = &expireAt
	kv.m[key] = v

	return true
}

func (kv *KV) Persist(key string) bool {
	v, ok := kv.m[key]
	if !ok || v.expireAt == nil {
		return false
	}

	v.expireAt = nil
	kv.m[key] = v

	return true
}
