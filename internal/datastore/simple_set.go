package datastore

type entrySimpleSet struct {
	mapVal map[string]struct{}
	// expireAt *time.Time
}

type SimpleSet struct {
	m map[string]entrySimpleSet
}

func NewSimpleSet() *SimpleSet {
	return &SimpleSet{
		m: make(map[string]entrySimpleSet),
	}
}

func (s *SimpleSet) SADD(key string, members []string) int {
	countAdded := 0
	if _, ok := s.m[key]; !ok {
		s.m[key] = entrySimpleSet{mapVal: make(map[string]struct{})}
	}
	for _, m := range members {
		if _, ok := s.m[key].mapVal[m]; !ok {
			s.m[key].mapVal[m] = struct{}{}
			countAdded++
		}
	}

	return countAdded
}

func (s *SimpleSet) SMembers(key string) []string {
	if _, ok := s.m[key]; !ok {
		return []string{}
	}

	m := []string{}
	for k := range s.m[key].mapVal {
		m = append(m, k)
	}
	return m
}

func (s *SimpleSet) SIsMember(key string, member string) int {
	if _, ok := s.m[key]; !ok {
		return 0
	}

	if _, ok := s.m[key].mapVal[member]; !ok {
		return 0
	}

	return 1
}

func (s *SimpleSet) SMIsMember(key string, members []string) []int {
	n := len(members)

	if _, ok := s.m[key]; !ok {
		return make([]int, n)
	}

	rs := make([]int, n)
	for i, m := range members {
		if _, ok := s.m[key].mapVal[m]; ok {
			rs[i] = 1
		}
	}

	return rs
}
