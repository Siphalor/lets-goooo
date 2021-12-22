// Part of the Let's Goooo project
// Copyright 2021; matriculation numbers: 1103207, 3106445, 4485500
// Let's goooo get this over together

package util

// Set based on https://yourbasic.org/golang/implement-set/

type void struct{}

var voidInst void

// StringSet is a collection of unique strings.
type StringSet struct {
	values map[string]void
}

// NewStringSet creates a new StringSet with the given start capacity.
func NewStringSet(size int) StringSet {
	return StringSet{
		values: make(map[string]void, size),
	}
}

// Add adds a new element to the set, if not already present.
func (set *StringSet) Add(value string) {
	set.values[value] = voidInst
}

// Contains checks if the given element is already in the set.
func (set *StringSet) Contains(value string) bool {
	_, exists := set.values[value]
	return exists
}

// Remove removes the given element from the set.
func (set *StringSet) Remove(value string) {
	delete(set.values, value)
}

// Size delivers the number of elements of the set.
func (set *StringSet) Size() int {
	return len(set.values)
}

// Values provides a channel to iterate over all elements in the set.
func (set *StringSet) Values() <-chan string {
	out := make(chan string)
	go func() {
		for val := range set.values {
			out <- val
		}
		close(out)
	}()
	return out
}
