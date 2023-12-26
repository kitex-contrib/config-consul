package utils

import "sync"

// ThreadSafeSet wrapper of Set.
type ThreadSafeSet struct {
	sync.RWMutex
	s Set
}

// DiffAndEmplace returns the keys that are not in other and emplace the old set.
func (ts *ThreadSafeSet) DiffAndEmplace(other Set) []string {
	ts.Lock()
	defer ts.Unlock()
	out := ts.s.Diff(other)
	ts.s = other
	return out
}

// Set map template.
type Set map[string]bool

// Diff returns the keys that are not in other
func (s Set) Diff(other Set) []string {
	out := make([]string, 0, len(s))
	for key := range s {
		if _, ok := other[key]; !ok {
			out = append(out, key)
		}
	}
	return out
}
