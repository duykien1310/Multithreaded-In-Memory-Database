package datastore

import (
	"backend/internal/config"
	"strconv"
)

type EntryZSetBPTree struct {
	dict map[string]float64 // member -> score
	tree *bptree
}

func (s *Datastore) getZSet(key string) (*EntryZSetBPTree, error) {
	e, ok := s.getEntry(key)
	if !ok {
		return nil, nil
	}
	zset, ok := e.val.(*EntryZSetBPTree)
	if !ok {
		return nil, config.ErrWrongType
	}
	return zset, nil
}

func (s *Datastore) ensureZSet(key string) (*EntryZSetBPTree, error) {
	e, ok := s.m[key]
	if ok && !s.isExpired(e) {
		if zset, ok := e.val.(*EntryZSetBPTree); ok {
			return zset, nil
		}
		return nil, config.ErrWrongType
	}

	newZSet := &EntryZSetBPTree{
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
			return count, config.ErrScoreIsNotFloat
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
	zset, err := s.getZSet(key)
	if err != nil {
		return 0, false, err
	}
	if zset == nil {
		return 0, false, nil
	}

	score, found := zset.dict[member]
	return score, found, nil
}

// ZCARD
func (s *Datastore) ZCard(key string) (int, error) {
	zset, err := s.getZSet(key)
	if err != nil {
		return 0, err
	}
	if zset == nil {
		return 0, nil
	}

	return len(zset.dict), nil
}

// ZRANK
func (s *Datastore) ZRank(key, member string) (int, bool, error) {
	zset, err := s.getZSet(key)
	if err != nil {
		return -1, false, err
	}
	if zset == nil {
		return 0, false, nil
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
	zset, err := s.getZSet(key)
	if err != nil {
		return nil, err
	}
	if zset == nil {
		return []string{}, nil
	}

	keys := zset.tree.rangeByRank(start, stop)
	if len(keys) == 0 {
		return []string{}, nil
	}

	res := make([]string, len(keys))
	for i, k := range keys {
		res[i] = k.member
	}
	return res, nil
}

// ZRANGE WITHSCORES
func (s *Datastore) ZRangeWithScore(key string, start, stop int) ([]string, error) {
	zset, err := s.getZSet(key)
	if err != nil {
		return nil, err
	}
	if zset == nil {
		return []string{}, nil
	}

	keys := zset.tree.rangeByRank(start, stop)
	if len(keys) == 0 {
		return []string{}, nil
	}

	res := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		res = append(res, k.member)
		res = append(res, strconv.FormatFloat(k.score, 'f', -1, 64))
	}
	return res, nil
}

// ZREM
func (s *Datastore) ZRem(key string, members []string) (int, error) {
	zset, err := s.getZSet(key)
	if err != nil {
		return 0, err
	}
	if zset == nil {
		return 0, nil
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
