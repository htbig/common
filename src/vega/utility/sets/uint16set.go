package sets

type Uint16set struct {
	data map[uint16]struct{}
}

func NewUint16Set() Uint16set {
	datamap := make(map[uint16]struct{})
	s := Uint16set{}
	s.data = datamap
	return s
}

func (set *Uint16set) Add(ints ...uint16) {
	for i := range ints {
		set.data[ints[i]] = struct{}{}
	}
}

func (set *Uint16set) Remove(i uint16) {
	delete(set.data, i)
}

func (set *Uint16set) Cardinality() int {
	return len(set.data)
}

func (set *Uint16set) DumpSlice() []uint16 {
	slice := []uint16{}
	for elem := range set.data {
		slice = append(slice, elem)
	}
	return slice
}
