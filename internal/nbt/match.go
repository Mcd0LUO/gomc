package nbt

// Matches reports whether the value `have` satisfies the filter `want`.
//
// This implements the 1.12.2 "subset" matching semantics used by selector
// `nbt=` filters and by `/testfor`:
//
//   - A compound filter matches when every key in the filter exists in the
//     subject with a matching value. Extra keys in the subject are allowed.
//   - A list filter matches when the subject is a list of the same length and
//     every element matches pairwise.
//   - Numeric filters match numerically, ignoring the exact width, so `1b`
//     matches a stored `1`.
//   - A zero/absent filter is not a wildcard: `{}` matches only an empty
//     compound, and a missing key never matches.
func Matches(subject, filter Value) bool {
	if subject == nil || filter == nil {
		return subject == nil && filter == nil
	}
	switch f := filter.(type) {
	case *Compound:
		s, ok := subject.(*Compound)
		if !ok {
			return false
		}
		for _, k := range f.Keys() {
			want := f.Get(k)
			got := s.Get(k)
			if got == nil {
				return false
			}
			if !Matches(got, want) {
				return false
			}
		}
		return true

	case *List:
		s, ok := subject.(*List)
		if !ok {
			return false
		}
		if f.Len() != s.Len() {
			return false
		}
		for i := 0; i < f.Len(); i++ {
			if !Matches(s.At(i), f.At(i)) {
				return false
			}
		}
		return true

	case ByteArray:
		s, ok := subject.(ByteArray)
		if !ok {
			return false
		}
		if len(s) != len(f) {
			return false
		}
		for i := range f {
			if s[i] != f[i] {
				return false
			}
		}
		return true

	case IntArray:
		s, ok := subject.(IntArray)
		if !ok {
			return false
		}
		if len(s) != len(f) {
			return false
		}
		for i := range f {
			if s[i] != f[i] {
				return false
			}
		}
		return true

	case String:
		s, ok := subject.(String)
		if !ok {
			return false
		}
		return string(s) == string(f)
	}

	// Numeric comparison across widths.
	if sf, sok := numeric(subject); sok {
		if ff, fok := numeric(filter); fok {
			return sf == ff
		}
		return false
	}
	return false
}

// numeric extracts a comparable numeric value.
func numeric(v Value) (float64, bool) {
	switch t := v.(type) {
	case Byte:
		return float64(t), true
	case Short:
		return float64(t), true
	case Int:
		return float64(t), true
	case Long:
		return float64(t), true
	case Float:
		return float64(t), true
	case Double:
		return float64(t), true
	}
	return 0, false
}

// Contains reports whether the compound holds every entry of the filter.
// It is the compound-level shorthand for Matches.
func Contains(subject *Compound, filter *Compound) bool {
	if subject == nil || filter == nil {
		return false
	}
	return Matches(subject, filter)
}
