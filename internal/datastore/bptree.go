package datastore

import (
	"sort"
)

// A simplified in-memory B+ tree for (score, member) keys.
// Keys are (score float64, member string) and order is by score, then member lexicographically.
// Internal nodes store separator keys and child pointers; leaves store actual entries and are linked.

const bptOrder = 3

type bptKey struct {
	score  float64
	member string
}

func (k bptKey) less(other bptKey) bool {
	if k.score != other.score {
		return k.score < other.score
	}
	return k.member < other.member
}

func (k bptKey) equal(other bptKey) bool {
	return k.score == other.score && k.member == other.member
}

type bptNode interface {
	isLeaf() bool
}

type bptLeaf struct {
	keys   []bptKey
	next   *bptLeaf
	prev   *bptLeaf
	parent *bptInternal
}

type bptInternal struct {
	sep    []bptKey  // separator keys (len = children-1)
	child  []bptNode // child pointers
	counts []int     // subtree counts per child (same len as child)
	parent *bptInternal
}

func (l *bptLeaf) isLeaf() bool     { return true }
func (n *bptInternal) isLeaf() bool { return false }

// new tree: root is leaf
type bptree struct {
	root bptNode
	size int
}

func newBPTree() *bptree {
	return &bptree{
		root: &bptLeaf{keys: make([]bptKey, 0)},
		size: 0,
	}
}

// findLeaf returns leaf node and index where key should be inserted or exists.
func (t *bptree) findLeaf(key bptKey) (*bptLeaf, int) {
	n := t.root
	for {
		if n.isLeaf() {
			l := n.(*bptLeaf)
			// binary search by key
			idx := sort.Search(len(l.keys), func(i int) bool {
				return !l.keys[i].less(key)
			})
			return l, idx
		}
		in := n.(*bptInternal)
		// choose child by separators
		idx := sort.Search(len(in.sep), func(i int) bool {
			return !in.sep[i].less(key)
		})
		// if idx == len(sep) -> choose last child
		if idx == len(in.sep) {
			// check last separator
			idx = len(in.child) - 1
		} else {
			// need to determine which child; sep[i] >= key => child i
			// standard B+ internal semantics: child index = idx
		}
		n = in.child[idx]
	}
}

// insert key into leaf at index idx (idx may be len(keys) meaning append)
func (t *bptree) insert(key bptKey) int {
	leaf, idx := t.findLeaf(key)
	// If exact equal, replace (duplicate member+score shouldn't happen since member unique; but we will allow update through higher-level)
	if idx < len(leaf.keys) && leaf.keys[idx].equal(key) {
		// already present; nothing
		return 0
	}
	// insert into slice
	leaf.keys = append(leaf.keys, bptKey{})  // grow
	copy(leaf.keys[idx+1:], leaf.keys[idx:]) // shift right
	leaf.keys[idx] = key
	t.size++
	// update ancestor counts
	t.incrementCounts(leaf, 1)
	// split if overflow
	if len(leaf.keys) >= bptOrder {
		t.splitLeaf(leaf)
	}

	return 1
}

func (t *bptree) incrementCounts(n bptNode, delta int) {
	switch n := n.(type) {
	case *bptLeaf:
		var node bptNode = n
		p := n.parent
		for p != nil {
			// find index of child
			var idx int
			for i, c := range p.child {
				if c == node {
					idx = i
					break
				}
			}
			p.counts[idx] += delta
			node = p
			p = p.parent
		}
	case *bptInternal:
		// shouldn't usually call with internal, but we can ascend
		var node bptNode = n
		p := n.parent
		for p != nil {
			var idx int
			for i, c := range p.child {
				if c == node {
					idx = i
					break
				}
			}
			p.counts[idx] += delta
			node = p
			p = p.parent
		}
	}
}

func (t *bptree) splitLeaf(l *bptLeaf) {
	mid := len(l.keys) / 2
	newLeaf := &bptLeaf{
		keys:   append([]bptKey(nil), l.keys[mid:]...),
		next:   l.next,
		prev:   l,
		parent: l.parent,
	}
	l.keys = l.keys[:mid]
	if l.next != nil {
		l.next.prev = newLeaf
	}
	l.next = newLeaf

	// attach into parent
	if l.parent == nil {
		// new root
		newRoot := &bptInternal{
			sep:    []bptKey{newLeaf.keys[0]},
			child:  []bptNode{l, newLeaf},
			counts: []int{len(l.keys), len(newLeaf.keys)},
		}
		l.parent = newRoot
		newLeaf.parent = newRoot
		t.root = newRoot
		return
	}

	p := l.parent
	insertIdx := 0
	for insertIdx < len(p.child) && p.child[insertIdx] != l {
		insertIdx++
	}

	// ✅ fix: allow inserting at end (rightmost)
	if insertIdx > len(p.sep) {
		insertIdx = len(p.sep)
	}

	// insert child
	p.child = append(p.child, nil)
	copy(p.child[insertIdx+1:], p.child[insertIdx:])
	p.child[insertIdx+1] = newLeaf

	p.sep = append(p.sep, bptKey{})
	copy(p.sep[insertIdx+1:], p.sep[insertIdx:])
	p.sep[insertIdx] = newLeaf.keys[0]

	p.counts = append(p.counts, 0)
	copy(p.counts[insertIdx+1:], p.counts[insertIdx:])
	p.counts[insertIdx] = len(l.keys)
	p.counts[insertIdx+1] = len(newLeaf.keys)

	newLeaf.parent = p
	t.fixUpAfterInsert(p)
}

