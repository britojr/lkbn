package factor

import (
	"math"
	"reflect"
	"testing"

	"github.com/britojr/lkbn/vars"
	"gonum.org/v1/gonum/floats"
)

const tol = 1e-14

func TestNew(t *testing.T) {
	cases := []struct {
		vs vars.VarList
	}{
		{[]*vars.Var{}},
		{[]*vars.Var{vars.New(0, 2, "", false)}},
		{[]*vars.Var{vars.New(1, 4, "", false), vars.New(3, 2, "", false)}},
	}
	for _, tt := range cases {
		got := New(tt.vs...)
		if !tt.vs.Equal(got.vs) {
			t.Errorf("wrong vars %v != %v", tt.vs, got.vs)
		}
		if tt.vs.NStates() != len(got.values) {
			t.Errorf("wrong size of values %v != %v", tt.vs.NStates(), got.values)
		}
		if floats.Sum(got.values) != 1 {
			t.Errorf("factor not normalized, sum == %v", floats.Sum(got.values))
		}
	}
}

func TestNewEmpty(t *testing.T) {
	got := New()
	if got.vs == nil {
		t.Errorf("vars cannot be nil %v", got.vs)
	}
	if len(got.vs) != 0 {
		t.Errorf("vars should be empty, not %v", got.vs)
	}
	if len(got.values) != 1 {
		t.Errorf("values should have len=1 not %v", got.values)
	}
	if got.values[0] != 1 {
		t.Errorf("values should be normalized %v != 1", got.values[0])
	}
}

func TestSetValues(t *testing.T) {
	cases := []struct {
		f      *Factor
		values []float64
	}{
		{New(vars.NewList([]int{1, 3}, nil)...), []float64{1, 1, 0.5, 0.5}},
	}
	for _, tt := range cases {
		got := tt.f
		got.SetValues(tt.values)
		if !reflect.DeepEqual(tt.values, got.values) {
			t.Errorf("wrong values %v != %v", tt.values, got.values)
		}
	}
}

func TestPlus(t *testing.T) {
	cases := []struct {
		f, g   *Factor
		result []float64
	}{{
		New(vars.NewList([]int{1, 3}, nil)...).SetValues([]float64{5, 6, 7, 8}),
		New(vars.NewList([]int{1, 3}, nil)...).SetValues([]float64{1, 2, 3, 4}),
		[]float64{6, 8, 10, 12},
	}, {
		New(vars.NewList([]int{1, 3}, []int{2, 3})...).SetValues([]float64{5, 6, 7, 8, 9, 10}),
		New(vars.NewList([]int{1}, []int{2})...).SetValues([]float64{1, 2}),
		[]float64{6, 8, 8, 10, 10, 12},
	}}
	for _, tt := range cases {
		got := tt.f
		got.Plus(tt.g)
		if !reflect.DeepEqual(tt.result, got.values) {
			t.Errorf("wrong result %v != %v", tt.result, got.values)
		}
	}
}

func TestMinus(t *testing.T) {
	cases := []struct {
		f, g   *Factor
		result []float64
	}{{
		New(vars.NewList([]int{1, 3}, nil)...).SetValues([]float64{5, 6, 7, 8}),
		New(vars.NewList([]int{1, 3}, nil)...).SetValues([]float64{1, 2, 3, 4}),
		[]float64{4, 4, 4, 4},
	}, {
		New(vars.NewList([]int{1, 3}, []int{2, 3})...).SetValues([]float64{5, 6, 7, 8, 9, 10}),
		New(vars.NewList([]int{1}, []int{2})...).SetValues([]float64{1, 2}),
		[]float64{4, 4, 6, 6, 8, 8},
	}}
	for _, tt := range cases {
		got := tt.f
		got.Minus(tt.g)
		if !reflect.DeepEqual(tt.result, got.values) {
			t.Errorf("wrong result %v != %v", tt.result, got.values)
		}
	}
}

