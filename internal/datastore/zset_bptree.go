package datastore

import (
	"errors"
	"fmt"
	"strconv"
)

type entryZSetBPTree struct {
	dict map[string]float64 // member -> score
	tree *bptree
}

func (s *Datastore) ensureZSet(key string) (*entryZSetBPTree, error) {
	e, ok := s.m[key]
	if ok && !s.isExpired(e) {
		if zset, ok := e.val.(*entryZSetBPTree); ok {
			return zset, nil
		}
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	newZSet := &entryZSetBPTree{
		dict: make(map[string]float64),
		tree: newBPTree(),
	}
	s.m[key] = Entry{val: newZSet}
	return newZSet, nil
}

// ZADD
func (s *Datastore) ZADD(key string, value []string) (int, error) {
	ent, err := s.ensureZSet(key)
	if err != nil {
		return 0, err
	}

	count := 0
	for i := 0; i < len(value); i += 2 {
		member := value[i+1]
		score, err := strconv.ParseFloat(value[i], 64)
		if err != nil {
			return count, errors.New("Score must be floating point number")
		}

		if oldScore, ok := ent.dict[member]; ok {
			if oldScore == score {
				continue
			}
			ent.tree.delete(bptKey{score: oldScore, member: member})
		}

		ent.dict[member] = score
		ent.tree.insert(bptKey{score: score, member: member})
		count++
	}

	return count, nil
}

// ZSCORE
func (s *Datastore) ZScore(key, member string) (float64, bool, error) {
	e, ok := s.getEntry(key)
	if !ok {
		return 0, false, nil
	}
	zset, ok := e.val.(*entryZSetBPTree)
	if !ok {
		return 0, false, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	score, found := zset.dict[member]
	return score, found, nil
}

// ZCARD
func (s *Datastore) ZCard(key string) (int, error) {
	e, ok := s.getEntry(key)
	if !ok {
		return 0, nil
	}
	zset, ok := e.val.(*entryZSetBPTree)
	if !ok {
		return 0, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return len(zset.dict), nil
}

// ZRANK
func (s *Datastore) ZRank(key, member string) (int, bool, error) {
	e, ok := s.getEntry(key)
	if !ok {
		return -1, false, nil
	}
	zset, ok := e.val.(*entryZSetBPTree)
	if !ok {
		return -1, false, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	score, exists := zset.dict[member]
	if !exists {
		return -1, false, nil
	}
	r := zset.tree.rankOf(bptKey{score: score, member: member})
	if r < 0 {
		return -1, false, nil
	}
	return r, true, nil
}

// ZRANGE
func (s *Datastore) ZRange(key string, start, stop int) ([]string, error) {
	e, ok := s.getEntry(key)
	if !ok {
		return nil, nil
	}
	zset, ok := e.val.(*entryZSetBPTree)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	bptKeys := zset.tree.rangeByRank(start, stop)
	rs := make([]string, 0, len(bptKeys))
	for _, k := range bptKeys {
		rs = append(rs, k.member)
	}
	return rs, nil
}

// ZRANGE WITHSCORES
func (s *Datastore) ZRangeWithScore(key string, start, stop int) ([]string, error) {
	e, ok := s.getEntry(key)
	if !ok {
		return nil, nil
	}
	zset, ok := e.val.(*entryZSetBPTree)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	bptKeys := zset.tree.rangeByRank(start, stop)
	rs := make([]string, 0, len(bptKeys)*2)
	for _, k := range bptKeys {
		rs = append(rs, k.member)
		rs = append(rs, fmt.Sprintf("%.2f", k.score))
	}
	return rs, nil
}

// ZREM
func (s *Datastore) ZRem(key string, members []string) (int, error) {
	e, ok := s.getEntry(key)
	if !ok {
		return 0, nil
	}
	zset, ok := e.val.(*entryZSetBPTree)
	if !ok {
		return 0, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	countDeleted := 0
	for _, member := range members {
		score, exists := zset.dict[member]
		if !exists {
			continue
		}
		deleted := zset.tree.delete(bptKey{score: score, member: member})
		if deleted {
			delete(zset.dict, member)
			countDeleted++
		}
		if len(zset.dict) == 0 {
			delete(s.m, key)
		}
	}

	return countDeleted, nil
}
