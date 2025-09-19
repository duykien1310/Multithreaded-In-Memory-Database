package datastore

import (
	"fmt"
	"sync"
)

// Assume bptree, bptKey, etc. are defined elsewhere in package datastore.

type entryZSetBPTree struct {
	dict map[string]float64 // member -> score
	tree *bptree
}

type ZSetBPTree struct {
	mu sync.RWMutex
	m  map[string]*entryZSetBPTree
}

func NewZSetBPTree() *ZSetBPTree {
	return &ZSetBPTree{
		m: make(map[string]*entryZSetBPTree),
	}
}

func (z *ZSetBPTree) ensureEntry(key string) *entryZSetBPTree {
	ent := z.m[key]
	if ent == nil {
		ent = &entryZSetBPTree{
			dict: make(map[string]float64),
			tree: newBPTree(), // assumes newBPTree() exists
		}
		z.m[key] = ent
	}
	return ent
}

// ZAdd: add or update member with score
func (z *ZSetBPTree) ZADD(key string, score float64, member string) int {
	z.mu.Lock()
	defer z.mu.Unlock()

	ent := z.ensureEntry(key)

	// If member existed, remove old score from tree first.
	if oldScore, ok := ent.dict[member]; ok {
		if oldScore == score {
			// nothing to do
			return 0
		}
		// remove old (score, member) from the tree
		ent.tree.delete(bptKey{score: oldScore, member: member})
	}

	// Insert/update
	ent.dict[member] = score
	return ent.tree.insert(bptKey{score: score, member: member})
}

// ZScore
func (z *ZSetBPTree) ZScore(key string, member string) (float64, bool) {
	z.mu.RLock()
	defer z.mu.RUnlock()

	ent := z.m[key]
	if ent == nil {
		return 0, false
	}
	score, ok := ent.dict[member]
	return score, ok
}

// ZCard
func (z *ZSetBPTree) ZCard(key string) int {
	z.mu.RLock()
	defer z.mu.RUnlock()

	ent := z.m[key]
	if ent == nil {
		return 0
	}
	return len(ent.dict)
}

// ZRank (0-based)
func (z *ZSetBPTree) ZRank(key string, member string) (int, bool) {
	z.mu.RLock()
	defer z.mu.RUnlock()

	ent := z.m[key]
	if ent == nil {
		return -1, false
	}
	score, ok := ent.dict[member]
	if !ok {
		return -1, false
	}
	r := ent.tree.rankOf(bptKey{score: score, member: member})
	if r < 0 {
		return -1, false
	}
	return r, true
}

// ZRange by rank (inclusive indices)
func (z *ZSetBPTree) ZRange(key string, start, stop int) []string {
	z.mu.RLock()
	defer z.mu.RUnlock()

	ent := z.m[key]
	if ent == nil {
		return nil
	}

	bptKeys := ent.tree.rangeByRank(start, stop)
	rs := []string{}
	for _, key := range bptKeys {
		rs = append(rs, key.member)
	}

	return rs
}

func (z *ZSetBPTree) ZRangeWithScore(key string, start, stop int) []string {
	z.mu.RLock()
	defer z.mu.RUnlock()

	ent := z.m[key]
	if ent == nil {
		return nil
	}

	bptKeys := ent.tree.rangeByRank(start, stop)
	rs := []string{}
	for _, key := range bptKeys {
		rs = append(rs, key.member)
		rs = append(rs, fmt.Sprintf("%.2f", key.score))
	}

	return rs
}

// ZRem removes a member and returns true if removed
func (z *ZSetBPTree) ZRem(key string, member string) bool {
	z.mu.Lock()
	defer z.mu.Unlock()

	ent := z.m[key]
	if ent == nil {
		return false
	}

	score, ok := ent.dict[member]
	if !ok {
		return false
	}

	deleted := ent.tree.delete(bptKey{score: score, member: member})
	if deleted {
		delete(ent.dict, member)
	}

	// If entry is now empty, remove it from top-level map to avoid growing empty entries.
	if len(ent.dict) == 0 {
		delete(z.m, key)
	}

	return deleted
}
