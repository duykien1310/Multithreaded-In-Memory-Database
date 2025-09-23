package datastore

import "backend/internal/config"

type EntrySimpleSet struct {
	mapVal map[string]struct{}
}

func (s *Datastore) getSimpleSet(key string) (*EntrySimpleSet, error) {
	e, ok := s.getEntry(key)
	if !ok {
		return nil, nil
	}
	set, ok := e.val.(*EntrySimpleSet)
	if !ok {
		return nil, config.ErrWrongType
	}
	return set, nil
}

func (s *Datastore) SADD(key string, members []string) (int, error) {
	e, ok := s.getEntry(key)
	if !ok {
		set := &EntrySimpleSet{mapVal: make(map[string]struct{})}
		s.m[key] = Entry{val: set}
		e = s.m[key]
	}

	set, ok := e.val.(*EntrySimpleSet)
	if !ok {
		return 0, config.ErrWrongType
	}

	countAdded := 0
	for _, m := range members {
		if _, exists := set.mapVal[m]; !exists {
			set.mapVal[m] = struct{}{}
			countAdded++
		}
	}
	return countAdded, nil
}

func (s *Datastore) SMembers(key string) ([]string, error) {
	set, err := s.getSimpleSet(key)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return []string{}, nil
	}

	members := make([]string, 0, len(set.mapVal))
	for k := range set.mapVal {
		members = append(members, k)
	}
	return members, nil
}

// SIsMember checks if a member exists in the set.
func (s *Datastore) SIsMember(key, member string) (int, error) {
	set, err := s.getSimpleSet(key)
	if err != nil {
		return 0, err
	}
	if set == nil {
		return 0, nil
	}

	if _, exists := set.mapVal[member]; exists {
		return 1, nil
	}
	return 0, nil
}

// SMIsMember checks multiple members at once.
func (s *Datastore) SMIsMember(key string, members []string) ([]int, error) {
	results := make([]int, len(members))

	set, err := s.getSimpleSet(key)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return results, nil
	}

	for i, m := range members {
		if _, exists := set.mapVal[m]; exists {
			results[i] = 1
		}
	}
	return results, nil
}