func TestTimes(t *testing.T) {
	cases := []struct {
		f, g   *Factor
		result []float64
	}{{
		New(vars.NewList([]int{1, 3}, nil)...).SetValues([]float64{5, 6, 7, 8}),
		New(vars.NewList([]int{1, 3}, nil)...).SetValues([]float64{1, 2, 3, 4}),
		[]float64{5, 12, 21, 32},
	}, {
		New(vars.NewList([]int{1, 3}, []int{2, 3})...).SetValues([]float64{5, 6, 7, 8, 9, 10}),
		New(vars.NewList([]int{1}, []int{2})...).SetValues([]float64{1, 2}),
		[]float64{5, 12, 7, 16, 9, 20},
	}, {
		New(vars.NewList([]int{3}, []int{3})...).SetValues([]float64{5, 6, 7}),
		New(vars.NewList([]int{1}, []int{2})...).SetValues([]float64{1, 2}),
		[]float64{5, 10, 6, 12, 7, 14},
	}, {
		New(vars.NewList([]int{0, 1}, []int{3})...).SetValues([]float64{0.5, 0.1, 0.3, 0.8, 0.0, 0.9}),
		New(vars.NewList([]int{1, 2}, nil)...).SetValues([]float64{0.5, 0.1, 0.7, 0.2}),
		[]float64{0.25, 0.05, 0.15, 0.08, 0.00, 0.09, 0.35, 0.07, 0.21, 0.16, 0.00, 0.18},
	}}
	for _, tt := range cases {
		got := tt.f
		got.Times(tt.g)
		if !floats.EqualApprox(tt.result, got.values, tol) {
			t.Errorf("wrong result %v != %v", tt.result, got.values)
		}
	}
}

func TestSumout(t *testing.T) {
	cases := []struct {
		f      *Factor
		vs     vars.VarList
		result []float64
	}{{
		New(vars.NewList([]int{0, 1, 2}, []int{3, 2, 2})...).SetValues(
			[]float64{0.25, 0.05, 0.15, 0.08, 0.00, 0.09, 0.35, 0.07, 0.21, 0.16, 0.00, 0.18},
		),
		vars.NewList([]int{1}, []int{2}),
		[]float64{0.33, 0.05, 0.24, 0.51, 0.07, 0.39},
	}, {
		New(vars.NewList([]int{0, 1, 2}, []int{3, 2, 2})...).SetValues(
			[]float64{0.25, 0.05, 0.15, 0.08, 0.00, 0.09, 0.35, 0.07, 0.21, 0.16, 0.00, 0.18},
		),
		vars.NewList([]int{2}, []int{2}),
		[]float64{0.60, 0.12, 0.36, 0.24, 0.0, 0.27},
	}, {
		New(vars.NewList([]int{0, 1, 2}, []int{3, 2, 2})...).SetValues(
			[]float64{0.25, 0.05, 0.15, 0.08, 0.00, 0.09, 0.35, 0.07, 0.21, 0.16, 0.00, 0.18},
		),
		vars.NewList([]int{0}, []int{3}),
		[]float64{0.45, 0.17, 0.63, 0.34},
	}, {
		New(vars.NewList([]int{0, 1, 2}, nil)...).SetValues(
			[]float64{0.25, 0.25, 0.25, 0.25, .5, .5, 1, 1},
		),
		vars.NewList([]int{0, 1, 2}, nil),
		[]float64{4},
	}, {
		New(vars.NewList([]int{0, 1, 2}, nil)...).SetValues(
			[]float64{0.25, 0.25, 0.25, 0.25, .5, .5, 1, 1},
		),
		vars.NewList([]int{0, 1}, nil),
		[]float64{1, 3},
	}, {
		New(vars.NewList([]int{0, 1, 2}, nil)...).SetValues(
			[]float64{0.25, 0.25, 0.25, 0.25, .5, .5, 1, 1},
		),
		vars.NewList([]int{}, nil),
		[]float64{0.25, 0.25, 0.25, 0.25, .5, .5, 1, 1},
	}}
	for _, tt := range cases {
		got := tt.f
		got.SumOut(tt.vs...)
		if !floats.EqualApprox(tt.result, got.values, tol) {
			t.Errorf("wrong result %v != %v", tt.result, got.values)
		}
	}
}

