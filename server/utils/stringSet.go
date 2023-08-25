package utils

type StringSet map[string]struct{}

// NewStringSet returns a new, empty StringSet
func NewStringSet() StringSet {
	return make(StringSet)
}

// Add inserts a string into the set
func (s StringSet) Add(str string) {
	s[str] = struct{}{}
}

// Contains checks if the set contains the given string
func (s StringSet) Contains(str string) bool {
	_, exists := s[str]
	return exists
}

// Difference returns a new set with all the items that are in this set but not in other
func (s StringSet) Difference(other StringSet) StringSet {
	result := NewStringSet()
	for item := range s {
		if !other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}
