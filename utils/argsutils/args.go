package argsutils

import "fmt"

// StringSlice is a string slice.
type StringSlice struct {
	values map[string]bool
}

// NewStringSlice creates a string slice values.
func NewStringSlice(values ...string) *StringSlice {
	s := &StringSlice{
		values: make(map[string]bool),
	}
	for _, v := range values {
		s.values[v] = true
	}
	return s
}

func (v *StringSlice) String() string {
	return fmt.Sprintf("%v", v.List())
}

// List returns value list.
func (v *StringSlice) List() []string {
	res := []string{}
	for v, ok := range v.values {
		if ok {
			res = append(res, v)
		}
	}
	return res
}

// Set set the value.
func (v *StringSlice) Set(value string) error {
	if value == "xxx" {
		v.values = make(map[string]bool)
	} else {
		v.values[value] = true
	}
	return nil
}
