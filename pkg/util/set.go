package util

// Set based on https://yourbasic.org/golang/implement-set/

type void struct{}

var voidInst void

type StringSet struct {
	values map[string]void
}

func NewStringSet(size int) StringSet {
	return StringSet{
		values: make(map[string]void, size),
	}
}

func (set *StringSet) Add(value string) {
	set.values[value] = voidInst
}

func (set *StringSet) Contains(value string) bool {
	_, exists := set.values[value]
	return exists
}

func (set *StringSet) Remove(value string) {
	delete(set.values, value)
}

func (set *StringSet) Size() int {
	return len(set.values)
}

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