func (t *bptree) fixUpAfterInsert(p *bptInternal) {
	if len(p.child) > bptOrder {
		// Split internal node.
		mid := len(p.child) / 2

		// The key to promote is sep[mid-1]
		promoted := p.sep[mid-1]

		// Left internal keeps children [0:mid)
		// Right internal keeps children [mid:]
		leftChild := append([]bptNode(nil), p.child[:mid]...)
		rightChild := append([]bptNode(nil), p.child[mid:]...)

		leftSep := append([]bptKey(nil), p.sep[:mid-1]...) // before promoted
		rightSep := append([]bptKey(nil), p.sep[mid:]...)  // after promoted

		leftCounts := append([]int(nil), p.counts[:mid]...)
		rightCounts := append([]int(nil), p.counts[mid:]...)

		// Modify left node (p) in place
		p.child = leftChild
		p.sep = leftSep
		p.counts = leftCounts

		right := &bptInternal{
			sep:    rightSep,
			child:  rightChild,
			counts: rightCounts,
			parent: p.parent,
		}
		for _, c := range right.child {
			if c.isLeaf() {
				c.(*bptLeaf).parent = right
			} else {
				c.(*bptInternal).parent = right
			}
		}

		if p.parent == nil {
			// new root
			newRoot := &bptInternal{
				sep:    []bptKey{promoted},
				child:  []bptNode{p, right},
				counts: []int{sumCounts(p), sumCounts(right)},
			}
			p.parent = newRoot
			right.parent = newRoot
			t.root = newRoot
			return
		}

		// insert into parent
		par := p.parent
		idx := 0
		for idx < len(par.child) && par.child[idx] != p {
			idx++
		}
		par.child = append(par.child, nil)
		copy(par.child[idx+1:], par.child[idx:])
		par.child[idx+1] = right

		par.sep = append(par.sep, bptKey{})
		copy(par.sep[idx+1:], par.sep[idx:])
		par.sep[idx] = promoted

		par.counts = append(par.counts, 0)
		copy(par.counts[idx+1:], par.counts[idx:])
		par.counts[idx] = sumCounts(p)
		par.counts[idx+1] = sumCounts(right)

		right.parent = par
		t.fixUpAfterInsert(par)
	} else {
		// recompute counts upwards
		cur := p
		for cur != nil {
			for i := range cur.counts {
				cur.counts[i] = nodeCount(cur.child[i])
			}
			cur = cur.parent
		}
	}
}

func sumCounts(n bptNode) int {
	// total items in node subtree
	if n.isLeaf() {
		return len(n.(*bptLeaf).keys)
	}
	sum := 0
	for _, c := range n.(*bptInternal).child {
		sum += nodeCount(c)
	}
	return sum
}

func nodeCount(n bptNode) int {
	if n.isLeaf() {
		return len(n.(*bptLeaf).keys)
	}
	total := 0
	for _, c := range n.(*bptInternal).child {
		total += nodeCount(c)
	}
	return total
}

// delete a (score, member) key from tree; returns true if removed
func (t *bptree) delete(key bptKey) bool {
	leaf, idx := t.findLeaf(key)
	if idx >= len(leaf.keys) || !leaf.keys[idx].equal(key) {
		return false
	}
	// remove
	leaf.keys = append(leaf.keys[:idx], leaf.keys[idx+1:]...)
	t.size--
	t.incrementCounts(leaf, -1)
	// merging/borrowing omitted for simplicity (can be added). We keep nodes under-full tolerant.
	return true
}

// rank returns 0-based rank of key if present, or -1
func (t *bptree) rankOf(key bptKey) int {
	// walk down, accumulate counts of left siblings
	var rank int
	n := t.root
	for {
		if n.isLeaf() {
			l := n.(*bptLeaf)
			idx := sort.Search(len(l.keys), func(i int) bool { return !l.keys[i].less(key) })
			if idx < len(l.keys) && l.keys[idx].equal(key) {
				return rank + idx
			}
			return -1
		}
		in := n.(*bptInternal)
		// choose child and accumulate counts of previous children
		idx := sort.Search(len(in.sep), func(i int) bool { return !in.sep[i].less(key) })
		// idx is child index
		// add counts of children before idx
		for i := 0; i < idx; i++ {
			rank += in.counts[i]
		}
		n = in.child[idx]
	}
}

// range by rank: returns slice of entries from start to stop inclusive (supports negative indices)
func (t *bptree) rangeByRank(start, stop int) []bptKey {
	if t == nil || t.size == 0 || t.root == nil {
		return nil
	}

	n := t.size
	// normalize negatives
	if start < 0 {
		start += n
	}
	if stop < 0 {
		stop += n
	}
	if start < 0 {
		start = 0
	}
	if stop >= n {
		stop = n - 1
	}
	if start > stop {
		return nil
	}

	need := stop - start + 1
	idx := start
	node := t.root

	// descend to leaf containing start
	for !node.isLeaf() {
		in := node.(*bptInternal)
		cum := 0
		childIdx := 0
		for i := 0; i < len(in.child); i++ {
			if cum+in.counts[i] > idx {
				childIdx = i
				break
			}
			cum += in.counts[i]
		}
		idx -= cum
		node = in.child[childIdx]
	}

	leaf := node.(*bptLeaf)
	result := make([]bptKey, 0, need)

	// iterate forward across leaves collecting members
	for curr := leaf; curr != nil && len(result) < need; curr = curr.next {
		for ; idx < len(curr.keys) && len(result) < need; idx++ {
			// copy key struct (cheap)
			k := curr.keys[idx]
			result = append(result, k)
		}
		idx = 0
	}
	return result
}
