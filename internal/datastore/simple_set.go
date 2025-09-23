package datastore

import "errors"

type EntrySimpleSet struct {
	mapVal map[string]struct{}
}

func (s *Datastore) SADD(key string, members []string) (int, error) {
	if _, ok := s.m[key]; !ok {
		s.m[key] = Entry{
			val: &EntrySimpleSet{
				mapVal: make(map[string]struct{}),
			},
		}
	}

	switch entryVal := s.m[key].val.(type) {
	case *EntrySimpleSet:
		countAdded := 0
		for _, m := range members {
			if _, exists := entryVal.mapVal[m]; !exists {
				entryVal.mapVal[m] = struct{}{}
				countAdded++
			}
		}
		return countAdded, nil
	default:
		return 0, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
}

func (s *Datastore) SMembers(key string) ([]string, error) {
	entry, ok := s.m[key]
	if !ok {
		return []string{}, nil
	}

	switch entryVal := entry.val.(type) {
	case *EntrySimpleSet:
		members := make([]string, 0, len(entryVal.mapVal))
		for k := range entryVal.mapVal {
			members = append(members, k)
		}
		return members, nil
	default:
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
}

func (s *Datastore) SIsMember(key string, member string) (int, error) {
	entry, ok := s.m[key]
	if !ok {
		return 0, nil
	}

	setEntry, ok := entry.val.(*EntrySimpleSet)
	if !ok {
		return 0, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	if _, exists := setEntry.mapVal[member]; exists {
		return 1, nil
	}
	return 0, nil
}

func (s *Datastore) SMIsMember(key string, members []string) ([]int, error) {
	n := len(members)
	results := make([]int, n)

	entry, ok := s.m[key]
	if !ok {
		return results, nil
	}

	setEntry, ok := entry.val.(*EntrySimpleSet)
	if !ok {
		return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	for i, m := range members {
		if _, exists := setEntry.mapVal[m]; exists {
			results[i] = 1
		}
	}

	return results, nil
}