func TestMarginalize(t *testing.T) {
	cases := []struct {
		f      *Factor
		vs     vars.VarList
		result []float64
	}{{
		New(vars.NewList([]int{0, 1, 2}, []int{3, 2, 2})...).SetValues(
			[]float64{0.25, 0.05, 0.15, 0.08, 0.00, 0.09, 0.35, 0.07, 0.21, 0.16, 0.00, 0.18},
		),
		vars.NewList([]int{0, 2}, []int{3, 2}),
		[]float64{0.33, 0.05, 0.24, 0.51, 0.07, 0.39},
	}, {
		New(vars.NewList([]int{0, 1, 2}, []int{3, 2, 2})...).SetValues(
			[]float64{0.25, 0.05, 0.15, 0.08, 0.00, 0.09, 0.35, 0.07, 0.21, 0.16, 0.00, 0.18},
		),
		vars.NewList([]int{0, 2, 4}, []int{3, 2, 2}),
		[]float64{0.33, 0.05, 0.24, 0.51, 0.07, 0.39},
	}, {
		New(vars.NewList([]int{0, 1, 2}, []int{3, 2, 2})...).SetValues(
			[]float64{0.25, 0.05, 0.15, 0.08, 0.00, 0.09, 0.35, 0.07, 0.21, 0.16, 0.00, 0.18},
		),
		vars.NewList([]int{0, 1}, []int{3, 2}),
		[]float64{0.60, 0.12, 0.36, 0.24, 0.0, 0.27},
	}, {
		New(vars.NewList([]int{0, 1, 2}, []int{3, 2, 2})...).SetValues(
			[]float64{0.25, 0.05, 0.15, 0.08, 0.00, 0.09, 0.35, 0.07, 0.21, 0.16, 0.00, 0.18},
		),
		vars.NewList([]int{1, 2}, []int{2, 2}),
		[]float64{0.45, 0.17, 0.63, 0.34},
	}, {
		New(vars.NewList([]int{0, 1, 2}, nil)...).SetValues(
			[]float64{0.25, 0.25, 0.25, 0.25, .5, .5, 1, 1},
		),
		vars.NewList([]int{}, nil),
		[]float64{4},
	}, {
		New(vars.NewList([]int{0, 1, 2}, nil)...).SetValues(
			[]float64{0.25, 0.25, 0.25, 0.25, .5, .5, 1, 1},
		),
		vars.NewList([]int{2}, nil),
		[]float64{1, 3},
	}, {
		New(vars.NewList([]int{0, 1, 2}, nil)...).SetValues(
			[]float64{0.25, 0.25, 0.25, 0.25, .5, .5, 1, 1},
		),
		vars.NewList([]int{0, 1, 2}, nil),
		[]float64{0.25, 0.25, 0.25, 0.25, .5, .5, 1, 1},
	}}
	for _, tt := range cases {
		got := tt.f
		got.Marginalize(tt.vs...)
		if !floats.EqualApprox(tt.result, got.values, tol) {
			t.Errorf("wrong result %v != %v", tt.result, got.values)
		}
	}
}

func TestNormalize(t *testing.T) {
	cases := []struct {
		f      *Factor
		vs     vars.VarList
		result []float64
	}{{
		New(vars.NewList([]int{0, 1, 2}, []int{3, 2, 2})...).SetValues(
			[]float64{10, 3, 7, 2, 3, 5, 4, 4, 12, 6, 2, 2},
		),
		vars.NewList([]int{0}, []int{3}),
		[]float64{.5, .15, .35, .2, .3, .5, .2, .2, .6, .6, .2, .2},
	}, {
		New(vars.NewList([]int{0, 1}, []int{2, 2})...).SetValues(
			[]float64{10, 20, 30, 40},
		),
		nil,
		[]float64{.1, .2, .3, .4},
	}, {
		New(vars.NewList([]int{0, 1}, []int{2, 2})...).SetValues(
			[]float64{.1, .2, .3, .4},
		),
		vars.NewList([]int{0, 1}, []int{2, 2}),
		[]float64{.1, .2, .3, .4},
	}, {
		New(vars.NewList([]int{0, 1}, []int{2, 2})...).SetValues(
			[]float64{10, 20, 30, 40},
		),
		vars.NewList([]int{0, 1}, []int{2, 2}),
		[]float64{.1, .2, .3, .4},
	}, {
		New(vars.NewList([]int{1, 2, 4}, []int{2, 2, 2})...).SetValues(
			[]float64{0.136787, 0.155550, 0.111151, 0.157961, 0.105447, 0.122897, 0.092158, 0.118050},
		),
		vars.NewList([]int{4}, []int{2}),
		[]float64{
			0.136787 / 0.242234, 0.155550 / 0.278447, 0.111151 / 0.203309, 0.157961 / 0.276011,
			0.105447 / 0.242234, 0.122897 / 0.278447, 0.092158 / 0.203309, 0.118050 / 0.276011,
		},
	}, {
		New(vars.NewList([]int{0, 1, 4}, []int{2, 2, 2})...).SetValues(
			[]float64{0.125050, 0.125433, 0.117747, 0.135743, 0.051529, 0.143530, 0.128461, 0.172506},
		),
		vars.NewList([]int{4}, []int{2}),
		[]float64{
			0.125050 / 0.176579, 0.125433 / 0.268963, 0.117747 / 0.246208, 0.135743 / 0.308249,
			0.051529 / 0.176579, 0.143530 / 0.268963, 0.128461 / 0.246208, 0.172506 / 0.308249,
		},
	}}
	for _, tt := range cases {
		got := tt.f
		got.Normalize(tt.vs...)
		if !floats.EqualApprox(tt.result, got.values, tol) {
			t.Errorf("wrong result:\n%v\n!=\n%v\n", tt.result, got.values)
		}
	}
}

