package util

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewStringSet(t *testing.T) {
	set := NewStringSet(2)
	assert.Equal(t, 0, len(set.values), "New set is not empty")
}

func TestStringSet_Add(t *testing.T) {
	set := NewStringSet(1)
	set.Add("Hello")
	_, exists := set.values["Hello"]
	assert.True(t, exists, "Failed to add value to set")
}

func TestStringSet_Contains(t *testing.T) {
	set := NewStringSet(1)
	val := "Hello"
	assert.False(t, set.Contains(val), "Contains returns true for non-existent value!")
	set.Add(val)
	assert.True(t, set.Contains(val), "Contains returns false for existent value")
}

func TestStringSet_Remove(t *testing.T) {
	set := StringSet{
		values: map[string]void{"Hello": voidInst},
	}
	set.Remove("Hello")
	assert.False(t, set.Contains("Hello"), "Removing value from set is not working")
	assert.NotPanics(t, func() {
		set.Remove("World")
	}, "Failed to remove non-existent value")
}

func TestStringSet_Size(t *testing.T) {
	set := NewStringSet(10)
	require.Equal(t, 0, set.Size(), "The size of an empty is not zero")
	set.Add("Hello")
	set.Add("World")
	assert.Equal(t, 2, set.Size(), "The size of a set with added values is not correct")
	set = StringSet{
		values: map[string]void{"Hello": voidInst, "World": voidInst},
	}
	assert.Equal(t, 2, set.Size(), "The size of an initialized set is not correct")
}

func TestStringSet_Values(t *testing.T) {
	set := StringSet{values: map[string]void{"Hello": voidInst, "World": voidInst}}
	encounters := map[string]int{
		"Hello": 0,
		"World": 0,
	}
	for value := range set.Values() {
		_, exists := encounters[value]
		if !exists {
			t.Errorf("Set iteration is producing non-existent value \"%s\"", value)
			continue
		}

		encounters[value]++
	}

	for value, encs := range encounters {
		if encs == 0 {
			t.Errorf("Set iteration skipped value \"%s\"", value)
		} else if encs > 1 {
			t.Errorf("Set iteration produced value \"%s\" %d times", value, encs)
		}
	}
}
