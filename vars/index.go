package vars

// Index defines a way to iterate through joints of variables
type Index struct {
	current       int
	attrb, stride []int
	vs            VarList
}

// NewIndexFor creates a new index to iterate through indexVars relative to forVars
func NewIndexFor(indexVars, forVars VarList) (ix *Index) {
	ix = new(Index)
	ix.vs = forVars
	ix.attrb = make([]int, len(forVars))
	ix.stride = make([]int, len(forVars))
	j := 0
	s := 1
	for _, v := range indexVars {
		for ; j < len(forVars) && forVars[j].ID() < v.ID(); j++ {
		}
		if j < len(forVars) && forVars[j].ID() == v.ID() {
			ix.stride[j] = s
			j++
		}
		s *= v.NState()
	}
	return
}

// I returns current index
func (ix *Index) I() int {
	return ix.current
}

// Attribution returns the current attribution map
func (ix *Index) Attribution() map[int]int {
	m := make(map[int]int)
	for i, v := range ix.vs {
		m[v.ID()] = ix.attrb[i]
	}
	return m
}

// Ended if index reached end value
func (ix *Index) Ended() bool {
	return ix.current < 0
}

// Reset set index to beginning value
func (ix *Index) Reset() {
	ix.current = 0
	for i := range ix.attrb {
		ix.attrb[i] = 0
	}
}

// Next iterates to next index value
// returns true if it found a valid index
func (ix *Index) Next() bool {
	if ix.current >= 0 {
		for i := range ix.attrb {
			ix.current += ix.stride[i]
			ix.attrb[i]++
			if ix.attrb[i] < ix.vs[i].NState() {
				return true
			}
			ix.current -= ix.stride[i] * ix.vs[i].NState()
			ix.attrb[i] = 0
		}
		ix.current = -1
	}
	return false
}

// NextRight iterates to next index value, from right to left
// returns true if it found a valid index
func (ix *Index) NextRight() bool {
	if ix.current >= 0 {
		for i := len(ix.attrb) - 1; i >= 0; i-- {
			ix.current += ix.stride[i]
			ix.attrb[i]++
			if ix.attrb[i] < ix.vs[i].NState() {
				return true
			}
			ix.current -= ix.stride[i] * ix.vs[i].NState()
			ix.attrb[i] = 0
		}
		ix.current = -1
	}
	return false
}

// OrderedIndex defines a way to iterate through joints of variables in arbitrary order
type OrderedIndex struct {
	*Index
}

// NewOrderedIndex creates a new index to iterate in arbitrary ordering
func NewOrderedIndex(forVars []*Var) (ix *OrderedIndex) {
	ix = &OrderedIndex{new(Index)}
	ix.vs = forVars
	ix.attrb = make([]int, len(forVars))
	ix.stride = make([]int, len(forVars))
	s := 1
	for j, v := range forVars {
		ix.stride[j] = s
		s *= v.NState()
	}
	return
}
