package datastore

type entry struct {
	val []byte
}

type KV struct {
	m map[string]entry
}

func NewKV() *KV {
	return &KV{
		m: make(map[string]entry),
	}
}

func (kv *KV) Set(key string, val []byte) {
	e := entry{
		val: val,
	}
	kv.m[key] = e
}

func (kv *KV) Get(key string) ([]byte, bool) {
	e, ok := kv.m[key]
	if !ok {
		return nil, false
	}

	return e.val, true
}