func TestRandom(t *testing.T) {
	xs := []*vars.Var{vars.New(1, 4, "", false), vars.New(3, 2, "", false)}
	got := New(xs...)

	got.RandomDistribute()
	if math.Abs(1.0-floats.Sum(got.values)) >= tol {
		t.Errorf("factor not normalized, sum == %v", floats.Sum(got.values))
	}
	c := append([]float64(nil), got.values...)
	got.RandomDistribute()
	if reflect.DeepEqual(c, got.values) {
		t.Errorf("sample same values %v != %v", c, got.values)
	}
	got.RandomDistribute(xs[0])
	if math.Abs(2.0-floats.Sum(got.values)) >= tol {
		t.Errorf("factor not normalized, sum == %v (%v)", floats.Sum(got.values), got.values)
	}
	got.RandomDistribute(xs[1])
	if math.Abs(4.0-floats.Sum(got.values)) >= tol {
		t.Errorf("factor not normalized, sum == %v (%v)", floats.Sum(got.values), got.values)
	}
}

func TestReduce(t *testing.T) {
	cases := []struct {
		f      *Factor
		e      map[int]int
		result []float64
	}{{
		New(vars.NewList([]int{0, 1}, []int{2, 2})...).SetValues(
			[]float64{10, 20, 30, 40},
		),
		map[int]int{},
		[]float64{10, 20, 30, 40},
	}, {
		New(vars.NewList([]int{0, 1}, []int{2, 2})...).SetValues(
			[]float64{10, 20, 30, 40},
		),
		map[int]int{1: 0},
		[]float64{10, 20, 0, 0},
	}, {
		New(vars.NewList([]int{0, 1}, []int{2, 2})...).SetValues(
			[]float64{10, 20, 30, 40},
		),
		map[int]int{0: 1, 1: 0, 2: 1},
		[]float64{0, 20, 0, 0},
	}, {
		New(vars.NewList([]int{0, 1, 2}, []int{3, 2, 2})...).SetValues(
			[]float64{10, 3, 7, 2, 3, 5, 4, 4, 12, 6, 2, 2},
		),
		map[int]int{0: 1, 2: 1, 4: 1},
		[]float64{0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 2, 0},
	}}
	for _, tt := range cases {
		got := tt.f
		got.Reduce(tt.e)
		if !reflect.DeepEqual(tt.result, got.values) {
			t.Errorf("wrong result %v != %v", tt.result, got.values)
		}
	}
}

func TestGet(t *testing.T) {
	cases := []struct {
		f      *Factor
		e      map[int]int
		result float64
	}{{
		New(vars.NewList([]int{0, 1, 2}, []int{3, 2, 2})...).SetValues(
			[]float64{10, 3, 7, 2, 3, 5, 4, 4, 12, 6, 2, 2},
		),
		map[int]int{0: 1, 1: 0, 2: 1}, 4,
	}}
	for _, tt := range cases {
		got := tt.f.Get(tt.e)
		if !floats.EqualApprox([]float64{tt.result}, []float64{got}, tol) {
			t.Errorf("wrong result %v != %v", tt.result, got)
		}
	}
}
