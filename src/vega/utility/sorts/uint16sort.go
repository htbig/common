package sorts

import "sort"

type Uint16Slice []uint16

func (p Uint16Slice) Len() int           { return len(p) }
func (p Uint16Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint16Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func Uint16s(a []uint16) {

	sort.Sort(Uint16Slice(a))
}
