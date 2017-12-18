package vars

import "testing"

func TestCopy(t *testing.T) {
	cases := []struct {
		vl VarList
	}{
		{[]*Var{}},
		{[]*Var(nil)},
		{[]*Var{New(0, 2, "", false), New(3, 2, "", false), New(5, 4, "", false)}},
	}
	for _, tt := range cases {
		w := tt.vl.Copy()
		if !tt.vl.Equal(w) {
			t.Errorf("not equal %v != %v", tt.vl, w)
		}
		// test if its safe for change
		if len(tt.vl) > 0 {
			w[0] = New(9, 2, "", false)
		} else {
			w = append(w, New(9, 2, "", false))
			tt.vl = append(tt.vl, New(8, 2, "", false))
		}
		if tt.vl.Equal(w) {
			t.Errorf("both are pointing to the same slice %v == %v", &tt.vl, &w)
		}
	}
}

func TestNStates(t *testing.T) {
	cases := []struct {
		vl VarList
		ns int
	}{
		{[]*Var{}, 1},
		{[]*Var(nil), 1},
		{[]*Var{New(0, 2, "", false), New(3, 2, "", false), New(5, 4, "", false)}, 16},
	}
	for _, tt := range cases {
		got := tt.vl.NStates()
		if tt.ns != got {
			t.Errorf("wrong number of states %v != %v", tt.ns, got)
		}
	}
}

func TestEqual(t *testing.T) {
	cases := []struct {
		va, vb VarList
		eq     bool
	}{
		{[]*Var{}, []*Var{}, true},
		{[]*Var(nil), []*Var(nil), true},
		{[]*Var{New(0, 2, "", false), New(3, 2, "", false), New(5, 4, "", false)},
			[]*Var{New(0, 2, "", false), New(3, 2, "", false), New(5, 4, "", false)}, true},
		{[]*Var{New(5, 4, "", false)}, []*Var{New(5, 3, "", false)}, true},
		{[]*Var{New(0, 2, "", false)}, []*Var{New(0, 2, "", false), New(1, 2, "", false)}, false},
		{[]*Var{New(0, 2, "", false), New(2, 2, "", false)}, []*Var{New(0, 2, "", false), New(1, 2, "", false)}, false},
	}
	for _, tt := range cases {
		got := tt.va.Equal(tt.vb)
		if tt.eq != got {
			t.Errorf("wrong compare result %v != %v", tt.eq, got)
		}
		if tt.vb.Equal(tt.va) != got {
			t.Errorf("equal function should be simetric %v != %v", tt.vb.Equal(tt.va), got)
		}
	}
}

func TestNewList(t *testing.T) {
	cases := []struct {
		vs, ns []int
		res    VarList
	}{
		{[]int{1, 5, 0}, []int{3, 2, 4}, []*Var{New(0, 4, "", false), New(1, 3, "", false), New(5, 2, "", false)}},
		{[]int{1, 5, 0}, []int(nil), []*Var{New(0, 2, "", false), New(1, 2, "", false), New(5, 2, "", false)}},
	}
	for _, tt := range cases {
		got := NewList(tt.vs, tt.ns)
		if !tt.res.Equal(got) {
			t.Errorf("wrong new list %v != %v", tt.res, got)
		}
	}
}

func TestAdd(t *testing.T) {
	cases := []struct {
		vs, ws, res VarList
	}{
		{NewList([]int{0, 1, 2}, nil), nil, NewList([]int{0, 1, 2}, nil)},
		{NewList([]int{1, 2}, nil), NewList([]int{4}, nil), NewList([]int{1, 2, 4}, nil)},
		{NewList([]int{1, 2}, nil), NewList([]int{4, 0}, nil), NewList([]int{0, 1, 2, 4}, nil)},
		{NewList(nil, nil), NewList([]int{4, 0, 5, 12, 2, 7, 1}, nil), NewList([]int{0, 1, 2, 4, 5, 7, 12}, nil)},
		{NewList([]int{1, 2}, nil), NewList([]int{4, 2, 1, 0}, nil), NewList([]int{0, 1, 2, 4}, nil)},
	}
	for _, tt := range cases {
		got := tt.vs
		for _, v := range tt.ws {
			got.Add(v)
		}
		if !tt.res.Equal(got) {
			t.Errorf("wrong add %v != %v", tt.res, got)
		}
	}
}

func TestRemove(t *testing.T) {
	cases := []struct {
		vs, ws, res VarList
	}{
		{NewList([]int{0, 1, 2}, nil), nil, NewList([]int{0, 1, 2}, nil)},
		{NewList([]int{1, 2, 4}, nil), NewList([]int{4}, nil), NewList([]int{1, 2}, nil)},
		{NewList([]int{1, 2, 4}, nil), NewList([]int{5}, nil), NewList([]int{1, 2, 4}, nil)},
		{NewList([]int{0, 1, 2, 4}, nil), NewList([]int{4, 0}, nil), NewList([]int{1, 2}, nil)},
		{NewList([]int{0, 1, 2, 4, 5, 7, 12}, nil), NewList([]int{4, 0, 5, 12, 2, 7, 1}, nil), NewList(nil, nil)},
		{NewList([]int{0, 1, 2, 4}, nil), NewList([]int{4, 2, 1, 0}, nil), nil},
	}
	for _, tt := range cases {
		got := tt.vs
		for _, v := range tt.ws {
			got.Remove(v.ID())
		}
		if !tt.res.Equal(got) {
			t.Errorf("wrong remove %v != %v", tt.res, got)
		}
	}
}

func TestDiff(t *testing.T) {
	cases := []struct {
		a, b, want []int
	}{
		{[]int(nil), []int(nil), []int(nil)},
		{[]int{2, 3}, []int{}, []int{2, 3}},
		{[]int{}, []int{3, 4}, []int(nil)},
		{[]int{1, 2, 3, 4, 5, 6}, []int{2, 4, 6, 8}, []int{1, 3, 5}},
		{[]int{3, 4, 5, 6}, []int{1, 2, 4, 6, 8}, []int{3, 5}},
	}
	for _, tt := range cases {
		a := NewList(tt.a, nil)
		b := NewList(tt.b, nil)
		got := a.Diff(b)
		for i, v := range tt.want {
			if v != got[i].ID() {
				t.Errorf("(%v) - (%v) = (%v) !=(%v)", a, b, tt.want, got)
			}
		}
	}
}

func TestUnion(t *testing.T) {
	cases := []struct {
		a, b, want []int
	}{
		{[]int(nil), []int(nil), []int(nil)},
		{[]int{1}, []int{1}, []int{1}},
		{[]int{3}, []int{1}, []int{1, 3}},
		{[]int{1}, []int{3}, []int{1, 3}},
		{[]int{2, 3}, []int{}, []int{2, 3}},
		{[]int{}, []int{3, 4}, []int{3, 4}},
		{[]int{6, 4, 2, 8}, []int{8, 9, 3, 1, 2}, []int{1, 2, 3, 4, 6, 8, 9}},
	}
	for _, tt := range cases {
		a := NewList(tt.a, nil)
		b := NewList(tt.b, nil)
		got := a.Union(b)
		for i, v := range tt.want {
			if v != got[i].ID() {
				t.Errorf("(%v) union (%v) = (%v) !=(%v)", a, b, tt.want, got)
			}
		}
	}
}
